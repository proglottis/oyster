package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"code.google.com/p/go.crypto/openpgp"
)

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
