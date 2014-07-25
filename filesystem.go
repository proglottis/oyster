package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

const (
	filePermission = 0600
	dirPermission  = 0700
)

type FileSystem interface {
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Remove(name string) error
	Walk(walkFn filepath.WalkFunc) error
}

type directoryFS struct {
	root string
}

func NewDirectory(root string) FileSystem {
	return &directoryFS{root: root}
}

func (fs directoryFS) Open(name string) (io.ReadCloser, error) {
	return os.Open(path.Join(fs.root, name))
}

func (fs directoryFS) Create(name string) (io.WriteCloser, error) {
	filepath := path.Join(fs.root, name)
	if err := os.MkdirAll(path.Dir(filepath), dirPermission); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission)
}

func (fs directoryFS) Remove(name string) error {
	filepath := path.Join(fs.root, name)
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

func (fs directoryFS) Walk(walkFn filepath.WalkFunc) error {
	return filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		path, _ = filepath.Rel(fs.root, path)
		return walkFn(path, info, err)
	})
}
