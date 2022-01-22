package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cheebz/go-auth/config"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// --- Configuration --- //

var conf config.Configuration

// --- Main --- //

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
	c, err := config.ReadConfig(ENV)
	if err != nil {
		log.Fatal(err)
	}
	conf = c

	db = connectDb(conf.Db)
	defer db.Close()

	// Workers
	go startPurgeRefresh()

	// Init router
	r := mux.NewRouter()
	r.HandleFunc("/auth/", home).Methods("GET")
	r.HandleFunc("/auth/login", loginPage).Methods("GET")
	r.HandleFunc("/auth/login", login).Methods("POST")
	r.HandleFunc("/auth/password", passwordPage).Methods("GET")
	r.HandleFunc("/auth/password", password).Methods("POST")
	r.HandleFunc("/auth/logout", logout).Methods("GET")
	r.HandleFunc("/auth/logoutAll", logoutAll).Methods("GET")
	if conf.Register {
		r.HandleFunc("/auth/register", registerPage).Methods("GET")
		r.HandleFunc("/auth/register", register).Methods("POST")
	}

	if conf.AllowedOrigins != "" {
		cors := cors.New(cors.Options{
			AllowedOrigins:   strings.Split(conf.AllowedOrigins, ","),
			AllowCredentials: true,
		}).Handler
		r.Use(cors)
	}

	// Run server
	port := conf.Port
	log.Println(fmt.Sprintf("Serving on port %d", port))

	if conf.SSLCert == "" {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), conf.SSLCert, conf.SSLKey, r))
}
