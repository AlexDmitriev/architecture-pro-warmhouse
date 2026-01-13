package models

import "time"

// Sensor представляет запись из таблицы sensors.
type Sensor struct {
	ID             string     `json:"id"`
	LocationID     *string    `json:"location_id,omitempty"`
	Name           string     `json:"name"`
	SerialNumber   string     `json:"serial_number"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	Type           string     `json:"type,omitempty"`
}

// SensorCreate входная модель для создания сенсора.
type SensorCreate struct {
	LocationID     *string    `json:"location_id,omitempty"`
	Name           string     `json:"name" binding:"required"`
	SerialNumber   string     `json:"serial_number" binding:"required"`
	Status         string     `json:"status" binding:"required"`
	Type           string     `json:"type" binding:"required"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

// SensorUpdate входная модель для обновления сенсора.
type SensorUpdate struct {
	LocationID     *string    `json:"location_id,omitempty"`
	Name           *string    `json:"name,omitempty"`
	Status         *string    `json:"status,omitempty"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	Type           *string    `json:"type,omitempty"`
}

