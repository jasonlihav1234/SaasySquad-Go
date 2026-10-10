package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type UserPayload struct {
	Email    string
	Password string
	Username string
}

type UserConfirmPayload struct {
	Email            string
	ConfirmationCode string
}

type CognitoActions struct {
	CognitoClient *cognitoidentityprovider.Client
}

func (actor CognitoActions) Register(ctx context.Context, clientId string, email string, password string) (string, error) {
	output, err := actor.CognitoClient.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(clientId),
		Username: aws.String(email),
		Password: aws.String(password),
	})

	if err != nil {
		return "", err
	}

	return aws.ToString(output.UserSub), err
}

func RegisterHandler(ctx context.Context, cfg aws.Config, cognitoClientId string, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload UserPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		email := payload.Email
		password := payload.Password

		// sdk split into shared settings and service client, client needs settings
		actor := CognitoActions{
			CognitoClient: cognitoidentityprovider.NewFromConfig(cfg),
		}

		userId, err := actor.Register(ctx, cognitoClientId, email, password)
		if err != nil {
			log.Print("Cognito user failed to create")
			return
		}

		_, err = db.Exec("INSERT INTO user_v2 (id, email) VALUES ($1, $2)", userId, email)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Println("Created User")
	}
}

func ConfirmRegisterHandler(cfg aws.Config, cognitoClientId string, db *sql.DB) http.HandlerFunc {
	actor := CognitoActions{
		CognitoClient: cognitoidentityprovider.NewFromConfig(cfg),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// payload would be email and confirmation code
		var payload UserConfirmPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		defer r.Body.Close()
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		email := payload.Email
		confirmationCode := payload.ConfirmationCode

		if email == "" || confirmationCode == "" {
			log.Print("No email or confirmation code provided")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		actor.ConfirmRegister()

		fmt.Println("Confirmed User Register")
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
