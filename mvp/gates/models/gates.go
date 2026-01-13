package models

import (
	"time"
)

type Gates struct {
	ID           string    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GatesUpdate struct {
	SerialNumber string `json:"serial_number" binding:"required"`
	Status       string `json:"status" binding:"required,oneof=open close"`
}
