package cryptofs

import (
	"io"

	"github.com/pkg/errors"
	"github.com/proglottis/oyster/config"
	"github.com/sourcegraph/rwvfs"
)

var (
	ErrNotFound = errors.New("Not found")
)

type Callback func() []byte

type CryptoFS interface {
	rwvfs.FileSystem

	Join(elem ...string) string

	CheckIdentities(ids []string) error
	SetIdentities(ids []string) error
	Identities() ([]string, error)

	OpenEncrypted(key string) (io.ReadCloser, error)
	CreateEncrypted(key string) (io.WriteCloser, error)
	SetCallback(cb Callback)
}

type NewFSFunc func(fs rwvfs.FileSystem, config *config.Config) CryptoFS

var fsFuncs = make(map[string]NewFSFunc)

func Register(name string, f NewFSFunc) {
	fsFuncs[name] = f
}

func New(name string, fs rwvfs.FileSystem, config *config.Config) (CryptoFS, error) {
	f, ok := fsFuncs[name]
	if !ok {
		return nil, errors.New("No such crypto engine")
	}
	return f(fs, config), nil
}
