package filehub

import (
	"time"
)

type FileInfo interface {
	Path() string
	Size() int64
	ModTime() time.Time
	Filehub() Filehub
}

// type WalkFunc func(path string, info FileInfo, err error) error

// filehub
type Filehub interface {
	List(path string) (fs []FileInfo, err error)
	Exists(path string) (exists bool, err error)
	Get(path string) (data []byte, contType string, err error)
	Put(path string, data []byte, contType string) (err error)
	Del(path string) error
}
