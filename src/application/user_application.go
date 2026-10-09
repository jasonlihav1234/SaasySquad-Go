package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"net/http"
)

type UserPayload struct {
	Email    string
	Password string
	Username string
}

type CognitoActions struct {
	CognitoClient *cognitoidentityprovider.Client
}

func (actor CognitoActions) Register(ctx context.Context, clientId string, email string, password string) (bool, error) {
	output, err := actor.CognitoClient.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(clientId),
		Username: aws.String(email),
		Password: aws.String(password),
	})

	if err != nil {
		return false, err
	}

	return output.UserConfirmed, err
}

func RegisterHandler(ctx context.Context, cfg aws.Config, cognitoClientId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload UserPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		email := payload.Email
		password := payload.Password

		actor := CognitoActions{
			CognitoClient: cognitoidentityprovider.NewFromConfig(cfg),
		}

		actor.Register(ctx, cognitoClientId, email, password)

		fmt.Println("Created User")
	}
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
