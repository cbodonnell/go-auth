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

func refresh(w http.ResponseWriter, r *http.Request) (*JWTClaims, error) {
	refreshClaims, err := checkRefreshClaims(r)
	if err != nil {
		return nil, err
	}
	err = validateRefresh(refreshClaims.UserID, refreshClaims.Id)
	if err != nil {
		return nil, err
	}
	user, err := getUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, err
	}
	groups, err := getUserGroups(user.ID)
	if err != nil {
		return nil, err
	}
	jwt, err := createJWT(user, groups)
	if err != nil {
		return nil, err
	}
	err = deleteRefresh(refreshClaims.Id)
	refreshToken, err := createRefresh(user.ID)
	if err != nil {
		return nil, err
	}
	err = saveRefresh(user.ID, refreshToken.JTI)
	if err != nil {
		return nil, err
	}
	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    jwt.Value,
		Path:     "/",
		Expires:  time.Now().Add(config.JWTExpiration * time.Minute),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		Expires:  time.Now().Add(config.RefreshExpiration * time.Minute),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, refreshCookie)
	return &jwt.Claims, err
}

// Clear session cookies
func clearCookies(w http.ResponseWriter) {
	clearedJWTCookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, clearedJWTCookie)
	clearedRefreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, clearedRefreshCookie)
}

// Clears the current login session
func clearSession(w http.ResponseWriter, r *http.Request) error {
	clearCookies(w)
	refreshClaims, err := checkRefreshClaims(r)
	if err != nil {
		return err
	}
	err = deleteRefresh(refreshClaims.Id)
	if err != nil {
		return err
	}
	return nil
}

// Clears the all login sessions for the user
func clearAllSessions(w http.ResponseWriter, r *http.Request) error {
	clearCookies(w)
	refreshClaims, err := checkRefreshClaims(r)
	if err != nil {
		return err
	}
	err = deleteAllRefresh(refreshClaims.UserID)
	if err != nil {
		return err
	}
	return nil
}

// / GET
func home(w http.ResponseWriter, r *http.Request) {
	refreshClaims, err := checkRefreshClaims(r)
	if err != nil {
		_ = clearSession(w, r)
		unauthorizedRequest(w, err)
		return
	}
	err = validateRefresh(refreshClaims.UserID, refreshClaims.Id)
	if err != nil {
		_ = clearSession(w, r)
		unauthorizedRequest(w, err)
		return
	}
	claims, err := checkJWTClaims(r)
	if err != nil {
		claims, err = refresh(w, r)
		if err != nil {
			_ = clearSession(w, r)
			unauthorizedRequest(w, err)
			return
		}
	}
	auth := &Auth{claims.Username, claims.UUID, claims.Groups}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth)
}

// /home GET
func homePage(w http.ResponseWriter, r *http.Request) {
	claims, err := checkJWTClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "index.html", claims)
}

// /register GET
func registerPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkJWTClaims(r)
	if err == nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	if config.HCaptchaSecret != "" {
		renderTemplate(w, "register_hCaptcha.html", nil)
	} else {
		renderTemplate(w, "register.html", nil)
	}
}

// /register POST
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}

	if config.HCaptchaSecret != "" {
		hCaptchaResponse := r.PostForm.Get("h-captcha-response")
		err = validateCaptcha(hCaptchaResponse)
		if err != nil {
			badRequest(w, err)
			return
		}
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
	_, err := checkJWTClaims(r)
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

	jwt, err := createJWT(user, groups)
	if err != nil {
		internalServerError(w, err)
		return
	}

	refreshToken, err := createRefresh(user.ID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	err = saveRefresh(user.ID, refreshToken.JTI)
	if err != nil {
		internalServerError(w, err)
		return
	}

	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    jwt.Value,
		Path:     "/",
		Expires:  time.Now().Add(config.JWTExpiration * time.Minute),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		Expires:  time.Now().Add(config.RefreshExpiration * time.Minute),
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, refreshCookie)

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect == "" {
		redirect = "/auth/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// /password GET
func passwordPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkJWTClaims(r)
	if err != nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "password.html", nil)
}

// /password POST
func password(w http.ResponseWriter, r *http.Request) {
	claims, err := checkJWTClaims(r)
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
	_, err := r.Cookie("refresh")
	if err != nil {
		_ = clearSession(w, r)
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	err = clearSession(w, r)
	if err != nil {
		internalServerError(w, err)
		return
	}

	fmt.Fprintln(w, "Logged out")
}

// /logoutAll GET
func logoutAll(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("refresh")
	if err != nil {
		_ = clearAllSessions(w, r)
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	err = clearAllSessions(w, r)
	if err != nil {
		internalServerError(w, err)
		return
	}

	fmt.Fprintln(w, "Logged out all sessions")
}
