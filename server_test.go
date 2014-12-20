package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeysHandlerPOST_auth(t *testing.T) {
	repo := setupFormRepo(t)
	if err := repo.Put(&Form{Key: "test.com", Fields: []Field{Field{Name: "username", Value: "bob"}, Field{Name: "password", Value: "password123"}}}); err != nil {
		t.Fatal(err)
	}

	handler := &keysHandler{repo: repo}
	body := strings.NewReader(`{"key":"test.com"}`)
	req, err := http.NewRequest("POST", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("oyster", "password")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %#v", w.Code)
	}
	expected := `{"key":"test.com","fields":[{"name":"password","value":"password123"},{"name":"username","value":"bob"}]}
`
	response := w.Body.String()
	if response != expected {
		t.Errorf("Expected %#v, got %#v", expected, response)
	}
}

func TestKeysHandlerPOST_bad_auth(t *testing.T) {
	repo := setupFormRepo(t)
	if err := repo.Put(&Form{Key: "test.com", Fields: []Field{Field{Name: "username", Value: "bob"}, Field{Name: "password", Value: "password123"}}}); err != nil {
		t.Fatal(err)
	}

	handler := &keysHandler{repo: repo}

	// No auth
	body := strings.NewReader(`{"key":"test.com"}`)
	req, err := http.NewRequest("POST", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("Expected 401, got %#v", w.Code)
	}

	// Bad auth
	body = strings.NewReader(`{"key":"test.com"}`)
	req, err = http.NewRequest("POST", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("oyster", "bad_password")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("Expected 401, got %#v", w.Code)
	}
}

func TestKeysHandlerPUT(t *testing.T) {
	repo := setupFormRepo(t)

	handler := &keysHandler{repo: repo}
	body := strings.NewReader(`{"key":"test.com","fields":[{"name":"username","value":"bob"},{"name":"password","value":"password123"}]}`)
	req, err := http.NewRequest("PUT", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %#v", w.Code)
	}
	expected := `{"key":"test.com","fields":[{"name":"username","value":"bob"},{"name":"password","value":"password123"}]}
`
	response := w.Body.String()
	if response != expected {
		t.Errorf("Expected %#v, got %#v", expected, response)
	}
}
