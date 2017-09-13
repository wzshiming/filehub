package file

import (
	"fmt"
	"os"

	"github.com/wzshiming/filehub"
)

type FileInfo struct {
	path string
	os.FileInfo
	filehub filehub.Filehub
}

func (a *FileInfo) Path() string {
	return a.path
}

func (a *FileInfo) Filehub() filehub.Filehub {
	return a.filehub
}

func (a *FileInfo) String() string {
	return fmt.Sprintf("%s %s %d", a.Path(), a.ModTime(), a.Size())
}
