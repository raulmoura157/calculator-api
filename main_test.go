package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPI(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{"sum", "/api/sum?term_one=4&term_two=1", 200, `{"result":5}`},
		{"sub", "/api/sub?term_one=4&term_two=1", 200, `{"result":3}`},
		{"mul", "/api/mul?term_one=4&term_two=3", 200, `{"result":12}`},
		{"div", "/api/div?term_one=8&term_two=2", 200, `{"result":4}`},
		{"div truncates", "/api/div?term_one=5&term_two=2", 200, `{"result":2}`},
		{"negative numbers", "/api/sub?term_one=-4&term_two=6", 200, `{"result":-10}`},
		{"div by zero", "/api/div?term_one=1&term_two=0", 400, `{"error":"division by zero"}`},
		{"missing param", "/api/sum?term_one=1", 400, `{"error":"term_one and term_two must be integers"}`},
		{"non integer", "/api/mul?term_one=a&term_two=2", 400, `{"error":"term_one and term_two must be integers"}`},
		{"decimal", "/api/sum?term_one=1.5&term_two=2", 400, `{"error":"term_one and term_two must be integers"}`},
		{"healthz", "/healthz", 200, `{"status":"ok"}`},
		{"unknown op", "/api/pow?term_one=2&term_two=2", 404, ""},
	}
	srv := newMux()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if tc.body == "" {
				return
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tc.body {
				t.Fatalf("body = %s, want %s", got, tc.body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("content-type = %q", ct)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/sum?term_one=1&term_two=1", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
