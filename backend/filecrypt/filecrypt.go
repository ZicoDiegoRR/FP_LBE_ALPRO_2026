// Package filecrypt mengenkripsi aliran data dengan AES-256-GCM,
// DES-CBC + HMAC-SHA256, atau RC4 + HMAC-SHA256.
//
// Format berkas:
//
//	header (26 byte): "PMOE" | versi | id cipher | ukuran potongan (uint32) | salt (16)
//	lalu rangkaian record: flag final (1) | panjang body (uint32) | body
//
// Kunci enkripsi dan kunci MAC diturunkan per berkas dengan HKDF-SHA256.
package filecrypt

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

type Cipher string

const (
	AES Cipher = "aes"
	DES Cipher = "des"
	RC4 Cipher = "rc4"
)

const (
	DefaultChunkSize      = 64 * 1024
	maxChunkSize          = 16 << 20
	saltSize              = 16
	macSize               = sha256.Size
	headerSize            = 4 + 1 + 1 + 4 + saltSize
	version          byte = 1
)

var magic = [4]byte{'P', 'M', 'O', 'E'}

var (
	ErrFormat    = errors.New("filecrypt: format berkas tidak valid")
	ErrAuth      = errors.New("filecrypt: autentikasi gagal (kunci salah atau data diubah)")
	ErrTruncated = errors.New("filecrypt: data terpotong")
)

func cipherID(c Cipher) (byte, error) {
	switch c {
	case AES:
		return 1, nil
	case DES:
		return 2, nil
	case RC4:
		return 3, nil
	}
	return 0, errors.New("filecrypt: cipher tidak dikenal")
}

func cipherFromID(id byte) (Cipher, error) {
	switch id {
	case 1:
		return AES, nil
	case 2:
		return DES, nil
	case 3:
		return RC4, nil
	}
	return "", ErrFormat
}

// codec memproses satu potongan. index dan final selalu ikut diautentikasi.
type codec interface {
	seal(index uint64, final bool, plain []byte) ([]byte, error)
	open(index uint64, final bool, body []byte) ([]byte, error)
}

func macOf(key, header []byte, index uint64, final bool, data []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(header)
	var meta [9]byte
	binary.BigEndian.PutUint64(meta[:8], index)
	if final {
		meta[8] = 1
	}
	m.Write(meta[:])
	m.Write(data)
	return m.Sum(nil)
}

func newCodec(c Cipher, master, header []byte) (codec, error) {
	if len(master) < 16 {
		return nil, errors.New("filecrypt: master key minimal 16 byte")
	}
	encLen := map[Cipher]int{AES: 32, DES: 8, RC4: 16}[c]
	if encLen == 0 {
		return nil, errors.New("filecrypt: cipher tidak dikenal")
	}
	salt := header[10:headerSize]
	okm, err := hkdf.Key(sha256.New, master, salt, "pmo-filecrypt-v1/"+string(c), encLen+macSize)
	if err != nil {
		return nil, err
	}
	encKey, macKey := okm[:encLen], okm[encLen:]
	hdr := append([]byte(nil), header...)

	switch c {
	case AES:
		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		return &aesCodec{gcm: gcm, header: hdr}, nil
	case DES:
		block, err := des.NewCipher(encKey)
		if err != nil {
			return nil, err
		}
		return &desCodec{block: block, macKey: macKey, header: hdr}, nil
	default:
		stream, err := rc4.NewCipher(encKey)
		if err != nil {
			return nil, err
		}
		return &rc4Codec{stream: stream, macKey: macKey, header: hdr}, nil
	}
}

// ---------- AES-256-GCM ----------

type aesCodec struct {
	gcm    cipher.AEAD
	header []byte
}

func (a *aesCodec) nonce(index uint64, final bool) []byte {
	n := make([]byte, a.gcm.NonceSize())
	binary.BigEndian.PutUint64(n, index)
	if final {
		n[8] = 1
	}
	return n
}

func (a *aesCodec) seal(index uint64, final bool, plain []byte) ([]byte, error) {
	return a.gcm.Seal(nil, a.nonce(index, final), plain, a.header), nil
}

func (a *aesCodec) open(index uint64, final bool, body []byte) ([]byte, error) {
	pt, err := a.gcm.Open(nil, a.nonce(index, final), body, a.header)
	if err != nil {
		return nil, ErrAuth
	}
	return pt, nil
}

// ---------- DES-CBC + HMAC ----------

type desCodec struct {
	block  cipher.Block
	macKey []byte
	header []byte
}

func (d *desCodec) seal(index uint64, final bool, plain []byte) ([]byte, error) {
	bs := des.BlockSize
	pad := bs - len(plain)%bs
	padded := append(append([]byte(nil), plain...), bytes.Repeat([]byte{byte(pad)}, pad)...)

	body := make([]byte, bs+len(padded))
	iv := body[:bs]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(d.block, iv).CryptBlocks(body[bs:], padded)
	return append(body, macOf(d.macKey, d.header, index, final, body)...), nil
}

func (d *desCodec) open(index uint64, final bool, body []byte) ([]byte, error) {
	bs := des.BlockSize
	if len(body) < 2*bs+macSize || (len(body)-macSize)%bs != 0 {
		return nil, ErrFormat
	}
	data, tag := body[:len(body)-macSize], body[len(body)-macSize:]
	if !hmac.Equal(tag, macOf(d.macKey, d.header, index, final, data)) {
		return nil, ErrAuth
	}
	iv, ct := data[:bs], data[bs:]
	pt := make([]byte, len(ct))
	cipher.NewCBCDecrypter(d.block, iv).CryptBlocks(pt, ct)
	pad := int(pt[len(pt)-1])
	if pad < 1 || pad > bs || pad > len(pt) {
		return nil, ErrFormat
	}
	return pt[:len(pt)-pad], nil
}

// ---------- RC4 + HMAC ----------
// Satu aliran RC4 dipakai berurutan untuk seluruh berkas. Kunci unik per berkas.

type rc4Codec struct {
	stream *rc4.Cipher
	macKey []byte
	header []byte
}

func (r *rc4Codec) seal(index uint64, final bool, plain []byte) ([]byte, error) {
	ct := make([]byte, len(plain))
	r.stream.XORKeyStream(ct, plain)
	return append(ct, macOf(r.macKey, r.header, index, final, ct)...), nil
}

func (r *rc4Codec) open(index uint64, final bool, body []byte) ([]byte, error) {
	if len(body) < macSize {
		return nil, ErrFormat
	}
	ct, tag := body[:len(body)-macSize], body[len(body)-macSize:]
	if !hmac.Equal(tag, macOf(r.macKey, r.header, index, final, ct)) {
		return nil, ErrAuth
	}
	pt := make([]byte, len(ct))
	r.stream.XORKeyStream(pt, ct)
	return pt, nil
}

// ---------- API ----------

func writeRecord(w io.Writer, final bool, body []byte) error {
	var h [5]byte
	if final {
		h[0] = 1
	}
	binary.BigEndian.PutUint32(h[1:], uint32(len(body)))
	if _, err := w.Write(h[:]); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}

// Encrypt membaca src sampai habis dan menulis hasil enkripsi ke dst.
func Encrypt(dst io.Writer, src io.Reader, c Cipher, master []byte) error {
	return encrypt(dst, src, c, master, DefaultChunkSize)
}

func encrypt(dst io.Writer, src io.Reader, c Cipher, master []byte, chunkSize int) error {
	id, err := cipherID(c)
	if err != nil {
		return err
	}
	if chunkSize <= 0 || chunkSize > maxChunkSize {
		return ErrFormat
	}
	header := make([]byte, headerSize)
	copy(header, magic[:])
	header[4] = version
	header[5] = id
	binary.BigEndian.PutUint32(header[6:10], uint32(chunkSize))
	if _, err := io.ReadFull(rand.Reader, header[10:]); err != nil {
		return err
	}
	cd, err := newCodec(c, master, header)
	if err != nil {
		return err
	}
	if _, err := dst.Write(header); err != nil {
		return err
	}

	br := bufio.NewReader(src)
	buf := make([]byte, chunkSize)
	for index := uint64(0); ; index++ {
		n, err := io.ReadFull(br, buf)
		final := false
		switch {
		case err == nil:
			// potongan penuh: cek apakah masih ada data setelahnya
			if _, perr := br.Peek(1); perr == io.EOF {
				final = true
			} else if perr != nil {
				return perr
			}
		case err == io.EOF || err == io.ErrUnexpectedEOF:
			final = true
		default:
			return err
		}
		body, err := cd.seal(index, final, buf[:n])
		if err != nil {
			return err
		}
		if err := writeRecord(dst, final, body); err != nil {
			return err
		}
		if final {
			return nil
		}
	}
}

// Decrypt membaca data terenkripsi dari src dan menulis plaintext ke dst.
// Setiap potongan diverifikasi sebelum ditulis, tetapi jika fungsi ini
// mengembalikan error, hasil yang sudah tertulis di dst HARUS dibuang.
func Decrypt(dst io.Writer, src io.Reader, master []byte) error {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(src, header); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return ErrTruncated
		}
		return err
	}
	if !bytes.Equal(header[:4], magic[:]) || header[4] != version {
		return ErrFormat
	}
	c, err := cipherFromID(header[5])
	if err != nil {
		return err
	}
	chunkSize := binary.BigEndian.Uint32(header[6:10])
	if chunkSize == 0 || chunkSize > maxChunkSize {
		return ErrFormat
	}
	cd, err := newCodec(c, master, header)
	if err != nil {
		return err
	}
	maxBody := chunkSize + 64

	var rh [5]byte
	for index := uint64(0); ; index++ {
		if _, err := io.ReadFull(src, rh[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return ErrTruncated
			}
			return err
		}
		if rh[0] > 1 {
			return ErrFormat
		}
		final := rh[0] == 1
		size := binary.BigEndian.Uint32(rh[1:])
		if size > maxBody {
			return ErrFormat
		}
		body := make([]byte, size)
		if _, err := io.ReadFull(src, body); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return ErrTruncated
			}
			return err
		}
		plain, err := cd.open(index, final, body)
		if err != nil {
			return err
		}
		if _, err := dst.Write(plain); err != nil {
			return err
		}
		if final {
			var one [1]byte
			if n, _ := io.ReadFull(src, one[:]); n > 0 {
				return ErrFormat // ada sisa data setelah potongan terakhir
			}
			return nil
		}
	}
}

// EncryptBytes dan DecryptBytes adalah pembungkus untuk data kecil.
func EncryptBytes(plain []byte, c Cipher, master []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := Encrypt(&out, bytes.NewReader(plain), c, master); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func DecryptBytes(data, master []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := Decrypt(&out, bytes.NewReader(data), master); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}