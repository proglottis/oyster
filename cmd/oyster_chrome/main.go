package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/proglottis/oyster"
	"github.com/proglottis/oyster/config"
	"github.com/proglottis/oyster/cryptofs"
	_ "github.com/proglottis/oyster/gpgme"
	_ "github.com/proglottis/oyster/openpgp"
	"github.com/sourcegraph/rwvfs"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SearchData struct {
	Query string `json:"query"`
}

type KeyData struct {
	Key string `json:"key"`
}

type PasswordData struct {
	Passphrase string `json:"passphrase"`
}

func writeMessages(w io.Writer, messages <-chan *Message) {
	enc := NewEncoder(w)
	for {
		msg := <-messages
		if err := enc.Encode(msg); err != nil {
			panic(err)
		}
	}
}

func readMessages(r io.Reader, messages chan<- *Message) {
	dec := NewDecoder(r)
	for {
		var req Message
		if err := dec.Decode(&req); err != nil {
			switch err {
			case io.EOF:
				close(messages)
				return
			default:
				panic(err)
			}
		}
		messages <- &req
	}
}

type RequestHandler struct {
	in   <-chan *Message
	out  chan<- *Message
	repo *oyster.FormRepo
}

func (h *RequestHandler) Handle(req *Message) {
	switch req.Type {
	case "LIST":
		forms, err := h.repo.List()
		if err != nil {
			h.errorResponse(err)
			return
		}
		h.formsResponse(forms)
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
		var data KeyData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			return
		}
		form, err := h.repo.Get(data.Key)
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
		h.okResponse()
	case "REMOVE":
		var data KeyData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			return
		}
		if err := h.repo.Remove(data.Key); err != nil {
			h.errorResponse(err)
			return
		}
		h.okResponse()
	default:
		h.errorResponse(fmt.Errorf("Unknown request type: %s", req.Type))
	}
}

func (h *RequestHandler) formsResponse(forms []oyster.Form) {
	var err error
	response := &Message{Type: "FORMS"}
	response.Data, err = json.Marshal(forms)
	if err != nil {
		h.errorResponse(err)
		return
	}
	h.out <- response
}

func (h *RequestHandler) formResponse(form *oyster.Form) {
	var err error
	response := &Message{Type: "FORM"}
	response.Data, err = json.Marshal(form)
	if err != nil {
		h.errorResponse(err)
		return
	}
	h.out <- response
}

func (h *RequestHandler) okResponse() {
	var err error
	response := &Message{Type: "OK"}
	response.Data, err = json.Marshal(map[string]interface{}{})
	if err != nil {
		h.errorResponse(err)
		return
	}
	h.out <- response
}

func (h *RequestHandler) errorResponse(err error) {
	var e error
	response := &Message{Type: "ERROR"}
	response.Data, e = json.Marshal(err.Error())
	if e != nil {
		panic(e)
	}
	h.out <- response
}

func (h *RequestHandler) passwordRequired() {
	var err error
	msg := &Message{Type: "GET_PASSWORD"}
	msg.Data, err = json.Marshal(map[string]interface{}{})
	if err != nil {
		h.errorResponse(err)
		return
	}
	h.out <- msg
}

func (h *RequestHandler) Password() []byte {
	var req *Message
	h.passwordRequired()
	for {
		req = <-h.in
		if req.Type != "PASSWORD" {
			h.errorResponse(errors.New("Expecting password"))
			continue
		}
		var data PasswordData
		if err := json.Unmarshal(req.Data, &data); err != nil {
			h.errorResponse(err)
			continue
		}
		return []byte(data.Passphrase)
	}
}

func (h *RequestHandler) Run() error {
	var req *Message
	for {
		req = <-h.in
		h.Handle(req)
	}
}

func main() {
	in := make(chan *Message)
	out := make(chan *Message)
	config, err := config.Read()
	if err != nil {
		panic(err)
	}
	fs, err := cryptofs.New("gpgme", rwvfs.OSPerm(config.Home(), 0600, 0700), config)
	if err != nil {
		panic(err)
	}

	handler := &RequestHandler{
		in:   in,
		out:  out,
		repo: oyster.NewFormRepo(fs),
	}
	fs.SetCallback(func() []byte {
		return handler.Password()
	})
	go handler.Run()

	go writeMessages(os.Stdout, out)
	readMessages(os.Stdin, in)
}
