package security

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah password plaintext menjadi hash bcrypt dengan salt acak
func HashPassword(password string) (string, error) {
	// Cost 10 adalah standar industri: seimbang antara keamanan tinggi dan efisiensi CPU
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword memvalidasi apakah password plaintext cocok dengan hash di database
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}