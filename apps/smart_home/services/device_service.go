package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DeviceService struct {
	BaseURL    string
	HTTPClient *http.Client
}

type DeviceRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name,omitempty"`
	Type     string         `json:"type,omitempty"`
	Location string         `json:"location,omitempty"`
	Status   string         `json:"status,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func NewDeviceService(baseURL string) *DeviceService {
	return &DeviceService{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *DeviceService) CreateDevice(ctx context.Context, req DeviceRequest) error {
	return s.sendJSON(ctx, http.MethodPost, "/api/devices", req)
}

func (s *DeviceService) UpdateDevice(ctx context.Context, deviceID string, req DeviceRequest) error {
	return s.sendJSON(ctx, http.MethodPut, "/api/devices/"+deviceID, req)
}

func (s *DeviceService) DeleteDevice(ctx context.Context, deviceID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.BaseURL+"/api/devices/"+deviceID, nil)
	if err != nil {
		return err
	}

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("delete device request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("device service returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *DeviceService) sendJSON(ctx context.Context, method, path string, req any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, s.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("device service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("device service returned status %d", resp.StatusCode)
	}
	return nil
}
