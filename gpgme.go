// +build gpgme

package oyster

import (
	"bufio"
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/proglottis/gpgme"
	"github.com/sourcegraph/rwvfs"
)

type gpgmeFS struct {
	rwvfs.FileSystem
	cb Callback
}

func NewCryptoFS(fs rwvfs.FileSystem, config *Config) CryptoFS {
	return &gpgmeFS{FileSystem: fs}
}

func (fs *gpgmeFS) SetCallback(cb Callback) {
	fs.cb = cb
}

func (fs *gpgmeFS) CheckIdentities(ids []string) error {
	return nil
}

func (fs *gpgmeFS) Identities() ([]string, error) {
	f, err := fs.Open(idFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("No selected key. Run `oyster init <your GPG key ID or email>`")
		}
		return nil, errors.Wrap(err, "open identities")
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	ids := []string{}
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}
	return ids, errors.Wrap(scanner.Err(), "scan identities")
}

func (fs *gpgmeFS) SetIdentities(ids []string) error {
	f, err := fs.Create(idFilename)
	if err != nil {
		return errors.Wrap(err, "create identities")
	}
	defer f.Close()
	for _, id := range ids {
		if _, err = io.WriteString(f, id+"\n"); err != nil {
			return errors.Wrap(err, "write identities")
		}
	}
	return nil
}

func (fs *gpgmeFS) keys(secureOnly bool) ([]*gpgme.Key, error) {
	ids, err := fs.Identities()
	if err != nil {
		return nil, errors.Wrap(err, "identities")
	}
	var keys []*gpgme.Key
	for _, id := range ids {
		ks, err := gpgme.FindKeys(id, secureOnly)
		if err != nil {
			return nil, errors.Wrap(err, "finding keys")
		}
		for _, k := range ks {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (fs *gpgmeFS) OpenEncrypted(name string) (io.ReadCloser, error) {
	ctx, err := gpgme.New()
	if err != nil {
		return nil, errors.Wrap(err, "context")
	}
	defer ctx.Release()
	if err := ctx.SetCallback(func(uidHint string, prevWasBad bool, f *os.File) error {
		_, err := f.Write(append(fs.cb(), '\n'))
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "callback")
	}
	f, err := fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "open file")
	}
	defer f.Close()
	cipher, err := gpgme.NewDataReader(f)
	if err != nil {
		return nil, errors.Wrap(err, "cipher data")
	}
	defer cipher.Close()
	plain, err := gpgme.NewData()
	if err != nil {
		return nil, errors.Wrap(err, "plain data")
	}
	err = ctx.Decrypt(cipher, plain)
	plain.Seek(0, gpgme.SeekSet)
	return plain, errors.Wrap(err, "decrypt")
}

type encryptOnClose struct {
	Ctx           *gpgme.Context
	File          io.Closer
	Plain, Cipher *gpgme.Data
	Keys          []*gpgme.Key
}

func (e encryptOnClose) Write(v []byte) (int, error) {
	return e.Plain.Write(v)
}

func (e encryptOnClose) Close() error {
	if _, err := e.Plain.Seek(0, gpgme.SeekSet); err != nil {
		return err
	}
	if err := e.Ctx.Encrypt(e.Keys, 0, e.Plain, e.Cipher); err != nil {
		return err
	}
	if err := e.Cipher.Close(); err != nil {
		return err
	}
	if err := e.Plain.Close(); err != nil {
		return err
	}
	return e.File.Close()
}

func (fs *gpgmeFS) CreateEncrypted(name string) (io.WriteCloser, error) {
	ctx, err := gpgme.New()
	if err != nil {
		return nil, errors.Wrap(err, "context")
	}
	f, err := fs.Create(name)
	if err != nil {
		return nil, errors.Wrap(err, "create file")
	}
	plain, err := gpgme.NewData()
	if err != nil {
		return nil, errors.Wrap(err, "plain data")
	}
	cipher, err := gpgme.NewDataWriter(f)
	if err != nil {
		return nil, errors.Wrap(err, "cipher data")
	}
	keys, err := fs.keys(false)
	if err != nil {
		return nil, errors.Wrap(err, "keys")
	}
	return encryptOnClose{Ctx: ctx, File: f, Plain: plain, Cipher: cipher, Keys: keys}, nil
}

func (fs *gpgmeFS) Join(elem ...string) string {
	return path.Join(elem...)
}
