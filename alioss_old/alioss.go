package alioss

import (
	"errors"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/denverdino/aliyungo/oss"
	"github.com/wzshiming/filehub"
)

type AliOss struct {
	cli *oss.Client
	buc string
	opt oss.Options
	acl oss.ACL
}

// alioss://{accessKeyId}:{accessKeySecret}@{bucket}.{Region}[-internal].aliyuncs.com
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
	ss := strings.Split(u.Host, ".")
	if len(ss) < 2 {
		return nil, errors.New("alioss: Invalid Host")
	}
	buc := ss[0]
	region := ss[1]
	region0 := strings.TrimSuffix(region, "-internal")
	return &AliOss{
		cli: oss.NewOSSClient(oss.Region(region0), region != region0, uname, pwd, true),
		buc: buc,
		acl: oss.PublicRead,
	}, nil
}

func (a *AliOss) List(path string) (fs []filehub.FileInfo, err error) {
	buc := a.cli.Bucket(a.buc)
	list, err := buc.List(path, "/", "", 1000)
	if err != nil {
		return nil, err
	}

	for _, v := range list.Contents {
		fs = append(fs, &AliOssFileInfo{
			key:     v,
			filehub: a,
		})
	}

	for _, v := range list.CommonPrefixes {
		fs0, err := a.List(v)
		if err != nil {
			return nil, err
		}
		fs = append(fs, fs0...)
	}

	return fs, nil
}

func (a *AliOss) Put(path string, data []byte, contType string) (err error) {
	buc := a.cli.Bucket(a.buc)
	return buc.Put(path, data, contType, a.acl, a.opt)
}

func (a *AliOss) Get(path string) (data []byte, contType string, err error) {
	buc := a.cli.Bucket(a.buc)
	resp, err := buc.GetResponse(path)
	if err != nil {
		return nil, "", err
	}
	if resp.Header != nil {
		contType = resp.Header.Get("Content-Type")
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	return
}

func (a *AliOss) Exists(path string) (exists bool, err error) {
	buc := a.cli.Bucket(a.buc)
	return buc.Exists(path)
}

func (a *AliOss) Del(path string) error {
	buc := a.cli.Bucket(a.buc)
	return buc.Del(path)
}
