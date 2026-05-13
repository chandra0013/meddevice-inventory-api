package store

import (
	"context"
	"database/sql"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			manufacturer TEXT NOT NULL,
			serial_number TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL CHECK (status IN ('in_service', 'maintenance', 'retired')),
			location TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
		CREATE INDEX IF NOT EXISTS idx_devices_manufacturer ON devices(manufacturer);
	`)
	return err
}

func (s *PostgresStore) List(ctx context.Context, limit, offset int) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, manufacturer, serial_number, status, location, created_at, updated_at
		FROM devices
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]Device, 0)
	for rows.Next() {
		device, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (s *PostgresStore) Create(ctx context.Context, input DeviceInput) (Device, error) {
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO devices (name, manufacturer, serial_number, status, location)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, manufacturer, serial_number, status, location, created_at, updated_at
	`, input.Name, input.Manufacturer, input.SerialNumber, input.Status, input.Location)
	return scanDevice(row)
}

func (s *PostgresStore) Get(ctx context.Context, id int64) (Device, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, manufacturer, serial_number, status, location, created_at, updated_at
		FROM devices
		WHERE id = $1
	`, id)
	return scanDevice(row)
}

func (s *PostgresStore) Update(ctx context.Context, id int64, input DeviceInput) (Device, error) {
	row := s.db.QueryRowContext(ctx, `
		UPDATE devices
		SET name = $1, manufacturer = $2, serial_number = $3, status = $4, location = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, manufacturer, serial_number, status, location, created_at, updated_at
	`, input.Name, input.Manufacturer, input.SerialNumber, input.Status, input.Location, id)
	return scanDevice(row)
}

func (s *PostgresStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDevice(row scanner) (Device, error) {
	var device Device
	err := row.Scan(
		&device.ID,
		&device.Name,
		&device.Manufacturer,
		&device.SerialNumber,
		&device.Status,
		&device.Location,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Device{}, ErrNotFound
	}
	return device, err
}
