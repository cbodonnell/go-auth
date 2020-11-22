package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Configuration struct
type Configuration struct {
	Debug         bool          `json:"debug"`
	Db            DataSource    `json:"db"`
	JWTKey        string        `json:"jwtKey"`
	JWTExpiration time.Duration `json:"jwtExpiration"`
}

// DataSource struct
type DataSource struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
}

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
}

// Group struct
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// JWTClaims struct
type JWTClaims struct {
	Username string  `json:"username"`
	Groups   []Group `json:"groups"`
	jwt.StandardClaims
}

// Auth struct
type Auth struct {
	Username string  `json:"username"`
	Groups   []Group `json:"groups"`
}

// Success struct
type Success struct {
	Title    string
	Route    string
	RouteMsg string
}
