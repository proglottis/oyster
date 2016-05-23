package oyster

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/sourcegraph/rwvfs"
	"golang.org/x/crypto/openpgp"
)

var (
	ErrCannotDecryptKey = errors.New("Cannot decrypt key")
	ErrNoMatchingKeys   = errors.New("No matching keys")
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
			if err := key.PrivateKey.Decrypt(passphrase); err != nil {
				return nil, ErrCannotDecryptKey
			}
			return nil, nil
		}
		return nil, ErrNoMatchingKeys
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

type GpgEntityRepo struct {
	root string
}

func NewGpgRepo(root string) GpgEntityRepo {
	return GpgEntityRepo{root: root}
}

func (r GpgEntityRepo) DefaultKeys() ([]string, error) {
	keyring, err := ReadKeyRing(path.Join(r.root, "secring.gpg"))
	if err != nil {
		return nil, err
	}
	if len(keyring) > 1 {
		return nil, errors.New("Ambiguous default key. Run `oyster init <your GPG key ID or email>`")
	}
	var ids []string
	for _, entity := range keyring {
		for _, key := range entity.Subkeys {
			ids = append(ids, key.PublicKey.KeyIdShortString())
		}
	}
	return ids, nil
}

func (r GpgEntityRepo) SecureKeyRing(ids []string) (openpgp.EntityList, error) {
	return EntitiesFromKeyRing(path.Join(r.root, "secring.gpg"), ids)
}

func (r GpgEntityRepo) PublicKeyRing(ids []string) (openpgp.EntityList, error) {
	return EntitiesFromKeyRing(path.Join(r.root, "pubring.gpg"), ids)
}

type Callback func() []byte

type CryptoFS struct {
	rwvfs.FileSystem
	Callback Callback
	entities GpgEntityRepo
}

func NewCryptoFS(fs rwvfs.FileSystem, entities GpgEntityRepo) *CryptoFS {
	return &CryptoFS{FileSystem: fs, entities: entities}
}

func (fs CryptoFS) Identities() ([]string, error) {
	f, err := fs.Open(idFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return fs.entities.DefaultKeys()
		}
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

func (fs CryptoFS) OpenEncrypted(name string) (io.ReadCloser, error) {
	ciphertext, err := fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
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
	return ReadEncrypted(ciphertext, el, fs.Callback())
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
