package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
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
func generateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func createJWT(username string, groups []Group) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := JWTClaims{
		username,
		groups,
		jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "dev",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTKey))
	return tokenString, err
}

func getClaims(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTKey), nil
	})
	return claims, err
}
