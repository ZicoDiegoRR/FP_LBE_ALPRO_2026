package crypto

import (
	"bytes"
	"testing"
)

func TestRSAFullLifecycle(t *testing.T) {
	// 1. Uji Pembuatan Pasangan Kunci RSA (2048-bit)
	pubKeyPEM, privKeyPEM, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("Gagal membuat pasangan kunci RSA: %v", err)
	}
	if pubKeyPEM == "" || privKeyPEM == "" {
		t.Fatal("Public Key atau Private Key tidak boleh kosong")
	}

	// 2. Uji Enkripsi & Dekripsi Kunci Simetris (RSA-OAEP)
	// Skenario: Alice mengenkripsi kunci AES-256 (32 byte) dengan Public Key Bob
	aesKeyRahasia := []byte("kunci-simetris-aes-harus-32-byte")

	ciphertextKunci, err := EncryptRSAOAEP(aesKeyRahasia, pubKeyPEM)
	if err != nil {
		t.Fatalf("Enkripsi RSA-OAEP gagal: %v", err)
	}

	// Bob mendekripsi ciphertext menggunakan Private Key miliknya
	hasilDekripsiKunci, err := DecryptRSAOAEP(ciphertextKunci, privKeyPEM)
	if err != nil {
		t.Fatalf("Dekripsi RSA-OAEP gagal: %v", err)
	}

	if !bytes.Equal(hasilDekripsiKunci, aesKeyRahasia) {
		t.Fatalf("Kunci hasil dekripsi tidak cocok! Didapat: %s", string(hasilDekripsiKunci))
	}
	t.Log("Uji Pertukaran Kunci Simetris via RSA-OAEP: SUKSES!")

	// 3. Uji Digital Signature (Tanda Tangan Digital pada Dokumen PDF)
	dokumenPDF := []byte("%PDF-1.4 Dokumen Kontrak Peminjaman Alat Lab Alpro 2026...")

	// Alice menandatangani dokumen dengan Private Key miliknya
	signature, err := SignDocument(dokumenPDF, privKeyPEM)
	if err != nil {
		t.Fatalf("Gagal menandatangani dokumen: %v", err)
	}

	// Siapa pun memverifikasi tanda tangan dokumen menggunakan Public Key Alice
	err = VerifyDocumentSignature(dokumenPDF, signature, pubKeyPEM)
	if err != nil {
		t.Fatalf("Verifikasi tanda tangan digital yang sah GAGAL: %v", err)
	}
	t.Log("Uji Verifikasi Tanda Tangan Digital Sah: SUKSES!")

	// 4. Uji Keamanan: Dokumen telah dimanipulasi peretas (Tampering Attack)
	dokumenDimanipulasi := []byte("%PDF-1.4 Dokumen Kontrak Peminjaman Alat Lab Alpro 2026... (DIUBAH PERETAS)")

	err = VerifyDocumentSignature(dokumenDimanipulasi, signature, pubKeyPEM)
	if err == nil {
		t.Fatal("Keamanan jebol: Tanda tangan digital seharusnya DITOLAK jika isi file diubah!")
	}
	t.Log("Uji Deteksi Manipulasi File (Integrity Check): SUKSES, Perubahan Terdeteksi!")
}