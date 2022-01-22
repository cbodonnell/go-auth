package main

import (
	"net/http"
	"time"

	"github.com/cheebz/go-auth/models"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// JWTClaims struct
type JWTClaims struct {
	UserID   int            `json:"user_id"`
	Username string         `json:"username"`
	UUID     string         `json:"uuid"`
	Groups   []models.Group `json:"groups"`
	jwt.StandardClaims
}

// JWT struct {
type JWT struct {
	Value  string
	Claims JWTClaims
}

// RefreshClaims struct
type RefreshClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// RefreshToken struct {
type RefreshToken struct {
	Value string
	JTI   string
}

func generateUUID() string {
	return uuid.New().String()
}

func createJWT(user models.User, groups []models.Group) (JWT, error) {
	expirationTime := time.Now().Add(time.Duration(conf.JWTMaxAge) * time.Minute)
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
	tokenString, err := token.SignedString([]byte(conf.JWTKey))
	jwt := JWT{Value: tokenString, Claims: claims}
	return jwt, err
}

func createRefresh(userID int) (RefreshToken, error) {
	expirationTime := time.Now().Add(time.Duration(conf.RefreshMaxAge) * time.Minute)
	claims := RefreshClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "dev",
			Id:        generateUUID(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(conf.JWTKey))
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
		return []byte(conf.JWTKey), nil
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
		return []byte(conf.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}
