package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chandra0013/meddevice-inventory-api/internal/store"
)

func TestAuthRequired(t *testing.T) {
	server := NewServer(store.NewMemoryStore(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/devices", nil)
	res := httptest.NewRecorder()

	server.Routes().ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestDeviceCRUD(t *testing.T) {
	server := NewServer(store.NewMemoryStore(), "secret")

	device := store.DeviceInput{
		Name:         "Cardiac Monitor",
		Manufacturer: "Acme MedTech",
		SerialNumber: "CM-100",
		Status:       "in_service",
		Location:     "ICU",
	}
	body, _ := json.Marshal(device)
	create := authed(http.MethodPost, "/devices", bytes.NewReader(body))
	createRes := httptest.NewRecorder()
	server.Routes().ServeHTTP(createRes, create)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	var created store.Device
	if err := json.NewDecoder(createRes.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() {
		t.Fatalf("expected persisted device metadata, got %+v", created)
	}

	get := authed(http.MethodGet, "/devices/1", nil)
	getRes := httptest.NewRecorder()
	server.Routes().ServeHTTP(getRes, get)
	if getRes.Code != http.StatusOK {
		t.Fatalf("get status = %d", getRes.Code)
	}

	updateInput := device
	updateInput.Status = "maintenance"
	updateBody, _ := json.Marshal(updateInput)
	update := authed(http.MethodPut, "/devices/1", bytes.NewReader(updateBody))
	updateRes := httptest.NewRecorder()
	server.Routes().ServeHTTP(updateRes, update)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update status = %d", updateRes.Code)
	}

	deleteReq := authed(http.MethodDelete, "/devices/1", nil)
	deleteRes := httptest.NewRecorder()
	server.Routes().ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", deleteRes.Code)
	}
}

func authed(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Authorization", "Bearer secret")
	return req
}
