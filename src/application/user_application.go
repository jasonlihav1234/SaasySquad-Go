package application

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserPayload struct {
	Email    string
	Password string
	Username string
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	email := payload.Email
	password := payload.Password
	username := payload.Username

	if email == "" || password == "" || username == "" {
		http.Error(w, "Email, password, and username required", 400)
		return
	}

	if len(password) < 7 {
		http.Error(w, "Password must be at leat 7 characters long", 400)
		return
	}

	fmt.Println("Generating user...")
}
