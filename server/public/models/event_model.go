package models

import (
	"fmt"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
	"github.com/ArkamFahry/hyperdrift-storage/server/packages/utils"
	"github.com/ArkamFahry/hyperdrift-storage/server/packages/validators"
)

type CreateEvent struct {
	Id        string         `json:"id"`
	Name      string         `json:"name"`
	Data      map[string]any `json:"data"`
	Producer  string         `json:"producer"`
	CreatedAt time.Time      `json:"created_at"`
}

func NewEventId() string {
	return fmt.Sprintf(`%s_%s`, "event", utils.NewId())
}

func NewEvent(name string, data map[string]any) *CreateEvent {
	return &CreateEvent{
		Id:        NewEventId(),
		Name:      name,
		Data:      data,
		Producer:  "heyperdrift.storage",
		CreatedAt: time.Now(),
	}
}

func (e *CreateEvent) Validate() error {
	if validators.IsEmptyString(e.Id) {
		return apperr.NewFieldError("id", "id is required")
	}

	if validators.IsEmptyString(e.Name) {
		return apperr.NewFieldError("name", "name is required")
	}

	if validators.ContainsAnyWhiteSpaces(e.Name) {
		return apperr.NewFieldError("name", "name should not contain any white spaces or tabs")
	}

	if e.Data == nil {
		return apperr.NewFieldError("data", "data is required")
	}

	if validators.IsEmptyString(e.Producer) {
		return apperr.NewFieldError("producer", "producer is required")
	}

	if validators.ContainsAnyWhiteSpaces(e.Producer) {
		return apperr.NewFieldError("producer", "producer should not contain any white spaces or tabs")
	}

	if e.CreatedAt.IsZero() {
		return apperr.NewFieldError("created_at", "created_at is required")
	}

	return nil
}
