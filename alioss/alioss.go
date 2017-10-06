package alioss

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/wzshiming/filehub"
)

type AliOss struct {
	buc *oss.Bucket
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

	return &AliOss{
		buc: buc,
	}, nil
}

func (a *AliOss) List(path string) (fs []filehub.FileInfo, err error) {
	marker := oss.Marker("")
	pre := oss.Prefix(path)
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

func (a *AliOss) Put(path string, data []byte, contType string) (err error) {
	return a.buc.PutObject(path, bytes.NewReader(data))
}

func (a *AliOss) Get(path string) (data []byte, contType string, err error) {
	resp, err := a.buc.GetObject(path)
	if err != nil {
		return nil, "", err
	}

	data, err = ioutil.ReadAll(resp)
	if err != nil {
		return nil, "", err
	}
	defer resp.Close()
	return
}

func (a *AliOss) Exists(path string) (exists bool, err error) {
	return a.buc.IsObjectExist(path)
}

func (a *AliOss) Del(path string) error {
	return a.buc.DeleteObject(path)
}
