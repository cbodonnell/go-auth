package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, template string, data interface{}) {
	err := templates.ExecuteTemplate(w, template, data)
	if err != nil {
		internalServerError(w, err)
	}
}

// / GET
func home(w http.ResponseWriter, r *http.Request) {
	claims, err := checkClaims(r)
	if err != nil {
		unauthorizedRequest(w, err)
		return
	}
	auth := &Auth{claims.Username, claims.UUID, claims.Groups}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth)
}

// /home GET
func homePage(w http.ResponseWriter, r *http.Request) {
	claims, err := checkClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "index.html", claims)
}

// /register GET
func registerPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkClaims(r)
	if err == nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "register.html", nil)
}

// /register POST
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}

	hCaptchaResponse := r.PostForm.Get("h-captcha-response")
	err = validateCaptcha(hCaptchaResponse)
	if err != nil {
		badRequest(w, err)
		return
	}

	username := r.PostForm.Get("username")
	user, err := getUserByName(username)
	if err == nil {
		badRequest(w, err)
		return
	}
	user.Username = username

	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirm-password")
	if password != confirmPassword {
		badRequest(w, err)
		return
	}

	user.Password, err = generateHash(password)
	if err != nil {
		internalServerError(w, err)
		return
	}

	user.UUID = generateUUID()
	user.Created = time.Now()
	user, err = createUser(user)
	if err != nil {
		internalServerError(w, err)
		return
	}

	fmt.Fprintln(w, "Registration successful")
}

// /login GET
func loginPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkClaims(r)
	if err == nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "login.html", nil)
}

// /login POST
func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	user, err := getUserByName(username)
	if err != nil {
		unauthorizedRequest(w, err)
		return
	}

	err = checkHash(user.Password, password)
	if err != nil {
		unauthorizedRequest(w, err)
		return
	}

	groups, err := getUserGroups(user.ID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	tokenString, err := createJWT(user, groups)
	if err != nil {
		internalServerError(w, err)
		return
	}

	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(config.JWTExpiration * time.Minute),
		HttpOnly: true,
	}
	if config.SSLCert != "" {
		jwtCookie.Secure = true
	}
	http.SetCookie(w, jwtCookie)

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect == "" {
		redirect = "/auth/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// /password GET
func passwordPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "password.html", nil)
}

// /password POST
func password(w http.ResponseWriter, r *http.Request) {
	claims, err := checkClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}
	currentPassword := r.PostForm.Get("current-password")
	newPassword := r.PostForm.Get("new-password")
	confirmPassword := r.PostForm.Get("confirm-password")

	if newPassword == currentPassword {
		badRequest(w, errors.New("new password is the same as current password"))
		return
	}

	if newPassword != confirmPassword {
		badRequest(w, errors.New("passwords do not match"))
		return
	}

	user, err := getUserByName(claims.Username)
	if err != nil {
		internalServerError(w, err)
		return
	}

	err = checkHash(user.Password, currentPassword)
	if err != nil {
		badRequest(w, errors.New("current password is incorrect"))
		return
	}

	password, err := generateHash(newPassword)
	if err != nil {
		internalServerError(w, err)
		return
	}

	err = updatePassword(user.ID, password)
	if err != nil {
		internalServerError(w, err)
		return
	}

	fmt.Fprintln(w, "Password changed")
}

// /logout GET
func logout(w http.ResponseWriter, r *http.Request) {
	_, err := checkClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, jwtCookie)

	fmt.Fprintln(w, "Logged out")
}
