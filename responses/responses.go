package responses

import "net/http"

type Responses interface {
	BadRequest(w http.ResponseWriter, err error)
	NotFound(w http.ResponseWriter, err error)
	UnauthorizedRequest(w http.ResponseWriter, err error)
	InternalServerError(w http.ResponseWriter, err error)
}
