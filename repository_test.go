package oyster

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/proglottis/oyster/config"
	"github.com/proglottis/oyster/cryptofs"
	_ "github.com/proglottis/oyster/gpgme"
	"github.com/sourcegraph/rwvfs"
)

var testKeys = []string{
	"example.com",
	"example.com/foo",
	"example.com/foo/bar",
	"www.example.com",
	"www.example.com/foo",
	"www.example.com/foo/bar",
	"www.example.com/foo/bar/baz",
}

func TestFormRepoPutGet(t *testing.T) {
	repo := setupFormRepo(t)

	if _, err := repo.Get("test"); err != cryptofs.ErrNotFound {
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

	readform, err := repo.Get("test")
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

func TestFormRepoList(t *testing.T) {
	repo := setupFormRepo(t)
	loadTestForms(t, repo)

	forms, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}

	if len(forms) != len(testKeys) {
		t.Fatalf("Expected %d forms, got %d", len(testKeys), len(forms))
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
	expected = []string{"example.com"}
	if len(forms) != len(expected) {
		t.Fatalf("Expected %d forms, got %d", len(expected), len(forms))
	}
	for i := range expected {
		if forms[i].Key != expected[i] {
			t.Errorf("Expected %#f, got %#v", expected[i], forms[i].Key)
		}
	}
}

func TestFormRepoSearch_no_fields(t *testing.T) {
	repo := setupFormRepo(t)
	repo.Put(&Form{
		Key:    "example.com",
		Fields: FieldSlice{},
	})
	forms, err := repo.Search("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(forms) > 0 {
		t.Errorf("Expected no results, got %#v", len(forms))
	}
}

func TestFormRepoRemove(t *testing.T) {
	repo := setupFormRepo(t)
	loadTestForms(t, repo)

	if err := repo.Remove(testKeys[0]); err != nil {
		t.Fatal(err)
	}
}

func TestFileRepoCreateOpen(t *testing.T) {
	repo := setupFileRepo(t)

	if _, err := repo.Open("test"); err != cryptofs.ErrNotFound {
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

	clearread, err := repo.Open("test")
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

	line, err := repo.Line("test")
	if err != nil {
		t.Fatal(err)
	}

	if line != "password123" {
		t.Error("Expected 'password123', got", line)
	}
}

func setupFormRepo(t testing.TB) *FormRepo {
	os.Setenv("GNUPGHOME", "./testdata/gpghome")
	fs, err := cryptofs.New("gpgme", rwvfs.Map(map[string]string{}), config.New())
	if err != nil {
		t.Fatal(err)
	}
	fs.SetCallback(func() []byte { return []byte("password") })
	if err := InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	return NewFormRepo(fs)
}

func setupFileRepo(t testing.TB) *FileRepo {
	os.Setenv("GNUPGHOME", "./testdata/gpghome")
	fs, err := cryptofs.New("gpgme", rwvfs.Map(map[string]string{}), config.New())
	if err != nil {
		t.Fatal(err)
	}
	fs.SetCallback(func() []byte { return []byte("password") })
	if err := InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	return NewFileRepo(fs)
}

func loadTestForms(t testing.TB, repo *FormRepo) {
	for _, key := range testKeys {
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
