package controllers

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prest/prest/adapters/postgres"
	"github.com/prest/prest/config"
	"github.com/prest/prest/testutils"
	"github.com/stretchr/testify/assert"
)

func initAuthRoutes() *mux.Router {
	r := mux.NewRouter()
	dbh := New()
	// if auth is enabled
	if config.PrestConf.AuthEnabled {
		r.HandleFunc("/auth/{database}", dbh.Auth).Methods("POST")
	}
	return r
}

func Test_basicPasswordCheck(t *testing.T) {
	config.Load()
	postgres.Load()
	db := DB{
		Adapter: config.PrestConf.Adapter,
		Config:  *config.PrestConf,
	}
	_, err := basicPasswordCheck(db, "test@postgres.rest", "123456")
	assert.Nil(t, err)
}

func Test_getSelectQuery(t *testing.T) {
	config.Load()

	expected := "SELECT * FROM prest_users WHERE username=$1 AND password=$2 LIMIT 1"
	query := getSelectQuery()
	assert.Equal(t, query, expected)
}

func Test_encrypt(t *testing.T) {
	config.Load()

	pwd := "123456"
	enc := encrypt(*config.PrestConf, pwd)
	md5Enc := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
	assert.Equal(t, enc, md5Enc)

	config.PrestConf.AuthEncrypt = "SHA1"
	enc = encrypt(*config.PrestConf, pwd)
	sha1Enc := fmt.Sprintf("%x", sha1.Sum([]byte(pwd)))
	assert.Equal(t, enc, sha1Enc)
}

func TestAuthDisable(t *testing.T) {
	server := httptest.NewServer(initAuthRoutes())
	defer server.Close()

	t.Log("/auth request POST method, disable auth")
	testutils.DoRequest(t, server.URL+"/auth/prest-test", nil, "POST", http.StatusNotFound, "AuthDisable")
}

func TestAuthEnable_GET(t *testing.T) {
	config.Load()
	postgres.Load()
	config.PrestConf.AuthEnabled = true

	server := httptest.NewServer(initAuthRoutes())
	defer server.Close()

	var testCase = struct {
		description string
		url         string
		method      string
		status      int
	}{"/auth request GET method", "/auth/prest-test", "GET", http.StatusMethodNotAllowed}
	t.Log(testCase.description)
	testutils.DoRequest(t, server.URL+testCase.url, nil, testCase.method, testCase.status, "AuthEnable")
}

func TestAuthEnable_POST(t *testing.T) {
	config.Load()
	postgres.Load()
	config.PrestConf.AuthEnabled = true

	server := httptest.NewServer(initAuthRoutes())
	defer server.Close()

	var testCase = struct {
		description string
		url         string
		method      string
		status      int
	}{"/auth request POST method", "/auth/prest-test", "POST", http.StatusUnauthorized}
	t.Log(testCase.description)
	testutils.DoRequest(t, server.URL+testCase.url, nil, testCase.method, testCase.status, "AuthEnable")
}
