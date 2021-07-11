package main

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Generate Salt
// func generateSalt() (string, error) {
// 	saltBytes := make([]byte, 32)
// 	_, err := io.ReadFull(rand.Reader, saltBytes)
// 	salt := hex.EncodeToString(saltBytes)
// 	return salt, err
// }

func generateUUID() string {
	return uuid.New().String()
}

// Generate a bcrypt Hash (see: https://en.wikipedia.org/wiki/Bcrypt)
func generateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func createJWT(user User, groups []Group) (JWT, error) {
	expirationTime := time.Now().Add(config.JWTExpiration * time.Minute)
	claims := JWTClaims{
		user.ID,
		user.Username,
		user.UUID,
		groups,
		jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "dev",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTKey))
	jwt := JWT{Value: tokenString, Claims: claims}
	return jwt, err
}

func createRefresh(userID int) (RefreshToken, error) {
	expirationTime := time.Now().Add(config.JWTExpiration * time.Minute)
	claims := RefreshClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "dev",
			Id:        generateUUID(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTKey))
	refreshToken := RefreshToken{Value: tokenString, JTI: claims.Id}
	return refreshToken, err
}

func checkJWTClaims(r *http.Request) (*JWTClaims, error) {
	jwtCookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, err
	}
	tokenString := jwtCookie.Value

	claims := &JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}

func checkRefreshClaims(r *http.Request) (*RefreshClaims, error) {
	refreshCookie, err := r.Cookie("refresh")
	if err != nil {
		return nil, err
	}
	tokenString := refreshCookie.Value
	claims := &RefreshClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}
