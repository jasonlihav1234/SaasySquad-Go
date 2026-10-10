package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
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

func (actor CognitoActions) ConfirmRegister(ctx context.Context, clientId string, email string, confirmationCode string) error {
	_, err := actor.CognitoClient.ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(clientId),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(confirmationCode),
	})

	return err
}

func (actor CognitoActions) Login(ctx context.Context, clientId string, email string, password string) (*cognitotypes.AuthenticationResultType, error) {
	output, err := actor.CognitoClient.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(clientId),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	})

	if err != nil {
		return nil, err
	}

	return output.AuthenticationResult, err
}

func RegisterHandler(ctx context.Context, cfg aws.Config, cognitoClientId string, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload UserPayload
		defer r.Body.Close()

		err := json.NewDecoder(r.Body).Decode(&payload)
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
		defer r.Body.Close()

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		email := payload.Email
		confirmationCode := payload.ConfirmationCode

		if email == "" || confirmationCode == "" {
			log.Print("No email or confirmation code provided")
			http.Error(w, "No email or confirmation code provided", http.StatusBadRequest)
			return
		}

		err = actor.ConfirmRegister(r.Context(), cognitoClientId, email, confirmationCode)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Confirmed User Register")
	}
}

func LoginHandler(cfg aws.Config, cognitoClientId string, db *sql.DB) http.HandlerFunc {
	actor := CognitoActions{
		CognitoClient: cognitoidentityprovider.NewFromConfig(cfg),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var payload UserPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		defer r.Body.Close()

		if err != nil {
			log.Println("Failed to unmarshal login information")
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.Email)
		password := payload.Password

		if email == "" || password == "" {
			log.Println("Email or password not passed in")
			http.Error(w, "Email and password required", 400)
			return
		}

		device := w.Header().Get("user-agent")
		if device == "" {
			device = "null"
		}

		output, err := actor.Login(r.Context(), cognitoClientId, email, password)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return

		}

		if output == nil || output.AccessToken == nil {
			http.Error(w, "Unexpected response from auth provider", http.StatusInternalServerError)
			return
		}

		accessToken := output.AccessToken
		refreshToken := output.RefreshToken

		if refreshToken != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    *refreshToken,
				Path:     "/auth/refresh", // browser only sends to this endpoint, not attached on every request
				MaxAge:   30 * 24 * 60 * 60,
				HttpOnly: true,                    // javascript cannot read it, limits damage from xss
				Secure:   false,                   // secure = true means that only sent over HTTPS
				SameSite: http.SameSiteStrictMode, // never sent on cross-site requests
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store") // browsers don't cache token
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": *accessToken,
			"token_type":   "Bearer",
			"expires_in":   output.ExpiresIn,
		})
	}
}
