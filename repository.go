package oyster

import (
	"bufio"
	"errors"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kr/fs"
	"github.com/sourcegraph/rwvfs"
)

const (
	idFilename    = ".gpg-id"
	fileExtension = ".gpg"
	hostSep       = "."
	pathSep       = "/"
)

var (
	ErrNotFound = errors.New("Not found")
)

func InitRepo(fs CryptoFS, ids []string) error {
	if err := fs.CheckIdentities(ids); err != nil {
		return err
	}
	if err := rwvfs.MkdirAll(fs, "/"); err != nil {
		return err
	}
	return fs.SetIdentities(ids)
}

type Form struct {
	Key    string     `json:"key"`
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

type FormRepo struct {
	fs CryptoFS
}

func NewFormRepo(fs CryptoFS) *FormRepo {
	return &FormRepo{fs: fs}
}

func (r *FormRepo) List() ([]Form, error) {
	forms := make([]Form, 0)
	walker := fs.WalkFS(".", r.fs)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return nil, err
		}
		if !walker.Stat().IsDir() {
			continue
		}
		form, err := r.Fields(walker.Path())
		switch err {
		case ErrNotFound: // Ignore
		case nil:
			if len(form.Fields) > 0 {
				forms = append(forms, *form)
			}
		default:
			return nil, err
		}
	}
	return forms, nil
}

func (r *FormRepo) Search(query string) ([]Form, error) {
	url, err := url.Parse(query)
	if err != nil {
		return nil, err
	}
	components := strings.Split(strings.Trim(url.Path, pathSep), pathSep)
	if components[0] == "" {
		components = components[1:]
	}
	domains := strings.Split(strings.Trim(url.Host, hostSep), hostSep)
	forms := make([]Form, 0, 8)
	for i := 0; i < len(components)+1; i++ {
		path := strings.Join(components[:len(components)-i], pathSep)
		for j := range domains {
			host := strings.Join(domains[j:], ".")
			key := strings.Trim(host+pathSep+path, pathSep)
			form, err := r.Fields(key)
			switch err {
			case ErrNotFound: // Ignore
			case nil:
				if len(form.Fields) > 0 {
					forms = append(forms, *form)
				}
			default:
				return nil, err
			}
		}
	}
	return forms, nil
}

func (r *FormRepo) Get(key string) (*Form, error) {
	fileinfos, err := r.fs.ReadDir(key)
	if err != nil {
		return nil, ErrNotFound
	}
	form := Form{
		Key:    key,
		Fields: make([]Field, 0, len(fileinfos)),
	}
	for _, fileinfo := range fileinfos {
		var err error
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		field := Field{Name: filename[:len(filename)-len(fileExtension)]}
		field.Value, err = r.getField(key, field.Name)
		if err != nil {
			return nil, err
		}
		form.Fields = append(form.Fields, field)
	}
	sort.Sort(form.Fields)
	return &form, nil
}

func (r *FormRepo) getField(key, name string) (string, error) {
	plaintext, err := r.fs.OpenEncrypted(r.fs.Join(key, name+fileExtension))
	if err != nil {
		return "", err
	}
	defer plaintext.Close()
	return readline(plaintext)
}

func (r *FormRepo) Fields(key string) (*Form, error) {
	fileinfos, err := r.fs.ReadDir(key)
	if err != nil {
		return nil, ErrNotFound
	}
	form := Form{
		Key:    key,
		Fields: make([]Field, 0, len(fileinfos)),
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

func (r *FormRepo) Remove(key string) error {
	fileinfos, err := r.fs.ReadDir(key)
	if err != nil {
		return ErrNotFound
	}
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if fileinfo.IsDir() || filepath.Ext(filename) != fileExtension {
			continue
		}
		if err := r.fs.Remove(r.fs.Join(key, filename)); err != nil {
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
	fs CryptoFS
}

func NewFileRepo(fs CryptoFS) *FileRepo {
	return &FileRepo{fs: fs}
}

func (r *FileRepo) Open(key string) (io.ReadCloser, error) {
	return r.fs.OpenEncrypted(key + fileExtension)
}

func (r *FileRepo) Line(key string) (string, error) {
	plaintext, err := r.Open(key)
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
