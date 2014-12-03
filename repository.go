package main

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"sort"

	"github.com/kr/fs"
	"github.com/sourcegraph/rwvfs"
)

const (
	idFilename    = ".gpg-id"
	fileExtension = ".gpg"
)

var (
	ErrNotFound = errors.New("Not found")
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
	Fields FieldSlice `json:"fields,omitempty"`
}

type FieldSlice []Field

func (p FieldSlice) Len() int           { return len(p) }
func (p FieldSlice) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p FieldSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
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
		return nil, ErrNotFound
	}
	form := Form{
		FormRequest: *request,
		Fields:      make([]Field, 0, len(fileinfos)),
	}
	for _, fileinfo := range fileinfos {
		var err error
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		field := Field{Name: filename[:len(filename)-len(fileExtension)]}
		field.Value, err = r.getField(request.Key, field.Name, passphrase)
		if err != nil {
			return nil, err
		}
		form.Fields = append(form.Fields, field)
	}
	sort.Sort(form.Fields)
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
		return nil, ErrNotFound
	}
	form := Form{
		FormRequest: *request,
		Fields:      make([]Field, 0, len(fileinfos)),
	}
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		field := Field{Name: filename[:len(filename)-len(fileExtension)]}
		form.Fields = append(form.Fields, field)
	}
	sort.Sort(form.Fields)
	return &form, nil
}

func (r *FormRepo) Put(form *Form) error {
	if err := form.ParseUrl(); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(r.fs, form.Key); err != nil {
		return err
	}
	for _, field := range form.Fields {
		if err := r.putField(form.Key, field); err != nil {
			return err
		}
	}
	return nil
}

func (r *FormRepo) putField(key string, field Field) error {
	plaintext, err := r.fs.CreateEncrypted(r.fs.Join(key, field.Name+fileExtension))
	if err != nil {
		return err
	}
	defer plaintext.Close()
	if _, err := io.WriteString(plaintext, field.Value); err != nil {
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
