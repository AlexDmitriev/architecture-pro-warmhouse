package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"telemetry/models"

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

// GetSensors retrieves all sensors from the database
func (db *DB) GetTelemetry(ctx context.Context) ([]models.Sensor, error) {
	query := `
		SELECT id, serial_number, value, created_at
		FROM telemetry
		ORDER BY created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying telemetry: %w", err)
	}
	defer rows.Close()

	var telemetry []models.Telemetry
	for rows.Next() {
		var t models.Telemetry
		err := rows.Scan(
			&t.ID,
			&t.SerialNumber,
			&t.Value,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning telemetry row: %w", err)
		}
		sensors = append(sensors, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	return telemetry, nil
}

// GetSensorByID retrieves a sensor by its ID
func (db *DB) GetTelemetryBySerialNumber(ctx context.Context, serial string) (models.Sensor, error) {
	query := `
    		SELECT id, serial_number, value, created_at
    		FROM telemetry
    		WHERE id = $1
    		ORDER BY created_at DESC
    	`

	var t models.Telemetry
	err := db.Pool.QueryRow(ctx, query, serial_number).Scan(
		&t.ID,
        &t.SerialNumber,
        &t.Value,
        &t.CreatedAt,
	)
	if err != nil {
		return models.Telemetry{}, fmt.Errorf("error getting telemetry by SerialNumber: %w", err)
	}

	return s, nil
}

// CreateSensor creates a new sensor in the database
func (db *DB) CreateTelemetry(ctx context.Context, s models.TelemetryCreate) (models.Telemetry, error) {
	query := `
		INSERT INTO sensors (name, type, location, unit, status, last_updated, created_at)
		VALUES ($1, $2, $3, $4, 'inactive', $5, $5)
		RETURNING id, name, type, location, value, unit, status, last_updated, created_at
	`

	now := time.Now()
	var sensor models.Sensor
	err := db.Pool.QueryRow(ctx, query,
		s.Name,
		s.Type,
		s.Location,
		s.Unit,
		now,
	).Scan(
		&sensor.ID,
		&sensor.Name,
		&sensor.Type,
		&sensor.Location,
		&sensor.Value,
		&sensor.Unit,
		&sensor.Status,
		&sensor.LastUpdated,
		&sensor.CreatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error creating sensor: %w", err)
	}

	return sensor, nil
}
