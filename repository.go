package main

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kr/fs"
	"github.com/proglottis/rwvfs"
)

const (
	idFilename    = ".gpg-id"
	fileExtension = ".gpg"
)

type Repository interface {
	Init(ids []string) error
	Open(key string, passphrase []byte) (io.ReadCloser, error)
	Line(key string, passphrase []byte) (string, error)
	Map(key string, passphrase []byte) (map[string]string, error)
	Create(key string) (io.WriteCloser, error)
	Remove(key string) error
	Walk(walkFn func(file string)) error
}

type fileRepository struct {
	fs       rwvfs.WalkableFileSystem
	entities EntityRepo
}

func NewRepository(fs rwvfs.WalkableFileSystem, entities EntityRepo) Repository {
	return &fileRepository{
		fs:       fs,
		entities: entities,
	}
}

func (r fileRepository) checkPublicKeyRingIds(ids []string) error {
	el, err := r.entities.PublicKeyRing(ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !IdMatchesAnyEntity(id, el) {
			return fmt.Errorf("No matching public key %s", id)
		}
	}
	return nil
}

func (r fileRepository) checkSecureKeyRingIds(ids []string) error {
	el, err := r.entities.SecureKeyRing(ids)
	if err != nil {
		return err
	}
	if len(el) < 1 {
		return fmt.Errorf("No matching secure keys")
	}
	return nil
}

func (r fileRepository) Init(ids []string) error {
	if err := r.checkPublicKeyRingIds(ids); err != nil {
		return err
	}
	if err := r.checkSecureKeyRingIds(ids); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(r.fs, "/"); err != nil {
		return err
	}
	idfile, err := r.fs.Create(idFilename)
	if err != nil {
		return err
	}
	defer idfile.Close()
	for _, id := range ids {
		if _, err = io.WriteString(idfile, id+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func (r fileRepository) Ids() ([]string, error) {
	file, err := r.fs.Open(idFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	ids := []string{}
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}
	return ids, scanner.Err()
}

func (r fileRepository) Open(key string, passphrase []byte) (io.ReadCloser, error) {
	ids, err := r.Ids()
	if err != nil {
		return nil, err
	}
	el, err := r.entities.SecureKeyRing(ids)
	if err != nil {
		return nil, err
	}
	ciphertext, err := r.fs.Open(key + fileExtension)
	if err != nil {
		return nil, err
	}
	return ReadEncrypted(ciphertext, el, passphrase)
}

func (r fileRepository) Line(key string, passphrase []byte) (string, error) {
	plaintext, err := r.Open(key, passphrase)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	scanner := bufio.NewScanner(plaintext)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func (r fileRepository) Map(key string, passphrase []byte) (map[string]string, error) {
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

func (r fileRepository) Create(key string) (io.WriteCloser, error) {
	ids, err := r.Ids()
	if err != nil {
		return nil, err
	}
	el, err := r.entities.PublicKeyRing(ids)
	if err != nil {
		return nil, err
	}
	if err := rwvfs.MkdirAll(r.fs, filepath.Dir(key)); err != nil {
		return nil, err
	}
	ciphertext, err := r.fs.Create(key + fileExtension)
	if err != nil {
		return nil, err
	}
	return WriteEncrypted(ciphertext, el)
}

func (r fileRepository) Remove(key string) error {
	return r.fs.Remove(key + fileExtension)
}

func (r fileRepository) Walk(walkFn func(file string)) error {
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
