package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Task type constants.
const (
	TaskTypeDataProcess = "data:process"
)

// DataProcessPayload is the payload for the data processing task.
type DataProcessPayload struct {
	ItemID string `json:"item_id"`
}

// NewDataProcessTask creates an Asynq task for data processing.
func NewDataProcessTask(itemID string) (*asynq.Task, error) {
	payload, err := json.Marshal(DataProcessPayload{ItemID: itemID})
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeDataProcess, payload,
		asynq.MaxRetry(3),
		asynq.Queue("default"),
	), nil
}

// DataProcessHandler handles the data processing task.
func DataProcessHandler(logger *slog.Logger) asynq.Handler {
	return &dataProcessHandler{logger: logger}
}

type dataProcessHandler struct {
	logger *slog.Logger
}

func (h *dataProcessHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload DataProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Validate item_id is a valid UUID before processing
	if _, err := uuid.Parse(payload.ItemID); err != nil {
		return fmt.Errorf("invalid item_id format: %w", err)
	}

	h.logger.Info("processing data", slog.String("item_id", payload.ItemID))
	// Processing logic here — the task is retried automatically on error
	return nil
}
