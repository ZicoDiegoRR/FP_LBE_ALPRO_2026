package handlers

import (
	"bufio"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"pmo-backend/db"
	"pmo-backend/filecrypt"
	"pmo-backend/models"
	"pmo-backend/utils"
)

const (
	vaultDir       = "./vault"
	maxUploadBytes = 200 << 20 // 200 MiB
)

var allowedCategories = map[string]bool{
	"identity_scan": true,
	"document":      true,
	"video":         true,
}

var allowedCiphers = map[string]filecrypt.Cipher{
	"aes": filecrypt.AES,
	"des": filecrypt.DES,
	"rc4": filecrypt.RC4,
}

// fileMasterKey membaca master key dari environment. Sengaja tanpa kunci
// cadangan: kalau .env kosong, unggah/unduh ditolak daripada memakai kunci lemah.
func fileMasterKey() ([]byte, error) {
	k := os.Getenv("ENCRYPTION_SECRET_KEY")
	if len(k) < 16 {
		return nil, errors.New("ENCRYPTION_SECRET_KEY kurang dari 16 karakter")
	}
	return []byte(k), nil
}

// cleanFileName membuang path dan karakter kontrol dari nama berkas asli.
// Nama ini hanya untuk tampilan dan header unduhan, tidak pernah dipakai sebagai path di disk.
func cleanFileName(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == '"' {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" {
		return "file"
	}
	if r := []rune(name); len(r) > 120 {
		name = string(r[:120])
	}
	return name
}

// POST /files (requires auth)
// Form multipart: file (wajib), category (identity_scan|document|video), cipher (aes|des|rc4, default aes)
func UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+(1<<20))

	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{false, "No file provided"})
		return
	}
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, APIResponse{false, "File too large (max 200 MB)"})
		return
	}

	category := c.PostForm("category")
	if !allowedCategories[category] {
		c.JSON(http.StatusBadRequest, APIResponse{false, "category must be identity_scan, document, or video"})
		return
	}

	cipherName := strings.ToLower(c.DefaultPostForm("cipher", "aes"))
	cph, ok := allowedCiphers[cipherName]
	if !ok {
		c.JSON(http.StatusBadRequest, APIResponse{false, "cipher must be aes, des, or rc4"})
		return
	}

	key, err := fileMasterKey()
	if err != nil {
		log.Printf("upload: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Server encryption key is not configured"})
		return
	}

	src, err := header.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not read uploaded file"})
		return
	}
	defer src.Close()

	// Deteksi jenis berkas dari isinya, jangan percaya Content-Type dari klien.
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(src, sniff)
	mimeType := http.DetectContentType(sniff[:n])
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not read uploaded file"})
		return
	}

	if err := os.MkdirAll(vaultDir, 0o700); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not prepare storage"})
		return
	}

	id := utils.GenerateID()
	finalPath := filepath.Join(vaultDir, id+".enc")

	tmp, err := os.CreateTemp(vaultDir, "upload-*.part")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not prepare storage"})
		return
	}
	tmpName := tmp.Name()
	fail := func(msg string) {
		tmp.Close()
		os.Remove(tmpName)
		c.JSON(http.StatusInternalServerError, APIResponse{false, msg})
	}

	bw := bufio.NewWriter(tmp)
	if err := filecrypt.Encrypt(bw, src, cph, key); err != nil {
		log.Printf("upload: encrypt failed (file %s, cipher %s): %v", id, cipherName, err)
		fail("Failed to encrypt file")
		return
	}
	if err := bw.Flush(); err != nil {
		fail("Failed to write encrypted file")
		return
	}
	if err := tmp.Sync(); err != nil {
		fail("Failed to write encrypted file")
		return
	}
	info, err := tmp.Stat()
	if err != nil {
		fail("Failed to write encrypted file")
		return
	}
	cipherSize := info.Size()
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to write encrypted file"})
		return
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		os.Remove(tmpName)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Failed to store encrypted file"})
		return
	}

	rec := models.StoredFile{
		ID:         id,
		OwnerID:    userID,
		Category:   category,
		OrigName:   cleanFileName(header.Filename),
		MimeType:   mimeType,
		Cipher:     cipherName,
		PlainSize:  header.Size,
		CipherSize: cipherSize,
		CreatedAt:  time.Now(),
	}
	if err := db.DB.Create(&rec).Error; err != nil {
		os.Remove(finalPath)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not save file record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "File encrypted and stored",
		"file":    fileJSON(rec),
	})
}

func fileJSON(f models.StoredFile) gin.H {
	return gin.H{
		"id":         f.ID,
		"name":       f.OrigName,
		"category":   f.Category,
		"mimeType":   f.MimeType,
		"cipher":     f.Cipher,
		"plainSize":  f.PlainSize,
		"cipherSize": f.CipherSize,
		"createdAt":  f.CreatedAt.Format("Jan 2, 2006 15:04"),
	}
}

// GET /files (requires auth) - hanya berkas milik pengguna yang login.
func ListFiles(c *gin.Context) {
	userID := c.GetString("user_id")

	var rows []models.StoredFile
	if err := db.DB.Where("owner_id = ?", userID).Order("created_at DESC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not load files"})
		return
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, fileJSON(r))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "files": out})
}

// GET /files/:id (requires auth)
// Berkas didekripsi ke berkas sementara dulu. Isinya baru dikirim ke klien
// setelah seluruh potongan lolos verifikasi, jadi data yang rusak/dipotong
// tidak pernah sampai ke klien.
func DownloadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	// Pemilik ikut di query: berkas milik orang lain dan berkas yang tidak ada
	// sama-sama 404, jadi ID orang lain tidak bisa ditebak keberadaannya.
	var rec models.StoredFile
	if err := db.DB.Where("id = ? AND owner_id = ?", c.Param("id"), userID).First(&rec).Error; err != nil {
		c.JSON(http.StatusNotFound, APIResponse{false, "File not found"})
		return
	}

	key, err := fileMasterKey()
	if err != nil {
		log.Printf("download: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Server encryption key is not configured"})
		return
	}

	enc, err := os.Open(filepath.Join(vaultDir, rec.ID+".enc"))
	if err != nil {
		log.Printf("download: encrypted file missing (file %s): %v", rec.ID, err)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Stored file is missing"})
		return
	}
	defer enc.Close()

	plain, err := os.CreateTemp("", "pmo-dec-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not prepare decryption"})
		return
	}
	defer func() {
		plain.Close()
		os.Remove(plain.Name())
	}()

	bw := bufio.NewWriter(plain)
	if err := filecrypt.Decrypt(bw, bufio.NewReader(enc), key); err != nil {
		log.Printf("download: verification failed (file %s, cipher %s): %v", rec.ID, rec.Cipher, err)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "File failed integrity verification"})
		return
	}
	if err := bw.Flush(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not decrypt file"})
		return
	}

	// Cek silang dengan ukuran yang dicatat saat unggah.
	if info, err := plain.Stat(); err != nil || info.Size() != rec.PlainSize {
		log.Printf("download: size mismatch (file %s)", rec.ID)
		c.JSON(http.StatusInternalServerError, APIResponse{false, "File failed integrity verification"})
		return
	}
	if _, err := plain.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{false, "Could not decrypt file"})
		return
	}

	// Hanya gambar dan mp4 yang boleh tampil di browser; sisanya diunduh.
	disposition := "attachment"
	contentType := "application/octet-stream"
	switch rec.MimeType {
	case "image/jpeg", "image/png", "video/mp4":
		disposition = "inline"
		contentType = rec.MimeType
	}
	cd := mime.FormatMediaType(disposition, map[string]string{"filename": rec.OrigName})
	if cd == "" {
		cd = disposition
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", cd)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")

	http.ServeContent(c.Writer, c.Request, rec.OrigName, rec.CreatedAt, plain)
}
