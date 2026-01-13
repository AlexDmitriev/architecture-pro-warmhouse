package db

import (
	"context"
	"fmt"

	"gates/models"

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

// UpdateGates обновляет или создает запись статуса для датчика ворот
func (db *DB) UpdateGates(ctx context.Context, g models.GatesUpdate) (models.Gates, error) {
	query := `
		INSERT INTO sensor_gates (serial_number, status, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (serial_number) 
		DO UPDATE SET 
			status = EXCLUDED.status,
			updated_at = NOW()
		RETURNING id, serial_number, status, created_at, updated_at
	`

	var gates models.Gates
	err := db.Pool.QueryRow(ctx, query,
		g.SerialNumber,
		g.Status,
	).Scan(
		&gates.ID,
		&gates.SerialNumber,
		&gates.Status,
		&gates.CreatedAt,
		&gates.UpdatedAt,
	)
	if err != nil {
		return models.Gates{}, fmt.Errorf("error updating gates: %w", err)
	}

	return gates, nil
}

// GetGates получает запись по serial_number
func (db *DB) GetGates(ctx context.Context, serialNumber string) (models.Gates, error) {
	query := `
		SELECT id, serial_number, status, created_at, updated_at
		FROM sensor_gates
		WHERE serial_number = $1
	`

	var gates models.Gates
	err := db.Pool.QueryRow(ctx, query, serialNumber).Scan(
		&gates.ID,
		&gates.SerialNumber,
		&gates.Status,
		&gates.CreatedAt,
		&gates.UpdatedAt,
	)
	if err != nil {
		return models.Gates{}, fmt.Errorf("error getting gates: %w", err)
	}

	return gates, nil
}
