package store

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("device not found")

type Device struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Manufacturer string    `json:"manufacturer"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	Location     string    `json:"location"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DeviceInput struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	SerialNumber string `json:"serial_number"`
	Status       string `json:"status"`
	Location     string `json:"location"`
}

func (d DeviceInput) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(d.Manufacturer) == "" {
		return errors.New("manufacturer is required")
	}
	if strings.TrimSpace(d.SerialNumber) == "" {
		return errors.New("serial_number is required")
	}
	switch d.Status {
	case "in_service", "maintenance", "retired":
		return nil
	default:
		return errors.New("status must be one of: in_service, maintenance, retired")
	}
}

type DeviceStore interface {
	List(ctx context.Context, limit, offset int) ([]Device, error)
	Create(ctx context.Context, input DeviceInput) (Device, error)
	Get(ctx context.Context, id int64) (Device, error)
	Update(ctx context.Context, id int64, input DeviceInput) (Device, error)
	Delete(ctx context.Context, id int64) error
}
