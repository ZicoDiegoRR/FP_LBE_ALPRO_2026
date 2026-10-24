package utils

import "crypto/rand"

const idCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const idLength = 20

// GenerateID produces a random 20-character alphanumeric ID matching the
// schema's char(20) primary key columns (e.g. "User".id).
func GenerateID() string {
	b := make([]byte, idLength)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = idCharset[i%len(idCharset)]
		}
		return string(b)
	}
	for i, v := range b {
		b[i] = idCharset[int(v)%len(idCharset)]
	}
	return string(b)
}
