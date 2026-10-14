package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

type UserRepository struct {
	DB *sql.DB
}

func (u *UserRepository) CreateItemQuery(
	itemId string,
	sellerId string,
	itemName string,
	description string,
	price float64,
	quantityAvailable int,
	imageUrl string,
	categoryName string,
	tags []string,
) error {
	if len(tags) == 0 {
		log.Println("No tags provided for item")
		return errors.New("No tags provided for item")
	}

	// need to pass db and lower every tag
	for i, s := range tags {
		tags[i] = strings.ToLower(s)
	}

	output, err := u.DB.Exec()
}
