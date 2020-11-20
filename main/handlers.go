package main

import (
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
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	auth := &Auth{Username: claims.Username, Groups: claims.Groups}
	renderTemplate(w, "index.html", auth)
}

// /register GET
func registerPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register.html", nil)
}

// /register POST
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}

	username := r.PostForm.Get("username")
	user, err := getUserByName(username)
	if err == nil {
		templateError := TemplateError{Msg: "username already exists"}
		renderTemplate(w, "register.html", templateError)
		return
	}
	user.Username = username

	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirm-password")
	if password != confirmPassword {
		templateError := TemplateError{Msg: "passwords do not match", Data: user}
		renderTemplate(w, "register.html", templateError)
		return
	}

	user.Password, err = generateHash(password)
	if err != nil {
		internalServerError(w, err)
		return
	}

	user.Created = time.Now()
	user, err = createUser(user)
	if err != nil {
		internalServerError(w, err)
		return
	}

	err = templates.ExecuteTemplate(w, "registrationSuccessful.html", nil)
	if err != nil {
		internalServerError(w, err)
		return
	}
}

// /login GET
func loginPage(w http.ResponseWriter, r *http.Request) {
	_, err := checkClaims(r)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	user, err := getUserByName(username)
	if err != nil {
		templateError := TemplateError{Msg: "user does not exist"}
		renderTemplate(w, "login.html", templateError)
		return
	}

	err = checkHash(user.Password, password)
	if err != nil {
		templateError := TemplateError{Msg: "invalid credentials"}
		renderTemplate(w, "login.html", templateError)
		return
	}

	groups, err := getUserGroups(user.ID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	tokenString, err := createJWT(user.Username, groups)
	if err != nil {
		internalServerError(w, err)
		return
	}

	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, jwtCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// /logout GET
func logout(w http.ResponseWriter, r *http.Request) {
	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, jwtCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
