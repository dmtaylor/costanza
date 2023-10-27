package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/healthcheck", nil)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(Healthcheck)
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d got status %d", http.StatusOK, w.Code)
	}
}
