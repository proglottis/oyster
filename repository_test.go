package main

import (
	"io/ioutil"
	"testing"

	"github.com/sourcegraph/rwvfs"
)

func TestFormRepoPutGet(t *testing.T) {
	repo := setupFormRepo(t)

	if _, err := repo.Get("test", []byte("password")); err != ErrNotFound {
		t.Error("Expected ErrNotFound, got", err)
	}

	writeform := &Form{
		Key: "test",
		Fields: []Field{
			Field{Name: "password", Value: "password123"},
			Field{Name: "username", Value: "bob"},
		},
	}
	if err := repo.Put(writeform); err != nil {
		t.Fatal(err)
	}

	readform, err := repo.Get("test", []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	if readform.Key != "test" {
		t.Errorf("Expected 'test', got %#v", readform.Key)
	}
	for i, field := range writeform.Fields {
		if readform.Fields[i] != field {
			t.Errorf("Expected %#v, got %#v", field, readform.Fields[i])
		}
	}
}

func TestFormRepoSearch(t *testing.T) {
	repo := setupFormRepo(t)
	loadTestForms(t, repo)

	// Remove parts of the URL finding all matches
	forms, err := repo.Search("http://www.example.com/foo/bar/baz")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"www.example.com/foo/bar/baz",
		"www.example.com/foo/bar",
		"example.com/foo/bar",
		"www.example.com/foo",
		"example.com/foo",
		"www.example.com",
		"example.com",
	}
	if len(forms) != len(expected) {
		t.Fatalf("Expected %d forms, got %d", len(expected), len(forms))
	}
	for i := range expected {
		if forms[i].Key != expected[i] {
			t.Errorf("Expected %#f, got %#v", expected[i], forms[i].Key)
		}
	}

	// URL is already small
	forms, err = repo.Search("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	expected = []string{ "example.com" }
	if len(forms) != len(expected) {
		t.Fatalf("Expected %d forms, got %d", len(expected), len(forms))
	}
	for i := range expected {
		if forms[i].Key != expected[i] {
			t.Errorf("Expected %#f, got %#v", expected[i], forms[i].Key)
		}
	}
}

func TestFileRepoCreateOpen(t *testing.T) {
	repo := setupFileRepo(t)

	if _, err := repo.Open("test", []byte("password")); err != ErrNotFound {
		t.Error("Expected ErrNotFound, got", err)
	}

	clearwrite, err := repo.Create("test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = clearwrite.Write([]byte("password123\nThe best password"))
	if err != nil {
		t.Fatal(err)
	}
	clearwrite.Close()

	clearread, err := repo.Open("test", []byte("password"))
	if err != nil {
		t.Fatal(err)
	}

	text, err := ioutil.ReadAll(clearread)
	if err != nil {
		t.Fatal(err)
	}

	clearread.Close()

	if string(text) != "password123\nThe best password" {
		t.Error("Expected 'password123\\nThe best password', got", string(text))
	}

	line, err := repo.Line("test", []byte("password"))
	if err != nil {
		t.Fatal(err)
	}

	if line != "password123" {
		t.Error("Expected 'password123', got", line)
	}
}

func setupFormRepo(t testing.TB) *FormRepo {
	gpg := NewGpgRepo("gpghome")
	fs := NewCryptoFS(rwvfs.Map(map[string]string{}), gpg)
	if err := InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	return NewFormRepo(fs)
}

func setupFileRepo(t testing.TB) *FileRepo {
	gpg := NewGpgRepo("gpghome")
	fs := NewCryptoFS(rwvfs.Map(map[string]string{}), gpg)
	if err := InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	return NewFileRepo(fs)
}

func loadTestForms(t testing.TB, repo *FormRepo) {
	keys := []string{
		"example.com",
		"example.com/foo",
		"example.com/foo/bar",
		"www.example.com",
		"www.example.com/foo",
		"www.example.com/foo/bar",
		"www.example.com/foo/bar/baz",
	}

	for _, key := range keys {
		form := &Form{
			Key: key,
			Fields: []Field{
				Field{Name: "password", Value: "password123"},
				Field{Name: "username", Value: "bob"},
			},
		}
		if err := repo.Put(form); err != nil {
			t.Fatal(err)
		}
	}
}
