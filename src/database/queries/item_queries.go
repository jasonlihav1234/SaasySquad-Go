package queries

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strings"
	"time"
)

type ItemRepository struct {
	DB *pgxpool.Pool
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

type Review struct {
	ReviewID   uuid.UUID
	UserID     uuid.UUID
	ItemID     uuid.UUID
	Review     string
	ReviewDate time.Time
	Rating     int
}

type Tag struct {
	TagId   uuid.UUID
	TagName string
}

type ItemTag struct {
	ItemId  uuid.UUID
	TagId   uuid.UUID
	TagName string
}

func (ir *ItemRepository) CreateItemQuery(
	itemId uuid.UUID,
	sellerId string,
	itemName string,
	description string,
	price string,
	quantityAvailable int,
	imageUrl string,
	categoryName string,
	tags []string,
	ctx context.Context,
) (uuid.UUID, error) {
	if len(tags) == 0 {
		log.Println("No tags provided for item")
		return uuid.Nil, errors.New("No tags provided for item")
	}

	// need to pass db and lower every tag
	for i, s := range tags {
		tags[i] = strings.ToLower(s)
	}

	// even if ctx dies mid-request, TCP connection with database severed and rolls back
	tx, err := ir.DB.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	newCategoryId, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	var categoryId uuid.UUID
	err = tx.QueryRow(ctx, `
			insert into categories (category_id, category_name)
			values ($1, $2)
			on conflict (category_name)
			do update set category_name = excluded.category_name
			returning category_id`,
		newCategoryId, strings.ToLower(categoryName),
	).Scan(&categoryId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert category: %w", err)
	}

	_, err = tx.Exec(ctx, `
			insert into items (
				item_id, seller_id, item_name, description, price,
				quantity_available, image_url, category_id
			) values ($1, $2, $3, $4, $5::numeric, $6, $7, $8)`,
		itemId, sellerId, categoryName, description, price,
		quantityAvailable, imageUrl, categoryId,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert item: %w", err)
	}

	tagIds := make([]uuid.UUID, len(tags))
	for i := range tags {
		if tagIds[i], err = uuid.NewV7(); err != nil {
			return uuid.Nil, err
		}
	}

	_, err = tx.Exec(ctx, `
			insert into tags (tag_id, tag_name)
			select * from unnest($1::uuid[], $2::text[])
			on conflict (tag_name) do nothing`,
		tagIds, tags,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert tags: %w", err)
	}

	_, err = tx.Exec(ctx, `
			insert into item_tags (item_id, tag_id)
			select $1, tag_id from tags where tag_name = any ($2)
			on conflict do nothing`,
		itemId, tags,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("link tags: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit: %w", err)
	}

	return itemId, nil
}

func (ir *ItemRepository) GetItemByItemIdQuery(ctx context.Context, itemId string) (*Item, error) {
	queryString := `
		select *
		from items
		where item_id = $1
	`

	var itemPayload Item
	err := ir.DB.QueryRow(ctx, queryString).Scan(&itemPayload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("Item not found")
		}

		return nil, fmt.Errorf("Query failed: %v", err)
	}

	return &itemPayload, nil
}

func (ir *ItemRepository) GetReviewsByItemIdQuery(ctx context.Context, itemId string) ([]Review, error) {
	queryString := `
		select *
		from reviews
		where item_id = $1
		order by review_date desc
	`

	rows, err := ir.DB.Query(ctx, queryString, itemId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	reviews, err := pgx.CollectRows(rows, pgx.RowToStructByName[Review])
	if err != nil {
		return nil, fmt.Errorf("collect rreviews for item %s: %w", itemId, err)
	}

	return reviews, nil
}

func (ir *ItemRepository) GetItemTagsByItemIdQuery(ctx context.Context, itemId string) ([]string, error) {
	queryString := `
		select t.tag_name
		from item_tags it
		join tags t on it.tag_id == t.tag_id
		where it.item_id = $1
	`

	rows, err := ir.DB.Query(ctx, queryString, itemId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("collect tags for item: %s: %w", itemId, err)
	}

	return tags, nil
}

func (ir *ItemRepository) DeleteItemFromIdQuery(ctx context.Context, itemId string) (int64, error) {
	queryString := `
		delete
		form items
		where item_id = $1
	`

	response, err := ir.DB.Exec(ctx, queryString, itemId)
	if err != nil {
		return -1, fmt.Errorf("query failed: %w", err)
	}

	return response.RowsAffected(), nil
}
