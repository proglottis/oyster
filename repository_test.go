package main

import (
	"io/ioutil"
	"testing"

	"github.com/proglottis/rwvfs"
)

func TestRepositoryCreateOpen(t *testing.T) {
	gpg := NewGpgRepo("gpghome")
	fs := NewCryptoFS(rwvfs.Map(map[string]string{}), gpg)
	repo := NewRepository(fs)

	if err := repo.Init([]string{"test@example.com"}); err != nil {
		t.Error(err)
	}

	clearwrite, err := repo.Create("test")
	if err != nil {
		t.Error(err)
	}
	_, err = clearwrite.Write([]byte("password123"))
	if err != nil {
		t.Error(err)
	}
	clearwrite.Close()

	clearread, err := repo.Open("test", []byte("password"))
	if err != nil {
		t.Error(err)
	}

	text, err := ioutil.ReadAll(clearread)
	if err != nil {
		t.Error(err)
	}

	clearread.Close()

	if string(text) != "password123" {
		t.Error("Expected 'password', got", string(text))
	}
}
