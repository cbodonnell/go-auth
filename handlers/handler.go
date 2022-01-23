package handlers

import "net/http"

type Handler interface {
	GetRouter() http.Handler
	AllowCORS(allowedOrigins []string)
	Home(w http.ResponseWriter, r *http.Request)
	RegisterPage(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	LoginPage(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	PasswordPage(w http.ResponseWriter, r *http.Request)
	Password(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	LogoutAll(w http.ResponseWriter, r *http.Request)
}
