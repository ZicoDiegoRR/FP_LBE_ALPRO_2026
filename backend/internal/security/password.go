package security

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah password plaintext menjadi hash bcrypt dengan salt acak
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword memvalidasi apakah password plaintext cocok dengan hash di database
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength memverifikasi kekuatan password sesuai standar NIST
// Minimal 8 karakter, mengandung huruf besar, huruf kecil, angka, dan karakter khusus
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password minimal harus 8 karakter")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password harus mengandung minimal satu huruf besar (A-Z)")
	}
	if !hasLower {
		return errors.New("password harus mengandung minimal satu huruf kecil (a-z)")
	}
	if !hasNumber {
		return errors.New("password harus mengandung minimal satu angka (0-9)")
	}
	if !hasSpecial {
		return errors.New("password harus mengandung minimal satu karakter khusus/simbol (!@#$%^&*)")
	}

	return nil
}