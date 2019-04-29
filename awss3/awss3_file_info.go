package awss3

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/wzshiming/filehub"
)

type AwsS3FileInfo struct {
	key     *s3.Object
	filehub filehub.Filehub
}

func (a *AwsS3FileInfo) Path() string {
	return *a.key.Key
}

func (a *AwsS3FileInfo) Size() int64 {
	return *a.key.Size
}

func (a *AwsS3FileInfo) ModTime() time.Time {
	return *a.key.LastModified
}

func (a *AwsS3FileInfo) Filehub() filehub.Filehub {
	return a.filehub
}

func (a *AwsS3FileInfo) String() string {
	return fmt.Sprintf("%s %s %d", a.Path(), a.ModTime(), a.Size())
}
