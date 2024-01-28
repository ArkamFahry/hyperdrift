package models

import (
	"time"
)

const EventProducer = "hyperdrift-storage"

type Event struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Payload   any       `json:"payload"`
	Status    string    `json:"status"`
	Producer  string    `json:"producer"`
	Timestamp time.Time `json:"timestamp"`
}

type EventCreate struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Payload   any       `json:"payload"`
	Status    string    `json:"status"`
	Producer  string    `json:"producer"`
	Timestamp time.Time `json:"timestamp"`
}

func NewEventCreate(id string, name string, payload any) *EventCreate {
	return &EventCreate{
		Id:        id,
		Name:      name,
		Payload:   payload,
		Status:    "pending",
		Producer:  EventProducer,
		Timestamp: time.Now(),
	}
}
