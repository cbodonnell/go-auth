package main

import (
	"net/http"

	"github.com/cheebz/logging"
)

func badRequest(w http.ResponseWriter, err error) {
	var msg string
	if conf.Debug {
		msg = err.Error()
	} else {
		msg = "Bad request"
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func unauthorizedRequest(w http.ResponseWriter, err error) {
	var msg string
	if conf.Debug {
		msg = err.Error()
	} else {
		msg = "Unauthorized"
	}
	http.Error(w, msg, http.StatusUnauthorized)
}

func internalServerError(w http.ResponseWriter, err error) {
	logging.LogCaller(err)
	var msg string
	if conf.Debug {
		msg = err.Error()
	} else {
		msg = "Internal server error"
	}
	http.Error(w, msg, http.StatusInternalServerError)
}
