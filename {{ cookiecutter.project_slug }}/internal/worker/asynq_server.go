package worker

import (
	"log/slog"

	"github.com/hibiken/asynq"
)

// NewServer creates a new Asynq server.
func NewServer(redisConn asynq.RedisConnOpt, logger *slog.Logger) *asynq.Server {
	srv := asynq.NewServer(
		redisConn,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
	logger.Info("asynq server configured")
	return srv
}

// NewMux creates the Asynq serve mux with all task handlers registered.
func NewMux(logger *slog.Logger) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Handle(TaskTypeDataProcess, DataProcessHandler(logger))
	return mux
}
