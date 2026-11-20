package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"{{ cookiecutter.import_path }}/internal/model"

	"github.com/google/uuid"
)

// ItemRead is the output schema for returning an item.
type ItemRead struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// toItemRead converts a model.Item to ItemRead.
func toItemRead(item model.Item) ItemRead {
	result := ItemRead{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	if item.CreatedBy != nil {
		result.CreatedBy = item.CreatedBy
	}
	return result
}

// GetItem retrieves an item by ID.
func GetItem(ctx context.Context, id uuid.UUID) (ItemRead, error) {
	item := &model.Item{}
	err := db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ItemRead{}, ErrNotFound
		}
		return ItemRead{}, err
	}
	return toItemRead(*item), nil
}

// ListItems returns a list of items with pagination.
func ListItems(ctx context.Context, limit, offset int64) ([]ItemRead, error) {
	var items []model.Item
	err := db.NewSelect().Model(&items).Limit(limit).Offset(offset).OrderExpr("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ItemRead, 0, len(items))
	for _, item := range items {
		result = append(result, toItemRead(item))
	}
	return result, nil
}

// ListItemsWithCount returns items and total count for pagination metadata.
// The count query uses the same base query (with same WHERE filters) as the list query.
func ListItemsWithCount(ctx context.Context, limit, offset int64) ([]ItemRead, int64, error) {
	var items []model.Item
	q := db.NewSelect().Model(&items)

	// Get total count from the same base query
	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering to the same query, then scan
	err = q.Limit(limit).Offset(offset).OrderExpr("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]ItemRead, 0, len(items))
	for _, item := range items {
		result = append(result, toItemRead(item))
	}
	return result, int64(total), nil
}

// CreateItem creates a new item.
func CreateItem(ctx context.Context, name, description string) (ItemRead, error) {
	if name == "" {
		return ItemRead{}, ErrInvalid
	}

	item := model.Item{
		BaseModel: model.BaseModel{
			ID: uuid.New(),
		},
		Name:        name,
		Description: description,
	}

	if _, err := db.NewInsert().Model(&item).Exec(ctx); err != nil {
		return ItemRead{}, err
	}

	return toItemRead(item), nil
}

// UpdateItem updates an existing item.
func UpdateItem(ctx context.Context, id uuid.UUID, name, description *string) (ItemRead, error) {
	item := &model.Item{}
	err := db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ItemRead{}, ErrNotFound
		}
		return ItemRead{}, err
	}

	if name != nil {
		item.Name = *name
	}
	if description != nil {
		item.Description = *description
	}

	if _, err := db.NewUpdate().Model(item).Where("id = ?", item.ID).Exec(ctx); err != nil {
		return ItemRead{}, err
	}

	return toItemRead(*item), nil
}

// DeleteItem soft-deletes an item (sets DeletedAt, doesn't remove the row).
// Bun's soft_delete tag on DeletedAt converts this to an UPDATE automatically.
func DeleteItem(ctx context.Context, id uuid.UUID) error {
	// Check if item exists first
	item := &model.Item{}
	err := db.NewSelect().Model(item).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	// Soft delete — bun's soft_delete tag converts this to UPDATE SET deleted_at=now
	_, err = db.NewDelete().Model(item).Where("id = ?", id).Exec(ctx)
	return err
}
