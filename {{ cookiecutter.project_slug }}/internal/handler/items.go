package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"{{ cookiecutter.import_path }}/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// ListResponse wraps a paginated list with metadata.
type ListResponse struct {
	Items  []service.ItemRead `json:"items" doc:"List of items"`
	Total  int64              `json:"total" doc:"Total number of items"`
	Limit  int64              `json:"limit" doc:"Items per page"`
	Offset int64              `json:"offset" doc:"Current offset"`
}

// Setup registers all API routes.
func Setup(api huma.API, logger *slog.Logger) {
	// Auth routes (login to get JWT token)
	SetupAuth(api, logger)

	// GET /api/v1/items — List items
	huma.Get(api, "/api/v1/items", func(ctx context.Context, input *struct {
		Limit  int `query:"limit" minimum:"1" maximum:"100" default:"20" doc:"Max items to return"`
		Offset int `query:"offset" minimum:"0" default:"0" doc:"Offset for pagination"`
	}) (*struct {
		Body ListResponse
	}, error) {
		items, total, err := service.ListItemsWithCount(ctx, int64(input.Limit), int64(input.Offset))
		if err != nil {
			logger.ErrorContext(ctx, "list items failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct{ Body ListResponse }{
			Body: ListResponse{
				Items:  items,
				Total:  total,
				Limit:  int64(input.Limit),
				Offset: int64(input.Offset),
			},
		}, nil
	})

	// POST /api/v1/items — Create item
	huma.Post(api, "/api/v1/items", func(ctx context.Context, input *struct {
		Body struct {
			Name        string `json:"name" maxLength:"255" doc:"Item name"`
			Description string `json:"description" maxLength:"10000" doc:"Item description"`
		}
	}) (*struct {
		Status int
		Body   service.ItemRead
	}, error) {
		item, err := service.CreateItem(ctx, input.Body.Name, input.Body.Description)
		if err != nil {
			if errors.Is(err, service.ErrInvalid) {
				return nil, huma.NewError(http.StatusBadRequest, "invalid input")
			}
			logger.ErrorContext(ctx, "create item failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct {
			Status int
			Body   service.ItemRead
		}{Status: http.StatusCreated, Body: item}, nil
	})

	// GET /api/v1/items/{id} — Get item
	huma.Get(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
		ID uuid.UUID `path:"id" doc:"Item ID"`
	}) (*struct {
		Body service.ItemRead
	}, error) {
		item, err := service.GetItem(ctx, input.ID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				return nil, huma.NewError(http.StatusNotFound, "item not found")
			}
			logger.ErrorContext(ctx, "get item failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct{ Body service.ItemRead }{Body: item}, nil
	})

	// PUT /api/v1/items/{id} — Update item
	huma.Put(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
		ID   uuid.UUID `path:"id" doc:"Item ID"`
		Body struct {
			Name        *string `json:"name" maxLength:"255" doc:"Item name"`
			Description *string `json:"description" maxLength:"10000" doc:"Item description"`
		}
	}) (*struct {
		Body service.ItemRead
	}, error) {
		item, err := service.UpdateItem(ctx, input.ID, input.Body.Name, input.Body.Description)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				return nil, huma.NewError(http.StatusNotFound, "item not found")
			}
			logger.ErrorContext(ctx, "update item failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct{ Body service.ItemRead }{Body: item}, nil
	})

	// DELETE /api/v1/items/{id} — Delete item
	huma.Delete(api, "/api/v1/items/{id}", func(ctx context.Context, input *struct {
		ID uuid.UUID `path:"id" doc:"Item ID"`
	}) (*struct {
		Status int
	}, error) {
		if err := service.DeleteItem(ctx, input.ID); err != nil {
			if errors.Is(err, service.ErrNotFound) {
				return nil, huma.NewError(http.StatusNotFound, "item not found")
			}
			logger.ErrorContext(ctx, "delete item failed", "error", err)
			return nil, huma.NewError(http.StatusInternalServerError, "internal server error")
		}
		return &struct{ Status int }{Status: http.StatusNoContent}, nil
	})
}
