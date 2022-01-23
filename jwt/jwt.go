package jwt

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

type JWTHelper struct {
	JWTKey        string
	JWTMaxAge     int
	RefreshMaxAge int
}

func NewJWTHelper(jwtKey string, jwtMaxAge int, refreshMaxAge int) *JWTHelper {
	return &JWTHelper{
		JWTKey:        jwtKey,
		JWTMaxAge:     jwtMaxAge,
		RefreshMaxAge: refreshMaxAge,
	}
}

func (j *JWTHelper) CreateJWT(user models.User, groups []models.Group) (JWT, error) {
	expirationTime := time.Now().Add(time.Duration(j.JWTMaxAge) * time.Minute)
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
	tokenString, err := token.SignedString([]byte(j.JWTKey))
	jwt := JWT{Value: tokenString, Claims: claims}
	return jwt, err
}

func (j *JWTHelper) CreateRefresh(userID int) (RefreshToken, error) {
	expirationTime := time.Now().Add(time.Duration(j.RefreshMaxAge) * time.Minute)
	claims := RefreshClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "dev",
			Id:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.JWTKey))
	refreshToken := RefreshToken{Value: tokenString, JTI: claims.Id}
	return refreshToken, err
}

func (j *JWTHelper) CheckJWTClaims(r *http.Request) (*JWTClaims, error) {
	jwtCookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, err
	}
	tokenString := jwtCookie.Value

	claims := &JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}

func (j *JWTHelper) CheckRefreshClaims(r *http.Request) (*RefreshClaims, error) {
	refreshCookie, err := r.Cookie("refresh")
	if err != nil {
		return nil, err
	}
	tokenString := refreshCookie.Value
	claims := &RefreshClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}
