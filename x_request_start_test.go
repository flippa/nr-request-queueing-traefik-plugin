package nr_request_queueing_traefik_plugin //nolint:revive,stylecheck

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCreateConfig(t *testing.T) {
	config := CreateConfig()
	if config == nil {
		t.Fatal("CreateConfig should return a non-nil config")
	}
}

func TestNew(t *testing.T) {
	config := CreateConfig()
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(context.Background(), nextHandler, config, "test-plugin")
	if err != nil {
		t.Fatalf("New should not return an error: %v", err)
	}

	if handler == nil {
		t.Fatal("New should return a non-nil handler")
	}

	// Type assertion to check if it's the correct type
	plugin, ok := handler.(*XRequestStart)
	if !ok {
		t.Fatal("New should return an XRequestStart instance")
	}

	if plugin.name != "test-plugin" {
		t.Errorf("Expected name to be 'test-plugin', got '%s'", plugin.name)
	}

	if plugin.next == nil {
		t.Error("Expected next handler to be set")
	}
}

func TestServeHTTP(t *testing.T) {
	config := CreateConfig()

	// Create a mock next handler that captures the request
	var capturedRequest *http.Request
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capturedRequest = req
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
	})

	handler, err := New(context.Background(), nextHandler, config, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	// Record the time before the request
	beforeTime := time.Now().UnixMilli()

	// Serve the request
	handler.ServeHTTP(rw, req)

	// Record the time after the request
	afterTime := time.Now().UnixMilli()

	// Check that the next handler was called
	if capturedRequest == nil {
		t.Fatal("Next handler was not called")
	}

	// Check that X-Request-Start header was added
	xRequestStart := capturedRequest.Header.Get("X-Request-Start")
	if xRequestStart == "" {
		t.Fatal("X-Request-Start header was not set")
	}

	// Validate the header format
	if !strings.HasPrefix(xRequestStart, "t=") {
		t.Errorf("X-Request-Start header should start with 't=', got: %s", xRequestStart)
	}

	// Extract and validate the timestamp
	timestampStr := strings.TrimPrefix(xRequestStart, "t=")
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		t.Errorf("Failed to parse timestamp from X-Request-Start header: %v", err)
	}

	// Check that the timestamp is reasonable (within the time window of the test)
	if timestamp < beforeTime || timestamp > afterTime {
		t.Errorf("Timestamp %d is not within expected range [%d, %d]", timestamp, beforeTime, afterTime)
	}

	// Check that the response was successful
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rw.Code)
	}

	if rw.Body.String() != "OK" {
		t.Errorf("Expected response body 'OK', got '%s'", rw.Body.String())
	}
}

func TestServeHTTP_MultipleRequests(t *testing.T) {
	config := CreateConfig()

	var timestamps []string
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		timestamps = append(timestamps, req.Header.Get("X-Request-Start"))
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(context.Background(), nextHandler, config, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Make multiple requests with small delays
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/test%d", i), nil)
		rw := httptest.NewRecorder()

		handler.ServeHTTP(rw, req)

		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// Verify we have 3 timestamps
	if len(timestamps) != 3 {
		t.Fatalf("Expected 3 timestamps, got %d", len(timestamps))
	}

	// Verify all timestamps are unique and properly formatted
	seen := make(map[string]bool)
	for i, ts := range timestamps {
		if ts == "" {
			t.Errorf("Timestamp %d is empty", i)
			continue
		}

		if !strings.HasPrefix(ts, "t=") {
			t.Errorf("Timestamp %d does not start with 't=': %s", i, ts)
			continue
		}

		if seen[ts] {
			t.Errorf("Duplicate timestamp found: %s", ts)
		}
		seen[ts] = true

		// Validate the timestamp can be parsed
		timestampStr := strings.TrimPrefix(ts, "t=")
		_, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			t.Errorf("Failed to parse timestamp %d: %v", i, err)
		}
	}
}

func TestUnixMilliStr(t *testing.T) {
	// Test that unixMilliStr returns a valid timestamp
	result := unixMilliStr()

	// Should be able to parse as int64
	timestamp, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Errorf("unixMilliStr should return a valid integer string: %v", err)
	}

	// Should be close to current time (within 1 second)
	now := time.Now().UnixMilli()
	if abs(timestamp-now) > 1000 {
		t.Errorf("Timestamp %d is too far from current time %d", timestamp, now)
	}
}

// Helper function for absolute value
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
