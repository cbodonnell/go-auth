package main

import (
	"time"

	"github.com/golang-jwt/jwt"
)

// TemplateError struct
type TemplateError struct {
	Msg  string
	Data interface{}
}

// User struct
type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
	UUID     string    `json:"uuid"`
}

// Group struct
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// JWTClaims struct
type JWTClaims struct {
	UserID   int     `json:"user_id"`
	Username string  `json:"username"`
	UUID     string  `json:"uuid"`
	Groups   []Group `json:"groups"`
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

// Auth struct that is returned to user upon authentication
type Auth struct {
	Username string  `json:"username"`
	UUID     string  `json:"uuid"`
	Groups   []Group `json:"groups"`
}

// Success struct
type Success struct {
	Title    string
	Route    string
	RouteMsg string
}

// HCaptchaValidation struct
type HCaptchaValidation struct {
	Success bool `json:"success"`
}
