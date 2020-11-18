package main

import (
	"encoding/json"
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

// register GET
func registerPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "register.html", nil)
}

// register POST
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

// login GET
func loginPage(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "login.html", nil)
}

// login POST
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

	// TODO: Create JWT
	auth := &Auth{Username: user.Username, Token: "JWT"}
	json.NewEncoder(w).Encode(auth)
}
