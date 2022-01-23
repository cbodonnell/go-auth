package responses

import (
	"net/http"

	"github.com/cheebz/logging"
)

type AuthResponses struct {
	Debug bool
}

func NewAuthResponses(debug bool) Responses {
	return &AuthResponses{
		Debug: debug,
	}
}

func (r *AuthResponses) BadRequest(w http.ResponseWriter, err error) {
	var msg string
	if r.Debug {
		msg = err.Error()
	} else {
		msg = "Bad request"
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func (r *AuthResponses) NotFound(w http.ResponseWriter, err error) {
	var msg string
	if r.Debug {
		msg = err.Error()
	} else {
		msg = "Not found"
	}
	http.Error(w, msg, http.StatusNotFound)
}

func (r *AuthResponses) UnauthorizedRequest(w http.ResponseWriter, err error) {
	var msg string
	if r.Debug {
		msg = err.Error()
	} else {
		msg = "Unauthorized"
	}
	http.Error(w, msg, http.StatusUnauthorized)
}

func (r *AuthResponses) InternalServerError(w http.ResponseWriter, err error) {
	logging.LogCaller(err)
	var msg string
	if r.Debug {
		msg = err.Error()
	} else {
		msg = "Internal server error"
	}
	http.Error(w, msg, http.StatusInternalServerError)
}
