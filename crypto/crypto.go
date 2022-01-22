package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// Generate Salt
// func generateSalt() (string, error) {
// 	saltBytes := make([]byte, 32)
// 	_, err := io.ReadFull(rand.Reader, saltBytes)
// 	salt := hex.EncodeToString(saltBytes)
// 	return salt, err
// }

// Generate a bcrypt Hash (see: https://en.wikipedia.org/wiki/Bcrypt)
func GenerateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
