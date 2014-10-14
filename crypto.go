package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"code.google.com/p/go.crypto/openpgp"
	"github.com/proglottis/rwvfs"
)

func EntityMatchesId(entity *openpgp.Entity, id string) bool {
	for _, identity := range entity.Identities {
		if identity.UserId.Email == id {
			return true
		}
	}
	id = strings.ToUpper(id)
	for _, key := range entity.Subkeys {
		if key.PublicKey.KeyIdShortString() == id || key.PublicKey.KeyIdString() == id {
			return true
		}
	}
	return entity.PrimaryKey.KeyIdShortString() == id || entity.PrimaryKey.KeyIdString() == id
}

func EntityMatchesAnyId(entity *openpgp.Entity, ids []string) bool {
	for _, id := range ids {
		if EntityMatchesId(entity, id) {
			return true
		}
	}
	return false
}

func IdMatchesAnyEntity(id string, el openpgp.EntityList) bool {
	for _, entity := range el {
		if EntityMatchesId(entity, id) {
			return true
		}
	}
	return false
}

func ReadKeyRing(keyRingName string) (openpgp.EntityList, error) {
	keyfile, err := os.Open(keyRingName)
	if err != nil {
		return nil, err
	}
	defer keyfile.Close()
	return openpgp.ReadKeyRing(keyfile)
}

func EntitiesFromKeyRing(keyRingName string, ids []string) (openpgp.EntityList, error) {
	keyring, err := ReadKeyRing(keyRingName)
	if err != nil {
		return nil, err
	}
	el := openpgp.EntityList{}
	for _, entity := range keyring {
		if EntityMatchesAnyId(entity, ids) {
			el = append(el, entity)
		}
	}
	return el, nil
}

type encryptedReader struct {
	ciphertext io.ReadCloser
	plaintext  io.Reader
}

func (f encryptedReader) Read(p []byte) (int, error) {
	return f.plaintext.Read(p)
}

func (f encryptedReader) Close() error {
	return f.ciphertext.Close()
}

func ReadEncrypted(ciphertext io.ReadCloser, el openpgp.EntityList, passphrase []byte) (io.ReadCloser, error) {
	md, err := openpgp.ReadMessage(ciphertext, el, func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if symmetric {
			return nil, fmt.Errorf("No support for symmetrical encryption")
		}
		for _, key := range keys {
			return nil, key.PrivateKey.Decrypt(passphrase)
		}
		return nil, fmt.Errorf("No matching keys")
	}, nil)
	if err != nil {
		return nil, err
	}
	return &encryptedReader{ciphertext, md.UnverifiedBody}, nil
}

type encryptedWriter struct {
	ciphertext io.Closer
	plaintext  io.WriteCloser
}

func (w encryptedWriter) Write(p []byte) (int, error) {
	return w.plaintext.Write(p)
}

func (w encryptedWriter) Close() error {
	w.plaintext.Close()
	w.ciphertext.Close()
	return nil
}

func WriteEncrypted(ciphertext io.WriteCloser, el openpgp.EntityList) (io.WriteCloser, error) {
	plaintext, err := openpgp.Encrypt(ciphertext, el, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &encryptedWriter{ciphertext, plaintext}, nil
}

type EntityRepo interface {
	SecureKeyRing(ids []string) (openpgp.EntityList, error)
	PublicKeyRing(ids []string) (openpgp.EntityList, error)
}

type gpgRepo struct {
	root string
}

func NewGpgRepo(root string) EntityRepo {
	return gpgRepo{root: root}
}

func (r gpgRepo) SecureKeyRing(ids []string) (openpgp.EntityList, error) {
	return EntitiesFromKeyRing(path.Join(r.root, "secring.gpg"), ids)
}

func (r gpgRepo) PublicKeyRing(ids []string) (openpgp.EntityList, error) {
	return EntitiesFromKeyRing(path.Join(r.root, "pubring.gpg"), ids)
}

type CryptoFS struct {
	rwvfs.FileSystem
	entities EntityRepo
}

func NewCryptoFS(fs rwvfs.FileSystem, entities EntityRepo) *CryptoFS {
	return &CryptoFS{FileSystem: fs, entities: entities}
}

func (fs CryptoFS) Identities() ([]string, error) {
	f, err := fs.Open(idFilename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	ids := []string{}
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}
	return ids, scanner.Err()
}

func (fs CryptoFS) CheckIdentities(ids []string) error {
	el, err := fs.entities.PublicKeyRing(ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !IdMatchesAnyEntity(id, el) {
			return fmt.Errorf("No matching public key %s", id)
		}
	}
	el, err = fs.entities.SecureKeyRing(ids)
	if err != nil {
		return err
	}
	if len(el) < 1 {
		return fmt.Errorf("No matching secure keys")
	}
	return nil
}

func (fs CryptoFS) SetIdentities(ids []string) error {
	f, err := fs.Create(idFilename)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, id := range ids {
		if _, err = io.WriteString(f, id+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func (fs CryptoFS) OpenEncrypted(name string, passphrase []byte) (io.ReadCloser, error) {
	ciphertext, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	ids, err := fs.Identities()
	if err != nil {
		return nil, err
	}
	el, err := fs.entities.SecureKeyRing(ids)
	if err != nil {
		return nil, err
	}
	return ReadEncrypted(ciphertext, el, passphrase)
}

func (fs CryptoFS) CreateEncrypted(name string) (io.WriteCloser, error) {
	ciphertext, err := fs.Create(name)
	if err != nil {
		return nil, err
	}
	ids, err := fs.Identities()
	if err != nil {
		return nil, err
	}
	el, err := fs.entities.PublicKeyRing(ids)
	if err != nil {
		return nil, err
	}
	return WriteEncrypted(ciphertext, el)
}

func (fs CryptoFS) Join(elem ...string) string {
	return path.Join(elem...)
}
