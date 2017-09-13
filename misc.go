package filehub

import (
	"fmt"
	"sort"
	"time"
)

// Copy 移动dst 没有或者不是最新的 到src中查找
func Copy(dst, src Filehub) error {
	dds, err := DiffHub(dst, src)
	if err != nil {
		return err
	}

	for _, v := range dds {
		if v.Src != nil {
			if v.Dst == nil || v.Src.ModTime().Add(time.Second).After(v.Dst.ModTime()) {
				p := v.Src.Path()
				d, t, err := src.Get(p)
				if err != nil {
					return err
				}
				dst.Put(p, d, t)
			}
		}
	}

	return nil
}

// DiffHub 比较
func DiffHub(dst, src Filehub) ([]*DiffInfo, error) {
	fd, err := dst.List("")
	if err != nil {
		return nil, err
	}

	fs, err := src.List("")
	if err != nil {
		return nil, err
	}

	return Diff(fd, fs), nil
}

// Diff 比较文件名差异
func Diff(fd, fs []FileInfo) (dds []*DiffInfo) {
	sort.SliceStable(fd, func(i, j int) bool {
		return fd[i].Path() < fd[j].Path()
	})
	sort.SliceStable(fs, func(i, j int) bool {
		return fs[i].Path() < fs[j].Path()
	})

	si := 0
	di := 0

loop:
	for ; di != len(fd); di++ {
		for {
			if si == len(fs) {
				break loop
			}
			vd := fd[di]
			vs := fs[si]
			if vd.Path() < vs.Path() {
				vs = nil
			} else if vd.Path() > vs.Path() {
				vd = nil
			}
			dds = append(dds, &DiffInfo{
				Dst: vd,
				Src: vs,
			})
			if vs != nil {
				si++
			}
			if vd != nil {
				continue loop
			}
		}
	}

	for ; si != len(fs); si++ {
		vs := fs[si]
		dds = append(dds, &DiffInfo{
			Dst: nil,
			Src: vs,
		})
	}

	for ; di != len(fd); di++ {
		vd := fd[di]
		dds = append(dds, &DiffInfo{
			Dst: vd,
			Src: nil,
		})
	}
	return dds
}

type DiffInfo struct {
	Dst FileInfo
	Src FileInfo
}

func (d *DiffInfo) String() string {
	return fmt.Sprintf("%v -=- %v", d.Dst, d.Src)
}
