package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeysHandlerPOST_auth(t *testing.T) {
	repo := setupFormRepo(t)
	if err := repo.Put(&Form{FormRequest: FormRequest{Key: "test.com"}, Fields: map[string]string{"username": "bob", "password": "password123"}}); err != nil {
		t.Fatal(err)
	}

	handler := &keysHandler{repo: repo}
	body := strings.NewReader(`{"url":"http://test.com"}`)
	req, err := http.NewRequest("POST", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("passd", "password")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %#v", w.Code)
	}
	expected := `{"key":"test.com","url":"http://test.com","fields":{"password":"password123","username":"bob"}}
`
	response := w.Body.String()
	if response != expected {
		t.Errorf("Expected %#v, got %#v", expected, response)
	}
}

func TestKeysHandlerPOST_no_auth(t *testing.T) {
	repo := setupFormRepo(t)
	if err := repo.Put(&Form{
		FormRequest: FormRequest{Key: "test.com"},
		Fields:      map[string]string{"username": "bob", "password": "password123"},
	}); err != nil {
		t.Fatal(err)
	}

	handler := &keysHandler{repo: repo}
	body := strings.NewReader(`{"url":"http://test.com"}`)
	req, err := http.NewRequest("POST", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %#v", w.Code)
	}
	expected := `{"key":"test.com","url":"http://test.com","fields":{"password":"","username":""}}
`
	response := w.Body.String()
	if response != expected {
		t.Errorf("Expected %#v, got %#v", expected, response)
	}
}

func TestKeysHandlerPUT(t *testing.T) {
	repo := setupFormRepo(t)

	handler := &keysHandler{repo: repo}
	body := strings.NewReader(`{"url":"http://test.com","fields":{"username":"bob","password":"password123"}}`)
	req, err := http.NewRequest("PUT", "/keys", body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %#v", w.Code)
	}
	expected := `{"key":"test.com","url":"http://test.com","fields":{"password":"password123","username":"bob"}}
`
	response := w.Body.String()
	if response != expected {
		t.Errorf("Expected %#v, got %#v", expected, response)
	}
}
