package application

import (
	// "context"
	// "database/sql"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	// cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	// "user_application"
	"SaasySquad-Go/src/database/queries"
	"SaasySquad-Go/src/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
)

type ItemPayload struct {
	ItemName          string
	Description       string
	Price             float64
	QuantityAvailable int
	ImageUrl          string
	CategoryName      string
	Tags              []string
}

func CreateItemHandler(cfg aws.Config, db *pgxpool.Pool) http.HandlerFunc {
	actor := CognitoActions{
		cognitoidentityprovider.NewFromConfig(cfg),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

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

		if payload.CategoryName == "" ||
			payload.Description == "" ||
			payload.ImageUrl == "" ||
			payload.ItemName == "" ||
			payload.Price == 0.0 ||
			payload.QuantityAvailable == 0 ||
			len(payload.Tags) == 0 {
			log.Println("Missing fields in payload")
			http.Error(w, "Missing fields in item payload", 500)
			return
		}

		itemId, err := uuid.NewV7()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Failed to create item id", 500)
			return
		}

		itemRepo := queries.ItemRepository{
			DB: db,
		}

		itemOutput, err := itemRepo.CreateItemQuery(itemId,
			*userId,
			payload.ItemName,
			payload.Description,
			strconv.FormatFloat(payload.Price, 'f', 2, 64),
			payload.QuantityAvailable,
			payload.ImageUrl,
			payload.CategoryName,
			payload.Tags,
			r.Context())
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Failed to create item", 500)
			return
		}

		err = json.NewEncoder(w).Encode(itemOutput)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Failed to return item", 500)
			return
		}
	}
}

func GetItemById(cfg aws.Config, db *pgxpool.Pool) http.HandlerFunc {
	actor := CognitoActions{
		CognitoClient: cognitoidentityprovider.NewFromConfig(cfg),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		itemId := r.PathValue("itemId")
		if itemId == "" {
			log.Println("No itemId provided")
			http.Error(w, "No itemId provided", 500)
			return
		}

		itemRepo := queries.ItemRepository{
			DB: db,
		}

		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok {
			log.Println("Failed to obtain claim")
			http.Error(w, "failed to obtain claim", 500)
			return
		}

		userStringID := claims.ClientID
		if userStringID == "" {
			log.Println("Invalid user id")
			http.Error(w, "Invalid user id", 500)
			return
		}

		userID, err := uuid.Parse(userStringID)
		if err != nil {
			log.Println("Failed to convert userID to uuid")
			http.Error(w, "Failed to convert userID to uuid", 500)
			return
		}

		item, err := itemRepo.GetItemByItemIdQuery(r.Context(), itemId)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Failed to get item by item id", 500)
			return
		}

		if item.SellerId != userID {
			log.Println("Seller ID does not match user")
			http.Error(w, "Seller ID does not match user", 500)
			return
		}
	}
}
