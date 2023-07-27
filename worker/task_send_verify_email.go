package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to convert payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enque task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", jsonPayload).Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqued task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshall payload: %w", asynq.SkipRetry)
	}

	_, err = processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with such username does not exsist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: send email to user
	log.Info().Msg("processed task")
	
	return nil
}
