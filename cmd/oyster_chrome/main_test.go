package main

import (
	"encoding/json"
	"testing"

	"github.com/proglottis/oyster"
	"github.com/sourcegraph/rwvfs"
)

func TestRequestHandler_GET(t *testing.T) {
	var response *Message
	in, out := setupHandler(t)

	in <- &Message{Type: "GET", Data: json.RawMessage("{\"key\":\"example.com\"}")}
	// We will get a callback for each file. username and password
	for i := 0; i < 2; i++ {
		response = <-out
		if response.Type != "GET_PASSWORD" {
			t.Fatalf("Expected GET_PASSWORD, got %s: %s", response.Type, string(response.Data))
		}
		in <- &Message{Type: "PASSWORD", Data: json.RawMessage("{\"passphrase\":\"password\"}")}
	}
	response = <-out
	if response.Type != "FORM" {
		t.Errorf("Expected FORM, got %s: %s", response.Type, string(response.Data))
	}
}

func setupHandler(t testing.TB) (chan<- *Message, <-chan *Message) {
	in := make(chan *Message)
	out := make(chan *Message)
	gpg := oyster.NewGpgRepo("../../testdata/gpghome")
	fs := oyster.NewCryptoFS(rwvfs.Map(map[string]string{}), gpg)
	if err := oyster.InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	repo := oyster.NewFormRepo(fs)
	form := &oyster.Form{
		Key: "example.com",
		Fields: []oyster.Field{
			{Name: "password", Value: "password123"},
			{Name: "username", Value: "bob"},
		},
	}
	if err := repo.Put(form); err != nil {
		t.Fatal(err)
	}
	handler := &RequestHandler{
		in:   in,
		out:  out,
		repo: repo,
	}
	fs.Callback = func() []byte {
		return handler.Password()
	}
	go handler.Run()
	return in, out
}
