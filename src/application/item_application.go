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
	"user_application"
)

type ItemPayload struct {
	itemName          string
	description       string
	price             float64
	quantityAvailable int
	imageUrl          string
	categoryName      string
	tags              []string
}

func createItem(cfg aws.Config) http.HandlerFunc {
	actor := CognitoActions{
		cognitoidentityprovider.NewFromConfig(cfg),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body.Close()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Failed to obtain authorization header")
			http.Error(w, "Failed to obtain authorization header", 500)
			return
		}

		authSplit := strings.SplitN(authHeader, " ", 2)

		if len(authSplit) != 2 || strings.ToLower(authSplit[0]) != "bearer" {
			log.Println("Authorization header invalid")
			http.Error(w, "Authorization header invalid", 500)
			return
		}

		accessToken := authSplit[1]
		output, err := actor.GetProfile(r.Context(), accessToken)

		if err != nil {
			log.Println("Failed to obtain profile")
			http.Error(w, "Failed to obtain profile", 500)
			return
		}

		userId := output.Username
		if userId == nil {
			log.Println("Failed to obtain userId")
			http.Error(w, "Failed to obtain userId", 500)
			return
		}

		var payload ItemPayload
		err = json.NewDecoder(r.Body).Decode(&payload)

		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Failed to decode body", 500)
			return
		}

		if payload.categoryName == "" ||
			payload.description == "" ||
			payload.imageUrl == "" ||
			payload.itemName == "" ||
			payload.price == 0.0 ||
			payload.quantityAvailable == 0 ||
			len(payload.tags) == 0 {
			log.Println("Missing fields in payload")
			http.Error(w, "Missing fields in item payload", 500)
			return
		}

	}
}
