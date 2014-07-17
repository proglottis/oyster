package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

const (
	idFilename     = ".gpg-id"
	fileExtension  = ".gpg"
	filePermission = 0600
	dirPermission  = 0700
)

type Repository interface {
	Init(ids []string) error
	Get(key string, passphrase []byte) (io.ReadCloser, error)
	GetLine(key string, passphrase []byte) (string, error)
	Put(key string) (io.WriteCloser, error)
	Remove(key string) error
	Walk(walkFn func(file string)) error
}

type fileRepository struct {
	root string
}

func NewRepository(root string) Repository {
	return &fileRepository{
		root: root,
	}
}

func checkPublicKeyRingIds(ids []string) error {
	keyringname := PublicKeyRingName()
	el, err := ReadKeyRing(keyringname)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !IdMatchesAnyEntity(id, el) {
			return fmt.Errorf("No matching public key %s in %s", id, keyringname)
		}
	}
	return nil
}

func checkSecureKeyRingIds(ids []string) error {
	keyringname := SecureKeyRingName()
	el, err := EntitiesFromKeyRing(keyringname, ids)
	if err != nil {
		return err
	}
	if len(el) < 1 {
		return fmt.Errorf("No matching secure keys in %s", keyringname)
	}
	return nil
}

func (r fileRepository) Init(ids []string) error {
	if err := checkPublicKeyRingIds(ids); err != nil {
		return err
	}
	if err := checkSecureKeyRingIds(ids); err != nil {
		return err
	}
	if err := os.MkdirAll(r.root, dirPermission); err != nil {
		return err
	}
	idfile, err := os.OpenFile(path.Join(r.root, idFilename), os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission)
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
	file, err := os.Open(path.Join(r.root, idFilename))
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

func (r fileRepository) Get(key string, passphrase []byte) (io.ReadCloser, error) {
	ids, err := r.Ids()
	if err != nil {
		return nil, err
	}
	el, err := EntitiesFromKeyRing(SecureKeyRingName(), ids)
	if err != nil {
		return nil, err
	}
	return OpenEncrypted(path.Join(r.root, key+fileExtension), el, passphrase)
}

func (r fileRepository) GetLine(key string, passphrase []byte) (string, error) {
	plaintext, err := r.Get(key, passphrase)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	scanner := bufio.NewScanner(plaintext)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func (r fileRepository) Put(key string) (io.WriteCloser, error) {
	ids, err := r.Ids()
	if err != nil {
		return nil, err
	}
	el, err := EntitiesFromKeyRing(PublicKeyRingName(), ids)
	if err != nil {
		return nil, err
	}
	filepath := path.Join(r.root, key+fileExtension)
	if err := os.MkdirAll(path.Dir(filepath), dirPermission); err != nil {
		return nil, err
	}
	return CreateEncrypted(filepath, filePermission, el)
}

func dirEmpty(name string) bool {
	dir, err := os.Open(name)
	if err != nil {
		return false
	}
	defer dir.Close()
	_, err = dir.Readdir(1)
	return err == io.EOF
}

func (r fileRepository) Remove(key string) error {
	filepath := path.Join(r.root, key+fileExtension)
	err := os.Remove(filepath)
	if err != nil {
		return err
	}
	if dirpath := path.Dir(filepath); dirpath != r.root && dirEmpty(dirpath) {
		if err = os.Remove(dirpath); err != nil {
			return err
		}
	}
	return nil
}

func (r fileRepository) Walk(walkFn func(file string)) error {
	filepath.Walk(r.root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != fileExtension {
			return nil
		}
		path, _ = filepath.Rel(r.root, path)
		walkFn(path[:len(path)-len(fileExtension)])
		return nil
	})
	return nil
}
