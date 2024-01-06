package entities

import (
	"time"
)

type Event struct {
	Id        string         `json:"id"`
	Name      string         `json:"name"`
	Data      map[string]any `json:"data"`
	Producer  string         `json:"producer"`
	CreatedAt time.Time      `json:"created_at"`
}
