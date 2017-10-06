package alioss

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/wzshiming/filehub"
)

type AliOssFileInfo struct {
	key     oss.ObjectProperties
	filehub filehub.Filehub
}

func (a *AliOssFileInfo) Path() string {
	return a.key.Key
}

func (a *AliOssFileInfo) Size() int64 {
	return a.key.Size
}

func (a *AliOssFileInfo) ModTime() time.Time {
	return a.key.LastModified
}

func (a *AliOssFileInfo) Filehub() filehub.Filehub {
	return a.filehub
}

func (a *AliOssFileInfo) String() string {
	return fmt.Sprintf("%s %s %d", a.Path(), a.ModTime(), a.Size())
}
