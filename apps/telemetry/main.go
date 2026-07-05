package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TelemetryReading struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ReadingRequest struct {
	DeviceID  string   `json:"device_id"`
	Metric    string   `json:"metric"`
	Value     *float64 `json:"value"`
	Unit      string   `json:"unit"`
	Timestamp string   `json:"timestamp"`
}

type Store struct {
	mu       sync.RWMutex
	readings map[string]TelemetryReading
}

func NewStore() *Store {
	return &Store{readings: make(map[string]TelemetryReading)}
}

func (s *Store) Add(req ReadingRequest) (TelemetryReading, error) {
	if strings.TrimSpace(req.DeviceID) == "" {
		return TelemetryReading{}, errors.New("device_id is required")
	}
	if strings.TrimSpace(req.Metric) == "" {
		return TelemetryReading{}, errors.New("metric is required")
	}
	if req.Value == nil {
		return TelemetryReading{}, errors.New("value is required")
	}

	timestamp := time.Now().UTC()
	if strings.TrimSpace(req.Timestamp) != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			return TelemetryReading{}, errors.New("timestamp must be RFC3339")
		}
		timestamp = parsed.UTC()
	}

	reading := TelemetryReading{
		ID:        newID(),
		DeviceID:  strings.TrimSpace(req.DeviceID),
		Metric:    strings.TrimSpace(req.Metric),
		Value:     *req.Value,
		Unit:      strings.TrimSpace(req.Unit),
		Timestamp: timestamp,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.readings[reading.ID] = reading
	return reading, nil
}

func (s *Store) List(deviceID, metric string, limit int) []TelemetryReading {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]TelemetryReading, 0, len(s.readings))
	for _, reading := range s.readings {
		if deviceID != "" && reading.DeviceID != deviceID {
			continue
		}
		if metric != "" && reading.Metric != metric {
			continue
		}
		result = append(result, reading)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	if limit > 0 && len(result) > limit {
		return result[:limit]
	}
	return result
}

func (s *Store) Get(id string) (TelemetryReading, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	reading, ok := s.readings[id]
	return reading, ok
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.readings[id]; !ok {
		return false
	}
	delete(s.readings, id)
	return true
}

func (s *Store) LatestByDevice(deviceID string) map[string]TelemetryReading {
	s.mu.RLock()
	defer s.mu.RUnlock()

	latest := make(map[string]TelemetryReading)
	for _, reading := range s.readings {
		if reading.DeviceID != deviceID {
			continue
		}
		current, ok := latest[reading.Metric]
		if !ok || reading.Timestamp.After(current.Timestamp) {
			latest[reading.Metric] = reading
		}
	}
	return latest
}

func main() {
	store := NewStore()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "telemetry"})
	})

	mux.HandleFunc("POST /api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		var req ReadingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		reading, err := store.Add(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, reading)
	})

	mux.HandleFunc("GET /api/telemetry", func(w http.ResponseWriter, r *http.Request) {
		limit := parseLimit(r.URL.Query().Get("limit"), 100)
		readings := store.List(
			r.URL.Query().Get("device_id"),
			r.URL.Query().Get("metric"),
			limit,
		)
		writeJSON(w, http.StatusOK, readings)
	})

	mux.HandleFunc("GET /api/telemetry/{id}", func(w http.ResponseWriter, r *http.Request) {
		reading, ok := store.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "telemetry reading not found")
			return
		}
		writeJSON(w, http.StatusOK, reading)
	})

	mux.HandleFunc("DELETE /api/telemetry/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !store.Delete(r.PathValue("id")) {
			writeError(w, http.StatusNotFound, "telemetry reading not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /api/telemetry/devices/{device_id}/latest", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.LatestByDevice(r.PathValue("device_id")))
	})

	port := env("PORT", "8083")
	log.Printf("telemetry service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withLogging(mux)); err != nil {
		log.Fatal(err)
	}
}

func parseLimit(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return fallback
	}
	return limit
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(bytes[:])
}
