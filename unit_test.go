package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePostRequest(t *testing.T) {
	// Mock input JSON
	input := map[string]interface{}{
		"message": "Hello, MovieVerse!",
	}
	inputBytes, _ := json.Marshal(input)

	req, err := http.NewRequest(http.MethodPost, "/post", bytes.NewBuffer(inputBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlePostRequest)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", response["status"])
	}
	if response["message"] != "Data successfully received" {
		t.Errorf("Expected message 'Data successfully received', got '%s'", response["message"])
	}
}

func TestHandlePostRequest_InvalidJSON(t *testing.T) {
	input := `{"invalid":}`
	req, err := http.NewRequest(http.MethodPost, "/post", bytes.NewBufferString(input))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlePostRequest)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandlePostRequest_EmptyMessage(t *testing.T) {
	input := map[string]interface{}{
		"message": "",
	}
	inputBytes, _ := json.Marshal(input)

	req, err := http.NewRequest(http.MethodPost, "/post", bytes.NewBuffer(inputBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlePostRequest)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
