package store

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.Mutex
	nextID  int64
	devices map[int64]Device
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{nextID: 1, devices: map[int64]Device{}}
}

func (s *MemoryStore) List(ctx context.Context, limit, offset int) ([]Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	devices := make([]Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	if offset >= len(devices) {
		return []Device{}, nil
	}
	end := offset + limit
	if end > len(devices) {
		end = len(devices)
	}
	return devices[offset:end], nil
}

func (s *MemoryStore) Create(ctx context.Context, input DeviceInput) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	device := Device{
		ID:           s.nextID,
		Name:         input.Name,
		Manufacturer: input.Manufacturer,
		SerialNumber: input.SerialNumber,
		Status:       input.Status,
		Location:     input.Location,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.devices[device.ID] = device
	s.nextID++
	return device, nil
}

func (s *MemoryStore) Get(ctx context.Context, id int64) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return Device{}, ErrNotFound
	}
	return device, nil
}

func (s *MemoryStore) Update(ctx context.Context, id int64, input DeviceInput) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[id]
	if !ok {
		return Device{}, ErrNotFound
	}
	device.Name = input.Name
	device.Manufacturer = input.Manufacturer
	device.SerialNumber = input.SerialNumber
	device.Status = input.Status
	device.Location = input.Location
	device.UpdatedAt = time.Now().UTC()
	s.devices[id] = device
	return device, nil
}

func (s *MemoryStore) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.devices[id]; !ok {
		return ErrNotFound
	}
	delete(s.devices, id)
	return nil
}
