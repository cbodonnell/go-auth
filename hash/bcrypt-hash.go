package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type BCryptHash struct {
	Cost int
}

func NewBCryptHash(cost int) Hash {
	return &BCryptHash{
		Cost: cost,
	}
}

// Generate a bcrypt Hash (see: https://en.wikipedia.org/wiki/Bcrypt)
func (b *BCryptHash) Generate(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.Cost)
	return string(bytes), err
}

func (b *BCryptHash) Check(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// Generate Salt
// func generateSalt() (string, error) {
// 	saltBytes := make([]byte, 32)
// 	_, err := io.ReadFull(rand.Reader, saltBytes)
// 	salt := hex.EncodeToString(saltBytes)
// 	return salt, err
// }
