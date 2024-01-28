package models

import (
	"encoding/json"
	"time"
)

type Event struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	Producer  string          `json:"producer"`
	Timestamp time.Time       `json:"timestamp"`
}
