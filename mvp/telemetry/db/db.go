package db

import (
	"context"
	"fmt"

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

// GetTelemetry retrieves all telemetry records from the database
// If serialNumber is provided, filters by serial_number
func (db *DB) GetTelemetry(ctx context.Context, serialNumber *string) ([]models.Telemetry, error) {
	var query string
	var args []interface{}

	if serialNumber != nil && *serialNumber != "" {
		query = `
			SELECT id, serial_number, value, created_at
			FROM telemetry
			WHERE serial_number = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{*serialNumber}
	} else {
		query = `
			SELECT id, serial_number, value, created_at
			FROM telemetry
			ORDER BY created_at DESC
		`
		args = nil
	}

	rows, err := db.Pool.Query(ctx, query, args...)
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
		telemetry = append(telemetry, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	return telemetry, nil
}

// GetTelemetryBySerialNumber retrieves telemetry records by serial number
func (db *DB) GetTelemetryBySerialNumber(ctx context.Context, serial string) ([]models.Telemetry, error) {
	return db.GetTelemetry(ctx, &serial)
}

// CreateTelemetry creates a new telemetry record in the database
func (db *DB) CreateTelemetry(ctx context.Context, t models.TelemetryCreate) (models.Telemetry, error) {
	query := `
		INSERT INTO telemetry (serial_number, value, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, serial_number, value, created_at
	`

	var telemetry models.Telemetry
	err := db.Pool.QueryRow(ctx, query,
		t.SerialNumber,
		t.Value,
	).Scan(
		&telemetry.ID,
		&telemetry.SerialNumber,
		&telemetry.Value,
		&telemetry.CreatedAt,
	)
	if err != nil {
		return models.Telemetry{}, fmt.Errorf("error creating telemetry: %w", err)
	}

	return telemetry, nil
}
