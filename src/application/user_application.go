package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"log"
	"net/http"
)

type UserPayload struct {
	Email    string
	Password string
	Username string
}

type CognitoActions struct {
	CognitoClient *cognitoidentityprovider.Client
	ClientId      string
}

func (actor CognitoActions) Register(ctx context.Context, email string, password string) (bool, error) {
	output, err := actor.CognitoClient.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(actor.ClientId),
		Username: aws.String(email),
		Password: aws.String(password),
	})

	if err != nil {
		return false, err
	}

	return output.UserConfirmed, err
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	email := payload.Email
	password := payload.Password

	if email == "" || password == "" {
		http.Error(w, "Email and password required", 400)
	}

	device := w.Header().Get("user-agent")
	if device == "" {
		device = "null"
	}

	// return the tokens
}
