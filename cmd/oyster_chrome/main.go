package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/proglottis/oyster"
	"github.com/sourcegraph/rwvfs"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SearchData struct {
	Query string `json:"query"`
}

type GetData struct {
	Key        string `json:"key"`
	Passphrase string `json:"passphrase"`
}

func readRequests(r io.Reader, requests chan<- *Message) {
	dec := NewDecoder(r)
	for {
		var req Message
		if err := dec.Decode(&req); err != nil {
			switch err {
			case io.EOF:
				close(requests)
				return
			default:
				panic(err)
			}
		}
		select {
		case requests <- &req:
		}
	}
}

type RequestHandler struct {
	requests <-chan *Message
	enc      *Encoder
	repo     *oyster.FormRepo
}

func (h *RequestHandler) Handle(req *Message) {
	switch req.Type {
	case "SEARCH":
		var data SearchData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			return
		}
		forms, err := h.repo.Search(data.Query)
		if err != nil {
			h.errorResponse(err)
			return
		}
		h.formsResponse(forms)
	case "GET":
		var data GetData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			return
		}
		form, err := h.repo.Get(data.Key, []byte(data.Passphrase))
		if err != nil {
			h.errorResponse(err)
			return
		}
		h.formResponse(form)
	case "PUT":
		var data oyster.Form
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			return
		}
		if err := h.repo.Put(&data); err != nil {
			h.errorResponse(err)
			return
		}
	default:
		h.errorResponse(errors.New("Unknown request type"))
	}
}

func (h *RequestHandler) formsResponse(forms []oyster.Form) {
	var err error
	response := &Message{Type: "FORMS"}
	response.Data, err = json.Marshal(forms)
	if err != nil {
		h.errorResponse(err)
	}
	if err := h.enc.Encode(response); err != nil {
		h.errorResponse(err)
	}
}

func (h *RequestHandler) formResponse(form *oyster.Form) {
	var err error
	response := &Message{Type: "FORM"}
	response.Data, err = json.Marshal(form)
	if err != nil {
		h.errorResponse(err)
	}
	if err := h.enc.Encode(response); err != nil {
		h.errorResponse(err)
	}
}

func (h *RequestHandler) errorResponse(err error) {
	var e error
	response := &Message{Type: "ERROR"}
	response.Data, e = json.Marshal(err.Error())
	if e != nil {
		panic(e)
	}
	if err := h.enc.Encode(response); err != nil {
		panic(err)
	}
}

func (h *RequestHandler) Run() error {
	for {
		select {
		case req := <-h.requests:
			h.Handle(req)
		}
	}
}

func main() {
	requests := make(chan *Message)
	gpg := oyster.NewGpgRepo(oyster.GpgHome())
	fs := oyster.NewCryptoFS(rwvfs.OSPerm(oyster.Home(), 0600, 0700), gpg)

	handler := &RequestHandler{
		requests: requests,
		enc:      NewEncoder(os.Stdout),
		repo:     oyster.NewFormRepo(fs),
	}
	go handler.Run()

	readRequests(os.Stdin, requests)
}
