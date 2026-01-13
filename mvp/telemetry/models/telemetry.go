package models

import (
	"encoding/json"
	"time"
)

type Telemetry struct {
	ID           string          `json:"id"`
	SerialNumber string          `json:"serial_number"`
	Value        json.RawMessage `json:"value"`
	CreatedAt    time.Time       `json:"created_at"`
}

type TelemetryCreate struct {
	SerialNumber string          `json:"serial_number"`
	Value        json.RawMessage `json:"value"`
}
