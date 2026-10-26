package filecrypt

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"
)

var master = []byte("0123456789abcdef0123456789abcdef")
var allCiphers = []Cipher{AES, DES, RC4}

func randBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

func encWith(t *testing.T, plain []byte, c Cipher, chunk int) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := encrypt(&out, bytes.NewReader(plain), c, master, chunk); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestRoundTrip(t *testing.T) {
	const chunk = 1024
	sizes := []int{0, 1, 7, 8, 9, chunk - 1, chunk, chunk + 1, 2 * chunk, 3*chunk + 5}
	for _, c := range allCiphers {
		for _, n := range sizes {
			t.Run(fmt.Sprintf("%s/%d", c, n), func(t *testing.T) {
				plain := randBytes(t, n)
				ct := encWith(t, plain, c, chunk)
				got, err := DecryptBytes(ct, master)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, plain) {
					t.Fatal("hasil dekripsi tidak sama dengan plaintext")
				}
			})
		}
	}
}

func TestRoundTripDefaultChunk(t *testing.T) {
	plain := randBytes(t, 200*1024+3)
	for _, c := range allCiphers {
		ct, err := EncryptBytes(plain, c, master)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecryptBytes(ct, master)
		if err != nil || !bytes.Equal(got, plain) {
			t.Fatalf("%s: round-trip gagal: %v", c, err)
		}
	}
}

func TestWrongKey(t *testing.T) {
	other := []byte("ffffffffffffffffffffffffffffffff")
	for _, c := range allCiphers {
		ct := encWith(t, []byte("data rahasia"), c, 1024)
		if _, err := DecryptBytes(ct, other); !errors.Is(err, ErrAuth) {
			t.Fatalf("%s: seharusnya ErrAuth, dapat %v", c, err)
		}
	}
}

func TestTamperEveryByte(t *testing.T) {
	plain := randBytes(t, 3000)
	for _, c := range allCiphers {
		ct := encWith(t, plain, c, 1024)
		for i := range ct {
			mod := append([]byte(nil), ct...)
			mod[i] ^= 0x01
			if _, err := DecryptBytes(mod, master); err == nil {
				t.Fatalf("%s: perubahan di byte %d tidak terdeteksi", c, i)
			}
		}
	}
}

func TestTruncation(t *testing.T) {
	plain := randBytes(t, 3000)
	for _, c := range allCiphers {
		ct := encWith(t, plain, c, 1024)
		for cut := 0; cut < len(ct); cut++ {
			if _, err := DecryptBytes(ct[:cut], master); err == nil {
				t.Fatalf("%s: pemotongan di %d byte tidak terdeteksi", c, cut)
			}
		}
	}
}

func TestTrailingData(t *testing.T) {
	for _, c := range allCiphers {
		ct := encWith(t, []byte("abc"), c, 1024)
		if _, err := DecryptBytes(append(ct, 0), master); err == nil {
			t.Fatalf("%s: sisa data tidak terdeteksi", c)
		}
	}
}

func TestCiphertextUniqueAndHidesPlaintext(t *testing.T) {
	plain := bytes.Repeat([]byte("RAHASIA-KTM-"), 100)
	for _, c := range allCiphers {
		a, _ := EncryptBytes(plain, c, master)
		b, _ := EncryptBytes(plain, c, master)
		if bytes.Equal(a, b) {
			t.Fatalf("%s: dua enkripsi plaintext sama menghasilkan ciphertext identik", c)
		}
		if bytes.Contains(a, []byte("RAHASIA")) {
			t.Fatalf("%s: plaintext terlihat di ciphertext", c)
		}
	}
}