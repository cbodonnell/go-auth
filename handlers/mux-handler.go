package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/cheebz/go-auth/captcha"
	"github.com/cheebz/go-auth/config"
	"github.com/cheebz/go-auth/hash"
	"github.com/cheebz/go-auth/jwt"
	"github.com/cheebz/go-auth/models"
	"github.com/cheebz/go-auth/repositories"
	"github.com/cheebz/go-auth/responses"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type MuxHandler struct {
	Conf      config.Configuration
	Responses responses.Responses
	Hasher    hash.Hash
	Repo      repositories.Repository
	JWT       *jwt.JWTHelper
	Templates *template.Template
	Router    *mux.Router
}

type MuxHandlerConfig struct {
	Conf      config.Configuration
	Resp      responses.Responses
	Hasher    hash.Hash
	Repo      repositories.Repository
	JWT       *jwt.JWTHelper
	Templates *template.Template
}

func NewMuxHandler(c MuxHandlerConfig) Handler {
	handler := &MuxHandler{
		Conf:      c.Conf,
		Responses: c.Resp,
		Hasher:    c.Hasher,
		Repo:      c.Repo,
		JWT:       c.JWT,
		Templates: c.Templates,
		Router:    mux.NewRouter(),
	}
	handler.setupRoutes()
	return handler
}

func (h *MuxHandler) setupRoutes() {
	h.Router.HandleFunc("/auth/", h.Home).Methods("GET")
	h.Router.HandleFunc("/auth/login", h.LoginPage).Methods("GET")
	h.Router.HandleFunc("/auth/login", h.Login).Methods("POST")
	h.Router.HandleFunc("/auth/password", h.PasswordPage).Methods("GET")
	h.Router.HandleFunc("/auth/password", h.Password).Methods("POST")
	h.Router.HandleFunc("/auth/logout", h.Logout).Methods("GET")
	h.Router.HandleFunc("/auth/logoutAll", h.LogoutAll).Methods("GET")
	if h.Conf.Register {
		h.Router.HandleFunc("/auth/register", h.RegisterPage).Methods("GET")
		h.Router.HandleFunc("/auth/register", h.Register).Methods("POST")
	}
}

func (h *MuxHandler) GetRouter() http.Handler {
	return h.Router
}

func (h *MuxHandler) AllowCORS(allowedOrigins []string) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
	}).Handler
	h.Router.Use(cors)
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

func (h *MuxHandler) refresh(w http.ResponseWriter, r *http.Request) (*jwt.JWTClaims, error) {
	refreshClaims, err := h.JWT.CheckRefreshClaims(r)
	if err != nil {
		return nil, err
	}
	err = h.Repo.ValidateRefresh(refreshClaims.UserID, refreshClaims.Id)
	if err != nil {
		return nil, err
	}
	user, err := h.Repo.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, err
	}
	groups, err := h.Repo.GetUserGroups(user.ID)
	if err != nil {
		return nil, err
	}
	jwt, err := h.JWT.CreateJWT(user, groups)
	if err != nil {
		return nil, err
	}
	err = h.Repo.InvalidateRefresh(refreshClaims.Id)
	if err != nil {
		log.Println("failed to invalidate refresh token", err)
	}
	refreshToken, err := h.JWT.CreateRefresh(user.ID)
	if err != nil {
		return nil, err
	}
	err = h.Repo.SaveRefresh(user.ID, refreshToken.JTI)
	if err != nil {
		return nil, err
	}
	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    jwt.Value,
		Path:     "/",
		MaxAge:   h.Conf.JWTMaxAge,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)
	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   h.Conf.RefreshMaxAge,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
	}
	http.SetCookie(w, refreshCookie)
	return &jwt.Claims, err
}

// Clear session cookies
func (h *MuxHandler) clearCookies(w http.ResponseWriter) {
	clearedJWTCookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
	}
	http.SetCookie(w, clearedJWTCookie)
	clearedRefreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
	}
	http.SetCookie(w, clearedRefreshCookie)
}

// Clears the current login session
func (h *MuxHandler) clearSession(w http.ResponseWriter, r *http.Request) error {
	h.clearCookies(w)
	refreshClaims, err := h.JWT.CheckRefreshClaims(r)
	if err != nil {
		return err
	}
	err = h.Repo.InvalidateRefresh(refreshClaims.Id)
	if err != nil {
		return err
	}
	return nil
}

// Clears the all login sessions for the user
func (h *MuxHandler) clearAllSessions(w http.ResponseWriter, r *http.Request) error {
	h.clearCookies(w)
	refreshClaims, err := h.JWT.CheckRefreshClaims(r)
	if err != nil {
		return err
	}
	err = h.Repo.DeleteAllRefresh(refreshClaims.UserID)
	if err != nil {
		return err
	}
	return nil
}

// / GET
func (h *MuxHandler) Home(w http.ResponseWriter, r *http.Request) {
	acceptJSON := acceptJSON(r)
	claims, err := h.JWT.CheckJWTClaims(r)
	if err != nil {
		claims, err = h.refresh(w, r)
		if err != nil {
			_ = h.clearSession(w, r)
			if !acceptJSON {
				http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
				return
			}
			h.Responses.UnauthorizedRequest(w, err)
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
		if err := h.Templates.ExecuteTemplate(w, "index.html", claims); err != nil {
			h.Responses.InternalServerError(w, err)
		}
		return
	}
	auth := &models.Auth{Username: claims.Username, UUID: claims.UUID, Groups: claims.Groups}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth)
}

// /register GET
func (h *MuxHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	_, err := h.JWT.CheckJWTClaims(r)
	if err == nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	if h.Conf.HCaptchaSecret != "" {
		if err := h.Templates.ExecuteTemplate(w, "register_hCaptcha.html", nil); err != nil {
			h.Responses.InternalServerError(w, err)
		}
	} else {
		if err := h.Templates.ExecuteTemplate(w, "register.html", nil); err != nil {
			h.Responses.InternalServerError(w, err)
		}
	}
}

// /register POST
func (h *MuxHandler) Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.Responses.BadRequest(w, err)
		return
	}

	if h.Conf.HCaptchaSecret != "" {
		hCaptchaResponse := r.PostForm.Get("h-captcha-response")
		err = captcha.ValidateHCaptcha(hCaptchaResponse, h.Conf.HCaptchaSecret)
		if err != nil {
			h.Responses.BadRequest(w, err)
			return
		}
	}

	username := r.PostForm.Get("username")
	user, err := h.Repo.GetUserByName(username)
	if err == nil {
		h.Responses.BadRequest(w, errors.New("user already exists"))
		return
	}
	user.Username = username

	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirm-password")
	if password != confirmPassword {
		h.Responses.BadRequest(w, errors.New("passwords to not match"))
		return
	}

	user.Password, err = h.Hasher.Generate(password)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	user.UUID = uuid.New().String()
	user.Created = time.Now()
	user, err = h.Repo.CreateUser(user)
	if err != nil {
		h.Responses.InternalServerError(w, err)
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
func (h *MuxHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	_, err := h.JWT.CheckJWTClaims(r)
	if err == nil {
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		h.Responses.InternalServerError(w, err)
	}
}

// /login POST
func (h *MuxHandler) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.Responses.BadRequest(w, err)
		return
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	user, err := h.Repo.GetUserByName(username)
	if err != nil {
		h.Responses.UnauthorizedRequest(w, err)
		return
	}

	err = h.Hasher.Check(user.Password, password)
	if err != nil {
		h.Responses.UnauthorizedRequest(w, err)
		return
	}

	groups, err := h.Repo.GetUserGroups(user.ID)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	jwt, err := h.JWT.CreateJWT(user, groups)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	refreshToken, err := h.JWT.CreateRefresh(user.ID)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	err = h.Repo.SaveRefresh(user.ID, refreshToken.JTI)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    jwt.Value,
		Path:     "/",
		MaxAge:   h.Conf.JWTMaxAge,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
	}
	http.SetCookie(w, jwtCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken.Value,
		Path:     "/",
		MaxAge:   h.Conf.RefreshMaxAge,
		HttpOnly: true,
		Secure:   h.Conf.SSLCert != "",
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
func (h *MuxHandler) PasswordPage(w http.ResponseWriter, r *http.Request) {
	_, err := h.JWT.CheckJWTClaims(r)
	if err != nil {
		_, err = h.refresh(w, r)
		if err != nil {
			_ = h.clearSession(w, r)
			h.Responses.UnauthorizedRequest(w, err)
			return
		}
	}
	if err := h.Templates.ExecuteTemplate(w, "password.html", nil); err != nil {
		h.Responses.InternalServerError(w, err)
	}
}

// /password POST
func (h *MuxHandler) Password(w http.ResponseWriter, r *http.Request) {
	claims, err := h.JWT.CheckJWTClaims(r)
	if err != nil {
		claims, err = h.refresh(w, r)
		if err != nil {
			_ = h.clearSession(w, r)
			http.Redirect(w, r, "/auth/", http.StatusSeeOther)
			return
		}
	}

	err = r.ParseForm()
	if err != nil {
		h.Responses.BadRequest(w, err)
		return
	}
	currentPassword := r.PostForm.Get("current-password")
	newPassword := r.PostForm.Get("new-password")
	confirmPassword := r.PostForm.Get("confirm-password")

	if newPassword == currentPassword {
		h.Responses.BadRequest(w, errors.New("new password is the same as current password"))
		return
	}

	if newPassword != confirmPassword {
		h.Responses.BadRequest(w, errors.New("passwords do not match"))
		return
	}

	user, err := h.Repo.GetUserByName(claims.Username)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	err = h.Hasher.Check(user.Password, currentPassword)
	if err != nil {
		h.Responses.BadRequest(w, errors.New("current password is incorrect"))
		return
	}

	password, err := h.Hasher.Generate(newPassword)
	if err != nil {
		h.Responses.InternalServerError(w, err)
		return
	}

	err = h.Repo.UpdatePassword(user.ID, password)
	if err != nil {
		h.Responses.InternalServerError(w, err)
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
func (h *MuxHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("refresh")
	if err != nil {
		_ = h.clearSession(w, r)
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	err = h.clearSession(w, r)
	if err != nil {
		h.Responses.InternalServerError(w, err)
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
func (h *MuxHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("refresh")
	if err != nil {
		_ = h.clearAllSessions(w, r)
		http.Redirect(w, r, "/auth/", http.StatusSeeOther)
		return
	}
	err = h.clearAllSessions(w, r)
	if err != nil {
		h.Responses.InternalServerError(w, err)
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
