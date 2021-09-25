package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// --- Configuration --- //

var config Configuration

// --- Main --- //

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
	if ENV == "" {
		ENV = "dev"
	}
	fmt.Println(fmt.Sprintf("Running in ENV: %s", ENV))
	c, err := ReadConfig(ENV)
	if err != nil {
		log.Fatal(err)
	}
	config = c

	db = connectDb(config.Db)
	defer db.Close()

	// Workers
	startPurgeRefresh()

	// Init router
	r := mux.NewRouter()
	r.HandleFunc("/auth/", home).Methods("GET")
	r.HandleFunc("/auth/login", loginPage).Methods("GET")
	r.HandleFunc("/auth/login", login).Methods("POST")
	r.HandleFunc("/auth/password", passwordPage).Methods("GET")
	r.HandleFunc("/auth/password", password).Methods("POST")
	r.HandleFunc("/auth/logout", logout).Methods("GET")
	r.HandleFunc("/auth/logoutAll", logoutAll).Methods("GET")
	if config.Register {
		r.HandleFunc("/auth/register", registerPage).Methods("GET")
		r.HandleFunc("/auth/register", register).Methods("POST")
	}

	// CORS in dev environment
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowCredentials: true,
	}).Handler

	// Run server
	port := config.Port
	fmt.Println(fmt.Sprintf("Serving on port %d", port))

	if config.SSLCert == "" {
		r.Use(cors)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), config.SSLCert, config.SSLKey, r))
}
