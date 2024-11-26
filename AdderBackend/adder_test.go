package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendAdder(t *testing.T) {
	mockCounter := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected method POST, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer mockCounter.Close()

	handler := addHandler(mockCounter.URL)

	req := httptest.NewRequest(http.MethodPost, "/add", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %s", resp.Status)
	}

	if strings.TrimSpace(string(body)) != "Increment sent to counter backend." {
		t.Fatalf("expected body to be 'Increment sent to counter backend.' got %s", string(body))
	}
}

func TestIndexHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	indexHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %s", resp.Status)
	}

	if !strings.Contains(string(body), "Adder Backend") {
		t.Fatalf("expected body to contain 'Adder Backend', got %s", body)
	}
}