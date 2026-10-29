package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	NewRouter("http://ping-2.invalid").ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if want := `{"status":"ok"}`; w.Body.String() != want {
		t.Fatalf("body = %s, want %s", w.Body.String(), want)
	}
}

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name       string
		status     int
		body       string
		wantStatus int
	}{
		{"pong", http.StatusOK, `{"message":"pong","service":"ping-2"}`, http.StatusOK},
		{"upstream error", http.StatusServiceUnavailable, `unavailable`, http.StatusBadGateway},
		{"malformed JSON", http.StatusOK, `{`, http.StatusBadGateway},
		{"missing pong", http.StatusOK, `{}`, http.StatusBadGateway},
		{"wrong service", http.StatusOK, `{"message":"pong","service":"ping-api"}`, http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/api/v1/ping" {
					t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if _, err := w.Write([]byte(tc.body)); err != nil {
					t.Errorf("write upstream response: %v", err)
				}
			}))
			defer upstream.Close()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)

			NewRouter(upstream.URL+"/").ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK && w.Body.String() != tc.body {
				t.Fatalf("body = %s, want %s", w.Body.String(), tc.body)
			}
		})
	}
}

func TestPingUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewServer(http.NotFoundHandler())
	upstream.Close()
	for _, upstreamURL := range []string{upstream.URL, ":invalid"} {
		t.Run(upstreamURL, func(t *testing.T) {
			w := httptest.NewRecorder()
			NewRouter(upstreamURL).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))
			if w.Code != http.StatusBadGateway {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
			}
		})
	}
}

func TestPingCanceledRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := httptest.NewServer(http.NotFoundHandler())
	defer upstream.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil).WithContext(ctx)
	NewRouter(upstream.URL).ServeHTTP(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}
