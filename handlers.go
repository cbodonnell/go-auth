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

func acceptJSON(r *http.Request) bool {
	h := r.Header.Values("Accept")
	for _, v := range h {
		if v == "application/json" {
			return true
		}
	}
	return false
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
	err = invalidateRefresh(refreshClaims.Id)
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
		MaxAge:   config.JWTMaxAge,
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   config.RefreshMaxAge,
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
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, clearedJWTCookie)
	clearedRefreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
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
	err = invalidateRefresh(refreshClaims.Id)
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
	acceptJSON := acceptJSON(r)
	claims, err := checkJWTClaims(r)
	if err != nil {
		claims, err = refresh(w, r)
		if err != nil {
			_ = clearSession(w, r)
			if !acceptJSON {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}
			unauthorizedRequest(w, err)
			return
		}
	}
	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	if !acceptJSON {
		renderTemplate(w, "index.html", claims)
		return
	}
	auth := &Auth{claims.Username, claims.UUID, claims.Groups}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth)
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
		badRequest(w, errors.New("user already exists"))
		return
	}
	user.Username = username

	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirm-password")
	if password != confirmPassword {
		badRequest(w, errors.New("passwords to not match"))
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

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
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
		MaxAge:   config.JWTMaxAge,
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   config.RefreshMaxAge,
		HttpOnly: true,
		Secure:   config.SSLCert != "",
	}
	http.SetCookie(w, refreshCookie)

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/auth/", http.StatusSeeOther)
}

// /password GET
func passwordPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkJWTClaims(r)
	if err != nil {
		_, err = refresh(w, r)
		if err != nil {
			_ = clearSession(w, r)
			unauthorizedRequest(w, err)
			return
		}
	}
	renderTemplate(w, "password.html", nil)
}

// /password POST
func password(w http.ResponseWriter, r *http.Request) {
	claims, err := checkJWTClaims(r)
	if err != nil {
		claims, err = refresh(w, r)
		if err != nil {
			_ = clearSession(w, r)
			http.Redirect(w, r, "/auth/", http.StatusSeeOther)
			return
		}
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

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
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

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
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

	query := r.URL.Query()
	redirect := query.Get("redirect")
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}
	fmt.Fprintln(w, "Logged out all sessions")
}
