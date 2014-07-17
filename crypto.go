package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strings"
)

import (
	"code.google.com/p/go.crypto/openpgp"
)

func gpgHome() string {
	home := os.Getenv("GNUPGHOME")
	if home == "" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		home = path.Join(user.HomeDir, ".gnupg")
	}
	return home
}

func SecureKeyRingName() string {
	return path.Join(gpgHome(), "secring.gpg")
}

func PublicKeyRingName() string {
	return path.Join(gpgHome(), "pubring.gpg")
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

func OpenEncrypted(name string, el openpgp.EntityList, passphrase []byte) (io.ReadCloser, error) {
	ciphertext, err := os.Open(name)
	if err != nil {
		return nil, err
	}
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
	ciphertext io.ReadCloser
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

func CreateEncrypted(name string, perm os.FileMode, el openpgp.EntityList) (io.WriteCloser, error) {
	ciphertext, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return nil, err
	}
	plaintext, err := openpgp.Encrypt(ciphertext, el, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return &encryptedWriter{ciphertext, plaintext}, nil
}
