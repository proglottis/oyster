package main

import (
	"bufio"
	"io"
	"net/url"
	"path/filepath"

	"github.com/kr/fs"
	"github.com/proglottis/rwvfs"
)

const (
	idFilename    = ".gpg-id"
	fileExtension = ".gpg"
)

func InitRepo(fs *CryptoFS, ids []string) error {
	if err := fs.CheckIdentities(ids); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(fs, "/"); err != nil {
		return err
	}
	return fs.SetIdentities(ids)
}

type FormRequest struct {
	Key string `json:"key"`
	Url string `json:"url,omitempty"`
}

func (f *FormRequest) ParseUrl() error {
	if len(f.Url) > 0 {
		keyurl, err := url.Parse(f.Url)
		if err != nil {
			return err
		}
		f.Key = keyurl.Host + keyurl.Path
	}
	return nil
}

type Form struct {
	FormRequest
	Fields map[string]string `json:"fields,omitempty"`
}

type FormRepo struct {
	fs *CryptoFS
}

func NewFormRepo(fs *CryptoFS) *FormRepo {
	return &FormRepo{fs: fs}
}

func (r *FormRepo) Get(request *FormRequest, passphrase []byte) (*Form, error) {
	if err := request.ParseUrl(); err != nil {
		return nil, err
	}
	fileinfos, err := r.fs.ReadDir(request.Key)
	if err != nil {
		return nil, err
	}
	form := Form{
		FormRequest: *request,
		Fields:      make(map[string]string),
	}
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		name := filename[:len(filename)-len(fileExtension)]
		value, err := r.getField(request.Key, name, passphrase)
		if err != nil {
			return nil, err
		}
		form.Fields[name] = value
	}
	return &form, nil
}

func (r *FormRepo) getField(key, name string, passphrase []byte) (string, error) {
	plaintext, err := r.fs.OpenEncrypted(r.fs.Join(key, name+fileExtension), passphrase)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	return readline(plaintext)
}

func (r *FormRepo) Fields(request *FormRequest) (*Form, error) {
	if err := request.ParseUrl(); err != nil {
		return nil, err
	}
	fileinfos, err := r.fs.ReadDir(request.Key)
	if err != nil {
		return nil, err
	}
	form := Form{
		FormRequest: *request,
		Fields:      make(map[string]string),
	}
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		name := filename[:len(filename)-len(fileExtension)]
		form.Fields[name] = ""
	}
	return &form, nil
}

func (r *FormRepo) Put(form *Form) error {
	if err := form.ParseUrl(); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(r.fs, form.Key); err != nil {
		return err
	}
	for field, value := range form.Fields {
		if err := r.putField(form.Key, field, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *FormRepo) putField(key, name, value string) error {
	plaintext, err := r.fs.CreateEncrypted(r.fs.Join(key, name+fileExtension))
	if err != nil {
		return err
	}
	defer plaintext.Close()
	if _, err := io.WriteString(plaintext, value); err != nil {
		return err
	}
	return nil
}

type FileRepo struct {
	fs *CryptoFS
}

func NewFileRepo(fs *CryptoFS) *FileRepo {
	return &FileRepo{fs: fs}
}

func (r *FileRepo) Open(key string, passphrase []byte) (io.ReadCloser, error) {
	return r.fs.OpenEncrypted(key+fileExtension, passphrase)
}

func (r *FileRepo) Line(key string, passphrase []byte) (string, error) {
	plaintext, err := r.Open(key, passphrase)
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	return readline(plaintext)
}

func (r *FileRepo) Create(key string) (io.WriteCloser, error) {
	if err := rwvfs.MkdirAll(r.fs, filepath.Dir(key)); err != nil {
		return nil, err
	}
	return r.fs.CreateEncrypted(key + fileExtension)
}

func (r *FileRepo) Remove(key string) error {
	return r.fs.Remove(key + fileExtension)
}

func (r *FileRepo) Walk(walkFn func(file string)) error {
	walker := fs.WalkFS(".", r.fs)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		path := walker.Path()
		if walker.Stat().IsDir() || filepath.Ext(path) != fileExtension {
			continue
		}
		walkFn(path[:len(path)-len(fileExtension)])
	}
	return nil
}

func readline(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
