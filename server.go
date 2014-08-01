package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
)

import (
	"github.com/codegangsta/negroni"
)

const (
	defaultPort = "45566"
)

func RunServer(repo Repository) {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	mux := http.NewServeMux()
	mux.Handle("/keys", &keysHandler{repo: repo})

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(noCache))
	n.UseHandler(mux)
	n.Run("localhost:" + port)
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
	case "PUT":
		h.PutKey(w, r)
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

func (h keysHandler) PutKey(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}
	keyurl, err := url.Parse(r.PostForm.Get("url"))
	if err != nil {
		panic(err)
	}
	key := keyurl.Host + keyurl.Path
	plaintext, err := h.repo.Put(key)
	if err != nil {
		panic(err)
	}
	value := r.PostForm.Get("text")
	io.WriteString(plaintext, value+"\n")
	plaintext.Close()
	JSON(w, item{Key: key, Value: value})
}

func noCache(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Cache-Control", "max-age=0, no-store, no-cache")
	next(w, r)
}

func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		panic(err)
	}
}
