package models

import (
	"time"
)

type Heating struct {
	ID           string    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Temperature  float64   `json:"temperature"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type HeatingUpdate struct {
	SerialNumber string  `json:"serial_number" binding:"required"`
	Temperature  float64 `json:"temperature" binding:"required"`
}
