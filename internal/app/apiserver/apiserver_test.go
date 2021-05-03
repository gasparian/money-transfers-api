package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO:
func TestAPIServer_HandleHealth(t *testing.T) {
	s := New(NewConfig())
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	s.handleHealth().ServeHTTP(recorder, req)
	if recorder.Body.String() != "OK" {
		t.Error("Healthcheck failed")
	}
}
