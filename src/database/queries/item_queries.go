package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type UserRepository struct {
	DB *sql.DB
}

type Item struct {
	ItemId            uuid.UUID
	SellerId          uuid.UUID
	ItemName          string
	Description       string
	Price             string
	QuantityAvailable int
	ImageUrl          string
	CreatedAt         time.Time
	LastUpdated       time.Time
	CategoryId        uuid.UUID
}

func (u *UserRepository) CreateItemQuery(
	itemId string,
	sellerId string,
	itemName string,
	description string,
	price string,
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

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	var categoryId uuid.UUID
	err = u.DB.QueryRowContext(ctx,
		`
		insert into categories (category_id, category_name)
		values ($1, $2)
		on conflict (category_name)
		do update set category_name = excluded.category_name
		returning category_id
		`,
		id, strings.ToLower(categoryName),
	).Scan(&categoryId)
	if err != nil {
		return err
	}

	var item Item
	itemQuery := `
	insert into items (
		item_id,
		seller_id,
		item_name,
		description,
		price,
		quantity_available,
		image_url,
		created_at,
		last_updated,
		category_id
	)
	values (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10
	)
	returning *
	`
	isoString := time.Now().Format(time.RFC3339)
	err = u.DB.QueryRowContext(ctx, itemQuery, itemId, sellerId, itemName, description, price, quantityAvailable, imageUrl, isoString, isoString, categoryId).Scan(&item)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
