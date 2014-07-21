package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	http.Handle("/keys", defaultHandler(http.StripPrefix("/keys", &keysHandler{repo: repo})))
	http.Handle("/keys/", defaultHandler(http.StripPrefix("/keys", &keyHandler{repo: repo})))
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
		h.GetKeys(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h keysHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	items := make([]item, 0)
	h.repo.Walk(func(file string) {
		items = append(items, item{
			Key: file,
		})
	})
	JSON(w, items)
}

type keyHandler struct {
	repo Repository
}

func (h keyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.GetKey(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h keyHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	passphrase := []byte(r.PostForm.Get("passphrase"))
	key := r.URL.Path
	plaintext, err := h.repo.Get(key, passphrase)
	if err != nil {
		panic(err)
	}
	defer plaintext.Close()
	value, err := ioutil.ReadAll(plaintext)
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
