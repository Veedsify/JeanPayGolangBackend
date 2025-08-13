package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/hibiken/asynq"
)

const (
	TypeActivityLog = "activity:log"
)

type ActivityJobPayload struct {
	UserID   uint   `json:"user_id"`
	Activity string `json:"activity"`
}

type ActivityJobClient struct {
	client *asynq.Client
}

func NewActivityJobClient() *ActivityJobClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &ActivityJobClient{
		client: client,
	}
}

// Close closes the email job client
func (ejc *ActivityJobClient) Close() error {
	return ejc.client.Close()
}

// Enqueue New Activity

func (ejc *ActivityJobClient) EnqueueNewActivity(userid uint, activity string) error {
	payload := ActivityJobPayload{
		UserID:   userid,
		Activity: activity,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeActivityLog, payloadBytes)

	opts := []asynq.Option{
		asynq.Queue("low"),
		asynq.MaxRetry(3),
		asynq.Timeout(5 * time.Minute),
	}

	info, err := ejc.client.Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue activity task: %w", err)
	}

	log.Printf("Enqueued activity task: id=%s queue=%s", info.ID, info.Queue)
	return nil
}

func HandleActivityJobTask(ctx context.Context, t *asynq.Task) error {
	var payload ActivityJobPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal activity")
	}
	createActivity := models.Activity{
		UserID:   payload.UserID,
		Activity: payload.Activity,
	}

	if err := database.DB.Create(&createActivity).Error; err != nil {
		return errors.New("sorry this account already exists")
	}

	return nil
}
