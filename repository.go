package main

import (
	"bufio"
	"io"
	"path/filepath"

	"github.com/kr/fs"
	"github.com/proglottis/rwvfs"
)

const (
	idFilename    = ".gpg-id"
	fileExtension = ".gpg"
)

type Repository struct {
	fs *CryptoFS
}

func NewRepository(fs *CryptoFS) *Repository {
	return &Repository{fs: fs}
}

func (r *Repository) Init(ids []string) error {
	if err := r.fs.CheckIdentities(ids); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(r.fs, "/"); err != nil {
		return err
	}
	return r.fs.SetIdentities(ids)
}

func (r *Repository) Open(key string, passphrase []byte) (io.ReadCloser, error) {
	return r.fs.OpenEncrypted(key+fileExtension, passphrase)
}

func (r *Repository) Line(key string, passphrase []byte) (string, error) {
	plaintext, err := r.Open(key, passphrase)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	scanner := bufio.NewScanner(plaintext)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func (r *Repository) Map(key string, passphrase []byte) (map[string]string, error) {
	fileinfos, err := r.fs.ReadDir(key)
	if err != nil {
		return nil, err
	}
	keys := make(map[string]string)
	for _, fileinfo := range fileinfos {
		name := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(name) != fileExtension {
			continue
		}
		valueKey := name[:len(name)-len(fileExtension)]
		value, err := r.Line(r.fs.Join(key, valueKey), passphrase)
		if err != nil {
			panic(err)
		}
		keys[valueKey] = value
	}
	return keys, nil
}

func (r *Repository) Create(key string) (io.WriteCloser, error) {
	if err := rwvfs.MkdirAll(r.fs, filepath.Dir(key)); err != nil {
		return nil, err
	}
	return r.fs.CreateEncrypted(key + fileExtension)
}

func (r *Repository) Remove(key string) error {
	return r.fs.Remove(key + fileExtension)
}

func (r *Repository) Walk(walkFn func(file string)) error {
	walker := fs.WalkFS(".", r.fs)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		path := walker.Path()
		if walker.Stat().IsDir() || filepath.Ext(path) != fileExtension {
			continue
		}
		walkFn(path[:len(path)-len(fileExtension)])
	}
	return nil
}
