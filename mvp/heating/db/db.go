package db

import (
	"context"
	"fmt"

	"heating/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new DB instance
func New(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// UpdateHeating обновляет или создает запись температуры для датчика отопления
func (db *DB) UpdateHeating(ctx context.Context, h models.HeatingUpdate) (models.Heating, error) {
	query := `
		INSERT INTO sensor_heating (serial_number, temperature, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (serial_number) 
		DO UPDATE SET 
			temperature = EXCLUDED.temperature,
			updated_at = NOW()
		RETURNING id, serial_number, temperature, created_at, updated_at
	`

	var heating models.Heating
	err := db.Pool.QueryRow(ctx, query,
		h.SerialNumber,
		h.Temperature,
	).Scan(
		&heating.ID,
		&heating.SerialNumber,
		&heating.Temperature,
		&heating.CreatedAt,
		&heating.UpdatedAt,
	)
	if err != nil {
		return models.Heating{}, fmt.Errorf("error updating heating: %w", err)
	}

	return heating, nil
}

// GetHeating получает запись по serial_number
func (db *DB) GetHeating(ctx context.Context, serialNumber string) (models.Heating, error) {
	query := `
		SELECT id, serial_number, temperature, created_at, updated_at
		FROM sensor_heating
		WHERE serial_number = $1
	`

	var heating models.Heating
	err := db.Pool.QueryRow(ctx, query, serialNumber).Scan(
		&heating.ID,
		&heating.SerialNumber,
		&heating.Temperature,
		&heating.CreatedAt,
		&heating.UpdatedAt,
	)
	if err != nil {
		return models.Heating{}, fmt.Errorf("error getting heating: %w", err)
	}

	return heating, nil
}
