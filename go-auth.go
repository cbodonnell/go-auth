package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// --- Configuration --- //

var config Configuration

// --- Main --- //

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
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

	if config.AllowedOrigins != "" {
		cors := cors.New(cors.Options{
			AllowedOrigins:   strings.Split(config.AllowedOrigins, ","),
			AllowCredentials: true,
		}).Handler
		r.Use(cors)
	}

	// Run server
	port := config.Port
	log.Println(fmt.Sprintf("Serving on port %d", port))

	if config.SSLCert == "" {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), config.SSLCert, config.SSLKey, r))
}
