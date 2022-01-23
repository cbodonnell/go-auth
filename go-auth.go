package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cheebz/go-auth/config"
	"github.com/cheebz/go-auth/handlers"
	"github.com/cheebz/go-auth/hash"
	"github.com/cheebz/go-auth/jwt"
	"github.com/cheebz/go-auth/repositories"
	"github.com/cheebz/go-auth/responses"
	"github.com/cheebz/go-auth/workers"
)

func main() {
	// Get configuration
	ENV := os.Getenv("ENV")
	c, err := config.ReadConfig(ENV)
	if err != nil {
		log.Fatal(err)
	}
	conf := c

	// create repository
	repo := repositories.NewPSQLRepository(conf)
	defer repo.Close()
	// create purge refresh worker
	purgeRefreshWorker := workers.NewPurgeRefreshWorker(repo)
	go purgeRefreshWorker.Start()
	// create response writer
	response := responses.NewAuthResponses(conf.Debug)
	// create hasher
	hasher := hash.NewBCryptHash(14)
	// create handler
	jwt := jwt.NewJWTHelper(conf.JWTKey, conf.JWTMaxAge, conf.RefreshMaxAge)
	// parse template files
	templates := template.Must(template.ParseGlob("templates/*.html"))
	// create handler
	handler := handlers.NewMuxHandler(handlers.MuxHandlerConfig{
		Conf:      conf,
		Resp:      response,
		Hasher:    hasher,
		Repo:      repo,
		JWT:       jwt,
		Templates: templates,
	})
	if conf.AllowedOrigins != "" {
		handler.AllowCORS(strings.Split(conf.AllowedOrigins, ","))
	}
	r := handler.GetRouter()

	// Run server
	port := conf.Port
	log.Println(fmt.Sprintf("Serving on port %d", port))

	if conf.SSLCert == "" {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	}
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), conf.SSLCert, conf.SSLKey, r))
}
