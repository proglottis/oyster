package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

import (
	"github.com/gorilla/handlers"
)

const (
	defaultPort = "45566"
)

func RunServer(repo Repository) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	http.Handle("/keys", defaultHandler(&keysHandler{repo: repo}))
	return http.ListenAndServe("localhost:"+port, nil)
}

func defaultHandler(handler http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, &panicHandler{next: handler})
}

type panicHandler struct {
	next http.Handler
}

func (h panicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("%s", r), http.StatusInternalServerError)
		}
	}()
	h.next.ServeHTTP(w, r)
}

type item struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
}

type keysHandler struct {
	repo Repository
}

func (h keysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.GetKey(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h keysHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}
	keyurl, err := url.Parse(r.PostForm.Get("url"))
	if err != nil {
		panic(err)
	}
	passphrase := []byte(r.PostForm.Get("passphrase"))
	key := keyurl.Host + keyurl.Path
	value, err := h.repo.GetLine(key, passphrase)
	if err != nil {
		panic(err)
	}
	JSON(w, item{Key: key, Value: value})
}

func JSON(w http.ResponseWriter, v interface{}) {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
