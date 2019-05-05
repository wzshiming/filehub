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
	Put(path string, data []byte, contType string) (p string, err error)
	PutExpire(path string, data []byte, conType string, expire time.Duration) (p string, err error)
	Del(path string) error
	Prefix() (string, error)
}
