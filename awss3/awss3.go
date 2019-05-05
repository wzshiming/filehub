package awss3

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/wzshiming/filehub"
)

type AwsS3 struct {
	s3     *s3.S3
	prefix string
	bucket string
	path   string
}

// awss3://{accessKeyId}:{accessKeySecret}@s3.{bucket}.amazonaws.com/{bucket}/{path}
func NewAwsS3(remote string) (*AwsS3, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return nil, err
	}

	accessKeyId := ""
	accessKeySecret := ""
	if u.User != nil {
		accessKeyId = u.User.Username()
		accessKeySecret, _ = u.User.Password()
	}
	region := strings.TrimSuffix(strings.TrimPrefix(u.Host, "s3."), ".amazonaws.com")
	bucket := ""
	path := ""
	bucketAndPath := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	bucket = bucketAndPath[0]
	if len(bucketAndPath) > 1 {
		path = bucketAndPath[1]
	}

	config := &aws.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
	}
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	return &AwsS3{
		prefix: `https://` + u.Host + "/" + strings.TrimSuffix(bucket+"/"+path, "/"),
		s3:     svc,
		bucket: bucket,
		path:   path,
	}, nil
}

func (a *AwsS3) List(pat string) (fs []filehub.FileInfo, err error) {
	var pi *string
	pat = path.Join(a.path, pat)
	if pat != "" {
		pi = &pat
	}
	listObjectsV2Output, err := a.s3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &a.bucket,
		Prefix: pi,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range listObjectsV2Output.Contents {
		fs = append(fs, &AwsS3FileInfo{
			key:     v,
			filehub: a,
		})
	}
	return fs, nil
}

func (a *AwsS3) Put(pat string, data []byte, contentType string) (p string, err error) {
	var pi *string
	pat = path.Join(a.path, pat)
	if pat != "" {
		pi = &pat
	}
	_, err = a.s3.PutObject(&s3.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         pi,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}
	return pat, nil
}

func (a *AwsS3) PutExpire(pat string, data []byte, contentType string, dur time.Duration) (p string, err error) {
	var pi *string
	pat = path.Join(a.path, pat)
	if pat != "" {
		pi = &pat
	}
	expires := time.Now().Add(dur)
	_, err = a.s3.PutObject(&s3.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         pi,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		Expires:     &expires,
	})
	if err != nil {
		return "", err
	}
	return pat, nil
}

func (a *AwsS3) Get(pat string) (data []byte, contentType string, err error) {
	pat = path.Join(a.path, pat)
	getObjectOutput, err := a.s3.GetObject(&s3.GetObjectInput{
		Bucket: &a.bucket,
		Key:    &pat,
	})
	if err != nil {
		return nil, "", err
	}
	body, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, "", err
	}
	getObjectOutput.Body.Close()
	if getObjectOutput.ContentType != nil {
		contentType = *getObjectOutput.ContentType
	}
	return body, contentType, nil
}

func (a *AwsS3) Exists(pat string) (exists bool, err error) {
	pat = path.Join(a.path, pat)
	getObjectOutput, err := a.s3.GetObject(&s3.GetObjectInput{
		Bucket: &a.bucket,
		Key:    &pat,
	})
	if err != nil {
		return false, err
	}
	getObjectOutput.Body.Close()
	return true, nil
}

func (a *AwsS3) Del(pat string) error {
	pat = path.Join(a.path, pat)
	_, err := a.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &a.bucket,
		Key:    &pat,
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *AwsS3) Prefix() (string, error) {
	return a.prefix, nil
}
