package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	defaultPort = "45566"
)

func RunServer(repo Repository) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	http.Handle("/", RepoHandler(repo))
	return http.ListenAndServe("localhost:"+port, nil)
}

type item struct {
	Key string `json:"key"`
}

type repoHandler struct {
	repo Repository
}

func RepoHandler(repo Repository) http.Handler {
	return &repoHandler{
		repo: repo,
	}
}

func (h repoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("%s", r), http.StatusInternalServerError)
		}
	}()
	h.handleIndex(w, r)
}

func (h repoHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	items := make([]item, 0)
	h.repo.Walk(func(file string) {
		items = append(items, item{
			Key: file,
		})
	})
	JSON(w, items)
}

func JSON(w http.ResponseWriter, v interface{}) {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
