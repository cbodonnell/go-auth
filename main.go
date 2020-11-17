package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// --- Configuration --- //

var config Configuration

func getConfig(ENV string) Configuration {
	file, _ := os.Open(fmt.Sprintf("config.%s.json", ENV))
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		panic(err)
	}
	return config
}

// --- Models --- //

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

// --- Data --- //

// db instance
var db *sql.DB

// connect to db
func connectDb(s DataSource) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		s.Host, s.Port, s.User, s.Password, s.Dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected to %s as %s\n", s.Dbname, s.User)
	return db
}

// --- Templates --- //

var templates = template.Must(template.ParseGlob("templates/*.html"))

// --- Handlers -- //

// register GET
func registerPage(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		internalServerError(w, err)
		return
	}
}

// register POST
func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}
	username := r.PostForm.Get("username")
	// TODO: check if user exists

	password := r.PostForm.Get("password")
	confirmPassword := r.PostForm.Get("confirm-password")
	if password != confirmPassword {
		badRequest(w, errors.New("passwords do not match"))
		return
	}

	hash, err := generateHash(password)
	if err != nil {
		internalServerError(w, err)
		return
	}

	// TODO: INSERT INTO users (username, hash, createdDate) VALUES ($1, $2, NOW());

	fmt.Fprintln(w, "Username: "+username)
	fmt.Fprintln(w, "Hashed PW: "+hash)
}

// login GET
func loginPage(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		internalServerError(w, err)
		return
	}
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

	// TODO: SELECT hash FROM users WHERE username = $1;
	hash := "$2a$14$V.N72q6u9MNOzkCbLdKOgOxW01mcrdXN0AuDDu5qWrv5ggUo0YnyC"
	// If there is no hash, then the user does not exist

	err = checkHash(hash, password)
	if err != nil {
		unauthorizedRequest(w, errors.New("invalid credentials"))
		return
	}

	fmt.Fprintln(w, fmt.Sprintf("Signed in as: %s", username))
}

// --- Crypto --- //

// Generate Salt
// func generateSalt() (string, error) {
// 	saltBytes := make([]byte, 32)
// 	_, err := io.ReadFull(rand.Reader, saltBytes)
// 	salt := hex.EncodeToString(saltBytes)
// 	return salt, err
// }

// Generate a bcrypt Hash (see: https://en.wikipedia.org/wiki/Bcrypt)
func generateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// --- Responses --- //

func badRequest(w http.ResponseWriter, err error) {
	var msg string
	if config.Debug {
		msg = err.Error()
	} else {
		msg = "Bad request"
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func unauthorizedRequest(w http.ResponseWriter, err error) {
	var msg string
	if config.Debug {
		msg = err.Error()
	} else {
		msg = "Unauthorized"
	}
	http.Error(w, msg, http.StatusUnauthorized)
}

func internalServerError(w http.ResponseWriter, err error) {
	var msg string
	if config.Debug {
		msg = err.Error()
	} else {
		msg = "Internal server error"
	}
	http.Error(w, msg, http.StatusInternalServerError)
}

// --- Main --- //

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
	if ENV == "" {
		ENV = "dev"
	}
	fmt.Println(fmt.Sprintf("Running in ENV: %s", ENV))
	config = getConfig(ENV)

	db = connectDb(config.Db)
	defer db.Close()

	// Init router
	r := mux.NewRouter()

	// Route handlers

	r.HandleFunc("/register", registerPage).Methods("GET")
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", loginPage).Methods("GET")
	r.HandleFunc("/login", login).Methods("POST")

	// CORS
	// handler := cors.Default().Handler(r)

	// Run server
	port := 8080
	fmt.Println(fmt.Sprintf("Serving on port %d", port))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
