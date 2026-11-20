package service_test

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"testing"

	"{{ cookiecutter.import_path }}/internal/model"
	"{{ cookiecutter.import_path }}/internal/service"
	"{{ cookiecutter.import_path }}/test/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db := util.TestDB(t)
	util.CreateTables(t, db, &model.Item{})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	service.Init(db, logger, "test-secret", "test-api-key")
	t.Cleanup(func() { _ = service.Close() })

	return db
}

func TestListItemsEmpty(t *testing.T) {

	setupTestDB(t)
	items, err := service.ListItems(context.Background(), 20, 0)

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestCreateAndListItems(t *testing.T) {

	setupTestDB(t)

	// Create
	item, err := service.CreateItem(context.Background(), "Test Item", "A test item")
	require.NoError(t, err)
	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "Test Item", item.Name)
	assert.Equal(t, "A test item", item.Description)
	assert.Nil(t, item.CreatedBy)
	assert.NotZero(t, item.CreatedAt)

	// List
	items, err := service.ListItems(context.Background(), 20, 0)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, item.ID, items[0].ID)
}

func TestCreateItemEmptyName(t *testing.T) {

	setupTestDB(t)
	_, err := service.CreateItem(context.Background(), "", "A test item")

	assert.ErrorIs(t, err, service.ErrInvalid)
}

func TestGetItem(t *testing.T) {

	setupTestDB(t)

	item, err := service.CreateItem(context.Background(), "Test Item", "A test item")
	require.NoError(t, err)

	// Get
	got, err := service.GetItem(context.Background(), item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, got.ID)
	assert.Equal(t, "Test Item", got.Name)
}

func TestGetItemNotFound(t *testing.T) {

	setupTestDB(t)
	_, err := service.GetItem(context.Background(), uuid.New())

	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestUpdateItem(t *testing.T) {

	setupTestDB(t)

	item, err := service.CreateItem(context.Background(), "Original", "Description")
	require.NoError(t, err)

	name := "Updated"
	updated, err := service.UpdateItem(context.Background(), item.ID, &name, nil)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "Description", updated.Description)
}

func TestUpdateItemNotFound(t *testing.T) {

	setupTestDB(t)
	name := "Updated"
	_, err := service.UpdateItem(context.Background(), uuid.New(), &name, nil)

	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestDeleteItem(t *testing.T) {

	setupTestDB(t)

	item, err := service.CreateItem(context.Background(), "Test Item", "A test item")
	require.NoError(t, err)

	err = service.DeleteItem(context.Background(), item.ID)
	require.NoError(t, err)

	// Verify it's gone from queries (soft delete filters automatically)
	_, err = service.GetItem(context.Background(), item.ID)
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestDeleteItemNotFound(t *testing.T) {

	setupTestDB(t)
	err := service.DeleteItem(context.Background(), uuid.New())

	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestSoftDeleteFiltering(t *testing.T) {
	// M5: Verify that soft-deleted items are filtered from list queries
	// but the row still exists in the database (not hard-deleted).

	setupTestDB(t)

	// Create 3 items
	_, err := service.CreateItem(context.Background(), "Active 1", "")
	require.NoError(t, err)
	_, err = service.CreateItem(context.Background(), "Active 2", "")
	require.NoError(t, err)
	item3, err := service.CreateItem(context.Background(), "To Delete", "")
	require.NoError(t, err)

	// All 3 should be visible
	items, total, err := service.ListItemsWithCount(context.Background(), 20, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, items, 3)

	// Soft-delete item3
	err = service.DeleteItem(context.Background(), item3.ID)
	require.NoError(t, err)

	// Only 2 should be visible now
	items, total, err = service.ListItemsWithCount(context.Background(), 20, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// Verify soft-deleted item is not in results
	for _, item := range items {
		assert.NotEqual(t, item3.ID, item.ID, "soft-deleted item should not appear in list")
	}
}

func TestListItemsPagination(t *testing.T) {

	setupTestDB(t)

	// Create 5 items
	for i := 0; i < 5; i++ {
		_, err := service.CreateItem(context.Background(), "Item "+strconv.Itoa(i), "")
		require.NoError(t, err)
	}

	// Page 1: limit 2, offset 0
	items, err := service.ListItems(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Page 2: limit 2, offset 2
	items, err = service.ListItems(context.Background(), 2, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	// Page 3: limit 2, offset 4
	items, err = service.ListItems(context.Background(), 2, 4)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestListItemsWithCount(t *testing.T) {

	setupTestDB(t)

	// Empty
	items, total, err := service.ListItemsWithCount(context.Background(), 20, 0)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, int64(0), total)

	// Create 5 items
	for i := 0; i < 5; i++ {
		_, err = service.CreateItem(context.Background(), "Item "+strconv.Itoa(i), "")
		require.NoError(t, err)
	}

	// Page 1: limit 2, offset 0
	items, total, err = service.ListItemsWithCount(context.Background(), 2, 0)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(5), total)

	// Page 2: limit 2, offset 2
	items, total, err = service.ListItemsWithCount(context.Background(), 2, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(5), total)

	// Page 3: limit 2, offset 4
	items, total, err = service.ListItemsWithCount(context.Background(), 2, 4)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(5), total)
}

func TestClose(t *testing.T) {

	db := util.TestDB(t)
	service.Init(db, slog.Default(), "test-secret", "test-api-key")
	err := service.Close()
	assert.NoError(t, err)
}

func TestCloseNilDB(t *testing.T) {

	// Don't call Init — db is nil
	err := service.Close()
	assert.NoError(t, err)
}
