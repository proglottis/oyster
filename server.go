package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/negroni"
)

const (
	defaultPort = "45566"
)

func RunServer(repo *FormRepo) {
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

type keysHandler struct {
	repo *FormRepo
}

func (h keysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.Search(w, r)
	case "POST":
		h.GetKey(w, r)
	case "PUT":
		h.PutKey(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h keysHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	forms, err := h.repo.Search(query)
	if err != nil {
		panic(err)
	}
	JSON(w, forms)
}

type formRequest struct {
	Key string `json:"key"`
}

func (h keysHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	var form *Form
	decoder := json.NewDecoder(r.Body)
	formRequest := formRequest{}
	if err := decoder.Decode(&formRequest); err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	basic := strings.TrimPrefix(r.Header.Get("Authorization"), "Basic ")
	if basic == "" {
		JSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(basic)
	if err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	pair := bytes.Split(decoded, []byte(":"))
	passphrase := pair[1]
	form, err = h.repo.Get(formRequest.Key, passphrase)
	switch err {
	case nil: // Ignore
	case ErrNotFound:
		JSONError(w, err.Error(), http.StatusNotFound)
		return
	default:
		panic(err)
	}
	JSON(w, form)
}

func (h keysHandler) PutKey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	form := Form{}
	if err := decoder.Decode(&form); err != nil {
		JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.repo.Put(&form); err != nil {
		panic(err)
	}
	JSON(w, form)
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

type jsonError struct {
	Message string `json:"message"`
}

func JSONError(w http.ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	JSON(w, jsonError{Message: error})
}
