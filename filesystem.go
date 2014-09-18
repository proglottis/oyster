package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/kr/fs"
)

const (
	filePermission = 0600
	dirPermission  = 0700
)

type FileSystem interface {
	fs.FileSystem
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Remove(name string) error
}

type directoryFS struct {
	root string
}

func NewDirectory(root string) FileSystem {
	return &directoryFS{root: root}
}

func (fs directoryFS) resolve(name string) string {
	name = path.Clean("/" + name)
	return path.Join(fs.root, name)
}

func (fs directoryFS) Open(name string) (io.ReadCloser, error) {
	return os.Open(fs.resolve(name))
}

func (fs directoryFS) Create(name string) (io.WriteCloser, error) {
	filepath := fs.resolve(name)
	if err := os.MkdirAll(path.Dir(filepath), dirPermission); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission)
}

func (fs directoryFS) Remove(name string) error {
	filepath := fs.resolve(name)
	if err := os.Remove(filepath); err != nil {
		return err
	}
	if dirpath := path.Dir(filepath); dirpath != fs.root && dirEmpty(dirpath) {
		if err := os.Remove(dirpath); err != nil {
			return err
		}
	}
	return nil
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

func (fs directoryFS) ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(fs.resolve(name))
}

func (fs directoryFS) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(fs.resolve(name))
}

func (fs directoryFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}
