package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	apiCfg := &ApiConfig{DB: nil}

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	apiCfg.HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", rr.Code)
	}

	expectedBody := "I am alive!"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q, but got %q", expectedBody, rr.Body.String())
	}

}

func TestAuthMiddleware(t *testing.T) {
	apiCfg := &ApiConfig{
		AuthSecret: "mysecret",
	}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handlerToTest := apiCfg.AuthMiddleware(nextHandler)

	req, _ := http.NewRequest("POST", "/users", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing token, got %d", rr.Code)
	}

	req, _ = http.NewRequest("POST", "/users", nil)
	req.Header.Set("Authorization", "Bearer wrong_password")
	rr = httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for wrong token, got %d", rr.Code)
	}

	req, _ = http.NewRequest("POST", "/users", nil)
	req.Header.Set("Authorization", "Bearer mysecret")
	rr = httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for correct token, got %d", rr.Code)
	}
}
