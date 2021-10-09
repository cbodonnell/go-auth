package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Auth struct returned upon successful authentication
type Auth struct {
	Username string  `json:"username"`
	UUID     string  `json:"uuid"`
	Groups   []Group `json:"groups"`
}

// Group struct
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func Authenticate(w http.ResponseWriter, r *http.Request, endpoint string) (Auth, error) {
	var auth Auth
	client := &http.Client{}
	authReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return auth, err
	}
	for _, cookie := range r.Cookies() {
		authReq.AddCookie(cookie)
	}
	authReq.Header.Set("Accept", "application/json")
	authResp, err := client.Do(authReq)
	if err != nil {
		return auth, err
	}
	for _, cookie := range authResp.Cookies() {
		http.SetCookie(w, cookie)
		r.AddCookie(cookie)
	}
	if authResp.StatusCode != http.StatusOK {
		return auth, fmt.Errorf("received status code %d from auth endpoint", authResp.StatusCode)
	}
	defer authResp.Body.Close()
	err = json.NewDecoder(authResp.Body).Decode(&auth)
	if err != nil {
		return auth, err
	}
	return auth, nil
}
