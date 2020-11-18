package main

import "time"

// Configuration struct
type Configuration struct {
	Debug bool       `json:"debug"`
	Db    DataSource `json:"db"`
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

// Auth struct
type Auth struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}
