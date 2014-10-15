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
	case "POST":
		h.GetKey(w, r)
	case "PUT":
		h.PutKey(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h keysHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	var form *Form
	decoder := json.NewDecoder(r.Body)
	formRequest := FormRequest{}
	if err := decoder.Decode(&formRequest); err != nil {
		panic(err)
	}
	basic := strings.TrimPrefix(r.Header.Get("Authorization"), "Basic ")
	if basic != "" {
		decoded, err := base64.StdEncoding.DecodeString(basic)
		if err != nil {
			panic(err)
		}
		pair := bytes.Split(decoded, []byte(":"))
		passphrase := pair[1]
		form, err = h.repo.Get(&formRequest, passphrase)
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		form, err = h.repo.Fields(&formRequest)
		if err != nil {
			panic(err)
		}
	}
	JSON(w, form)
}

func (h keysHandler) PutKey(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	form := Form{}
	if err := decoder.Decode(&form); err != nil {
		panic(err)
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
