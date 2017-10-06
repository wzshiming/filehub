package file

import (
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/wzshiming/filehub"
)

type File string

// file:///{path} or ./{path}
func NewFile(root string) (filehub.Filehub, error) {
	return File(root), nil
}

func (f File) RelPath(path string) string {
	return filepath.Join(string(f), filepath.Clean(path))
}

func (f File) List(path string) (fs []filehub.FileInfo, err error) {
	rp := f.RelPath(path)
	err = filepath.Walk(rp, func(path string, info os.FileInfo, err error) error {
		path, err = filepath.Rel(string(f), path)
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}
		path = strings.Replace(path, `\`, `/`, -1)
		fs = append(fs, &FileInfo{
			path:     path,
			FileInfo: info,
			filehub:  f,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

func (f File) Put(path string, data []byte, contType string) (err error) {
	rp := f.RelPath(path)
	err = os.MkdirAll(filepath.Dir(rp), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(rp, data, os.ModePerm)
}

func (f File) Get(path string) (data []byte, contType string, err error) {
	rp := f.RelPath(path)
	data, err = ioutil.ReadFile(rp)
	contType = mime.TypeByExtension(filepath.Ext(rp))
	return
}

func (f File) Exists(path string) (exists bool, err error) {
	rp := f.RelPath(path)
	stat, err := os.Stat(rp)
	if err != nil {
		return true, err
	}
	return stat != nil, nil
}

func (f File) Del(path string) error {
	rp := f.RelPath(path)
	return os.Remove(rp)
}
