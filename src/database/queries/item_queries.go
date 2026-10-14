package database

import (
	"context"
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
	ctx context.Context,
) error {
	if len(tags) == 0 {
		log.Println("No tags provided for item")
		return errors.New("No tags provided for item")
	}

	// need to pass db and lower every tag
	for i, s := range tags {
		tags[i] = strings.ToLower(s)
	}

	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	output, err := u.DB.ExecContext(ctx,
		`
		insert into categories (category_id, category_name)
		values ($1, $2)
		on conflict (category_name)
		do updat set category_name = excluded.category_name
		returning category_id
		`,
	)
}
