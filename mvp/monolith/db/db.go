package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"monolith/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB обертка над пулом соединений Postgres.
type DB struct {
	Pool *pgxpool.Pool
}

// New создает подключение к БД и выполняет ping.
func New(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close закрывает пул.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// CreateSensor добавляет запись в таблицу sensors.
func (db *DB) CreateSensor(ctx context.Context, input models.SensorCreate) (models.Sensor, error) {
	query := `
		INSERT INTO sensors (location_id, name, serial_number, sensor_type_code, status, last_activity_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, location_id, name, serial_number, sensor_type_code, status, created_at, last_activity_at
	`

	var sensor models.Sensor
	err := db.Pool.QueryRow(ctx, query,
		input.LocationID,
		input.Name,
		input.SerialNumber,
		input.Type,
		input.Status,
		input.LastActivityAt,
	).Scan(
		&sensor.ID,
		&sensor.LocationID,
		&sensor.Name,
		&sensor.SerialNumber,
		&sensor.Type,
		&sensor.Status,
		&sensor.CreatedAt,
		&sensor.LastActivityAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error creating sensor: %w", err)
	}

	// Тип не хранится в таблице, возвращаем из запроса.
	sensor.Type = input.Type
	return sensor, nil
}

// ListSensors возвращает все сенсоры.
func (db *DB) ListSensors(ctx context.Context) ([]models.Sensor, error) {
	query := `
		SELECT id, location_id, name, serial_number, sensor_type_code, status, created_at, last_activity_at
		FROM sensors
		ORDER BY created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors: %w", err)
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var s models.Sensor
		var loc sql.NullString
		if err := rows.Scan(
			&s.ID,
			&loc,
			&s.Name,
			&s.SerialNumber,
			&s.Type,
			&s.Status,
			&s.CreatedAt,
			&s.LastActivityAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		if loc.Valid {
			s.LocationID = &loc.String
		}
		sensors = append(sensors, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sensors rows: %w", err)
	}

	return sensors, nil
}

// GetSensor возвращает сенсор по id.
func (db *DB) GetSensor(ctx context.Context, id string) (models.Sensor, error) {
	query := `
		SELECT id, location_id, name, serial_number, sensor_type_code, status, created_at, last_activity_at
		FROM sensors
		WHERE id = $1
	`

	var sensor models.Sensor
	var loc sql.NullString
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&sensor.ID,
		&loc,
		&sensor.Name,
		&sensor.SerialNumber,
		&sensor.Type,
		&sensor.Status,
		&sensor.CreatedAt,
		&sensor.LastActivityAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error getting sensor: %w", err)
	}
	if loc.Valid {
		sensor.LocationID = &loc.String
	}

	return sensor, nil
}

// UpdateSensor обновляет указанные поля сенсора.
func (db *DB) UpdateSensor(ctx context.Context, id string, input models.SensorUpdate) (models.Sensor, error) {
	setParts := make([]string, 0, 4)
	args := make([]interface{}, 0, 5)
	argIdx := 1

	if input.LocationID != nil {
		setParts = append(setParts, fmt.Sprintf("location_id = $%d", argIdx))
		args = append(args, *input.LocationID)
		argIdx++
	}
	if input.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Type != nil {
		setParts = append(setParts, fmt.Sprintf("sensor_type_code = $%d", argIdx))
		args = append(args, *input.Type)
		argIdx++
	}
	if input.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.LastActivityAt != nil {
		setParts = append(setParts, fmt.Sprintf("last_activity_at = $%d", argIdx))
		args = append(args, *input.LastActivityAt)
		argIdx++
	}

	if len(setParts) == 0 {
		return db.GetSensor(ctx, id)
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE sensors
		SET %s
		WHERE id = $%d
		RETURNING id, location_id, name, serial_number, sensor_type_code, status, created_at, last_activity_at
	`, strings.Join(setParts, ", "), argIdx)

	var sensor models.Sensor
	var loc sql.NullString
	err := db.Pool.QueryRow(ctx, query, args...).Scan(
		&sensor.ID,
		&loc,
		&sensor.Name,
		&sensor.SerialNumber,
		&sensor.Type,
		&sensor.Status,
		&sensor.CreatedAt,
		&sensor.LastActivityAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error updating sensor: %w", err)
	}
	if loc.Valid {
		sensor.LocationID = &loc.String
	}

	// Тип не хранится в таблице, возвращаем как есть (если был в запросе).
	if input.Type != nil {
		sensor.Type = *input.Type
	}

	return sensor, nil
}

// DeleteSensor удаляет запись по id.
func (db *DB) DeleteSensor(ctx context.Context, id string) error {
	cmdTag, err := db.Pool.Exec(ctx, "DELETE FROM sensors WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting sensor: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("sensor not found")
	}
	return nil
}

