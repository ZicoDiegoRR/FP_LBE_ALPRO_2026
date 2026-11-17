package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims mendefinisikan payload data yang disimpan di dalam token JWT
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken membuat token JWT baru bertanda tangan digital (HMAC-SHA256)
func GenerateToken(userID uint, username, role, secretKey string) (string, error) {
	// Token berlaku selama 24 jam
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pmo-security-system",
		},
	}

	// Buat token dengan algoritma penandatanganan HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Tandatangani token dengan secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken memverifikasi keaslian signature token dan masa berlakunya
func ValidateToken(tokenString, secretKey string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validasi bahwa metode signing token adalah HMAC (mencegah None Algorithm Attack)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing token tidak valid")
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("token tidak valid atau sudah kadaluarsa")
	}

	return claims, nil
}