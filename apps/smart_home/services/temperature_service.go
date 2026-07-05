package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TemperatureService handles fetching temperature data from external API
type TemperatureService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// TemperatureResponse represents the response from the temperature API
type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Temperature float64   `json:"temperature"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorIDAlt string    `json:"sensorId"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

// NewTemperatureService creates a new temperature service
func NewTemperatureService(baseURL string) *TemperatureService {
	return &TemperatureService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetTemperature fetches temperature data for a specific location
func (s *TemperatureService) GetTemperature(location string) (*TemperatureResponse, error) {
	url := fmt.Sprintf("%s/temperature?location=%s", s.BaseURL, url.QueryEscape(location))

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching temperature data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var temperatureResp TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&temperatureResp); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	temperatureResp.Normalize()
	return &temperatureResp, nil
}

// GetTemperatureByID fetches temperature data for a specific sensor ID
func (s *TemperatureService) GetTemperatureByID(sensorID string) (*TemperatureResponse, error) {
	url := fmt.Sprintf("%s/temperature?sensorId=%s", s.BaseURL, url.QueryEscape(sensorID))

	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching temperature data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var temperatureResp TemperatureResponse
	if err := json.NewDecoder(resp.Body).Decode(&temperatureResp); err != nil {
		return nil, fmt.Errorf("error decoding temperature response: %w", err)
	}

	temperatureResp.Normalize()
	return &temperatureResp, nil
}

func (r *TemperatureResponse) Normalize() {
	if r.Value == 0 && r.Temperature != 0 {
		r.Value = r.Temperature
	}
	if r.SensorID == "" {
		r.SensorID = r.SensorIDAlt
	}
	if r.Unit == "" {
		r.Unit = "C"
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if r.SensorType == "" {
		r.SensorType = "temperature"
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now().UTC()
	}
}
