package alioss

import (
	"bytes"
	"errors"
	"io/ioutil"
	"mime"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/wzshiming/filehub"
)

type AliOss struct {
	buc    *oss.Bucket
	prefix string
	path   string
}

// alioss://{accessKeyId}:{accessKeySecret}@{bucket}.{endpoint}
func NewAliOss(remote string) (filehub.Filehub, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return nil, err
	}

	user := u.User
	if user == nil {
		return nil, errors.New("alioss: Invalid Access")
	}
	pwd, _ := user.Password()
	uname := user.Username()

	si := strings.Index(u.Host, ".")
	if si == -1 {
		return nil, errors.New("alioss: Invalid Host")
	}
	bucs := u.Host[:si]
	endpoint := u.Host[si+1:]

	cli, err := oss.New(`https://`+endpoint, uname, pwd)
	if err != nil {
		return nil, err
	}

	buc, err := cli.Bucket(bucs)
	if err != nil {
		return nil, err
	}

	pat := path.Clean(u.Path)
	pat = strings.TrimPrefix(pat, "/")
	return &AliOss{
		prefix: `https://` + u.Host + "/" + pat,
		buc:    buc,
		path:   pat,
	}, nil
}

func (a *AliOss) List(pat string) (fs []filehub.FileInfo, err error) {
	marker := oss.Marker("")
	pat = path.Join(a.path, pat)
	pre := oss.Prefix(pat)
	maxkey := oss.MaxKeys(1000)
	for {
		lsRes, err := a.buc.ListObjects(pre, maxkey, marker)
		if err != nil {
			return nil, err
		}

		for _, v := range lsRes.Objects {
			fs = append(fs, &AliOssFileInfo{
				key:     v,
				filehub: a,
			})
		}

		marker = oss.Marker(lsRes.NextMarker)

		if !lsRes.IsTruncated {
			break
		}
	}

	return fs, nil
}

func (a *AliOss) Put(pat string, data []byte, contentType string) (p string, err error) {
	pat = path.Join(a.path, pat)
	return pat, a.buc.PutObject(pat, bytes.NewReader(data))
}

func (a *AliOss) PutExpire(pat string, data []byte, contentType string, dur time.Duration) (p string, err error) {
	pat = path.Join(a.path, pat)
	signUrl, err := a.buc.SignURL(pat, oss.HTTPPut, int64(dur/time.Second))
	if err != nil {
		return "", err
	}
	return signUrl, a.buc.PutObjectWithURL(signUrl, bytes.NewReader(data))
}

func (a *AliOss) Get(pat string) (data []byte, contentType string, err error) {
	pat = path.Join(a.path, pat)
	resp, err := a.buc.GetObject(pat)
	if err != nil {
		return nil, "", err
	}

	data, err = ioutil.ReadAll(resp)
	if err != nil {
		return nil, "", err
	}
	defer resp.Close()

	contentType = mime.TypeByExtension(filepath.Ext(pat))
	return
}

func (a *AliOss) Exists(pat string) (exists bool, err error) {
	pat = path.Join(a.path, pat)
	return a.buc.IsObjectExist(pat)
}

func (a *AliOss) Del(pat string) error {
	pat = path.Join(a.path, pat)
	return a.buc.DeleteObject(pat)
}

func (a *AliOss) Prefix() (string, error) {
	return a.prefix, nil
}
