package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/chandra0013/meddevice-inventory-api/internal/store"
)

type Server struct {
	store  store.DeviceStore
	apiKey string
}

func NewServer(deviceStore store.DeviceStore, apiKey string) *Server {
	return &Server{store: deviceStore, apiKey: apiKey}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.Handle("GET /devices", s.auth(http.HandlerFunc(s.listDevices)))
	mux.Handle("POST /devices", s.auth(http.HandlerFunc(s.createDevice)))
	mux.Handle("GET /devices/{id}", s.auth(http.HandlerFunc(s.getDevice)))
	mux.Handle("PUT /devices/{id}", s.auth(http.HandlerFunc(s.updateDevice)))
	mux.Handle("DELETE /devices/{id}", s.auth(http.HandlerFunc(s.deleteDevice)))
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)
	offset := parsePositiveInt(r.URL.Query().Get("offset"), 0)
	if limit > 100 {
		limit = 100
	}

	devices, err := s.store.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list devices")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"devices": devices, "limit": limit, "offset": offset})
}

func (s *Server) createDevice(w http.ResponseWriter, r *http.Request) {
	var input store.DeviceInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	device, err := s.store.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create device")
		return
	}
	writeJSON(w, http.StatusCreated, device)
}

func (s *Server) getDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	device, err := s.store.Get(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get device")
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (s *Server) updateDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var input store.DeviceInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	device, err := s.store.Update(r.Context(), id, input)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update device")
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (s *Server) deleteDevice(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := s.store.Delete(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "device not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete device")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			token = r.Header.Get("X-API-Key")
		}
		if token != s.apiKey {
			writeError(w, http.StatusUnauthorized, "missing or invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return 0, false
	}
	return id, true
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
