package models

import (
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid/v2"
	"time"
)

const (
	EventStatusPending    = "pending"
	EventStatusProcessing = "processing"
	EventStatusCompleted  = "completed"
	EventStatusFailed     = "failed"
)

const EventProducer = "hyperdrift.storage"

func NewEventId() string {
	return fmt.Sprintf(`event_%s`, ulid.Make().String())
}

type Event[T any] struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Content   T          `json:"content"`
	Status    string     `json:"status"`
	Retries   int        `json:"retries"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewEvent[T any](name string, content T) *Event[T] {
	return &Event[T]{
		Id:        NewEventId(),
		Name:      name,
		Content:   content,
		Status:    EventStatusPending,
		Retries:   0,
		ExpiresAt: nil,
		CreatedAt: time.Now(),
	}
}

func (e *Event[T]) ContentToByte() (error, []byte) {
	contentByte, err := json.Marshal(e.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal '%s' event content error: %w", e.Name, err), nil
	}

	return nil, contentByte
}
