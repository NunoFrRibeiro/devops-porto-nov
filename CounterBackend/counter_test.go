package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCounterHandler(t *testing.T) {
	mutex.Lock()
	counter = 0
	mutex.Unlock()

	req := httptest.NewRequest(http.MethodPost, "/increment", nil)
	w := httptest.NewRecorder()

	counterHandler(w, req)

	resp := w.Result()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %v", resp.Status)
	}

	expected := "1"
	if string(body) != expected {
		t.Fatalf("expected body to be '%s' got: '%s'", expected, body)
	}
}

func TestCounterPartialHandler(t *testing.T) {
	mutex.Lock()
	counter = 5
	mutex.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/counter", nil)
	w := httptest.NewRecorder()

	counterPartialHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	t.Logf("Counter Value Before Request: %d", counter)
	t.Logf("Response Body: '%s'", string(body))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got: %v", resp.Status)
	}

	expected := "5"
	if string(body) != expected {
		t.Fatalf("expected body to be '%s', got: '%s'", expected, body)
	}
}
