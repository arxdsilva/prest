package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prest/prest/adapters/postgres"
	"github.com/prest/prest/config"
	"github.com/prest/prest/testutils"
)

func init() {
	// load configuration at the beginning of the tests
	config.Load()
	postgres.Load()
}

func TestInitRouter(t *testing.T) {
	initRouter()
	if router == nil {
		t.Errorf("Router should not be nil.")
	}
}

func TestRoutes(t *testing.T) {
	router = nil
	r := Routes()
	if r == nil {
		t.Errorf("Should return a router.")
	}
}

func TestDefaultRouters(t *testing.T) {
	server := httptest.NewServer(GetRouter())
	defer server.Close()

	var testCases = []struct {
		url    string
		method string
		status int
	}{
		{"/databases", "GET", http.StatusOK},
		{"/schemas", "GET", http.StatusOK},
		{"/_QUERIES/{database}/{queriesLocation}/{script}", "GET", http.StatusBadRequest},
		{"/schemas/{database}/{schema}", "GET", http.StatusBadRequest},
		{"/show/{database}/{schema}/{table}", "GET", http.StatusBadRequest},
		{"/operations/{database}/{schema}/{table}", "GET", http.StatusUnauthorized},
		{"/operations/{database}/{schema}/{table}", "POST", http.StatusUnauthorized},
		{"/operations/batch/{database}/{schema}/{table}", "POST", http.StatusBadRequest},
		{"/operations/{database}/{schema}/{table}", "DELETE", http.StatusUnauthorized},
		{"/operations/{database}/{schema}/{table}", "PUT", http.StatusUnauthorized},
		{"/operations/{database}/{schema}/{table}", "PATCH", http.StatusUnauthorized},
		{"/auth/{databasse}", "GET", http.StatusNotFound},
		{"/", "GET", http.StatusNotFound},
	}
	for _, tc := range testCases {
		t.Log(tc.method, "\t", tc.url)
		testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, tc.url)
	}
}

func TestAuthRouterActive(t *testing.T) {
	config.PrestConf.AuthEnabled = true
	initRouter()
	server := httptest.NewServer(GetRouter())
	testutils.DoRequest(t, server.URL+"/auth/prest-test", nil, "GET", http.StatusNotFound, "AuthEnable")
	testutils.DoRequest(t, server.URL+"/auth/prest-test", nil, "POST", http.StatusUnauthorized, "AuthEnable")
}
