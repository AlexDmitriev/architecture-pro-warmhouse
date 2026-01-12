package models

import (
	"time"
)

type Telemetry struct {
	ID           string    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Value        string    `json:"type"`
	CreatedAt   time.Time  `json:"created_at"`
}

type TelemetryCreate struct {
	SerialNumber string    `json:"serial_number"`
    Value        string    `json:"type"`
}
