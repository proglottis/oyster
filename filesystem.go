package main

import (
	"path"

	"github.com/proglottis/rwvfs"
)

type walkableFileSystem struct {
	rwvfs.FileSystem
}

func (_ walkableFileSystem) Join(elem ...string) string {
	return path.Join(elem...)
}

func Walkable(fs rwvfs.FileSystem) rwvfs.WalkableFileSystem {
	return walkableFileSystem{fs}
}
