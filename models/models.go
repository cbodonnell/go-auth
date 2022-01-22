package models

import (
	"time"
)

// User struct -- This is the user model
type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
	UUID     string    `json:"uuid"`
}

// Group struct -- This is the group model
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Auth struct that is returned to user upon authentication
type Auth struct {
	Username string  `json:"username"`
	UUID     string  `json:"uuid"`
	Groups   []Group `json:"groups"`
}
