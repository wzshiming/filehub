package filehub

import (
	"fmt"
	"sort"

	"github.com/wzshiming/fork"
	ffmt "gopkg.in/ffmt.v1"
)

type CopyType uint32

const (
	None    CopyType = 0
	_       CopyType = 1 << (32 - iota - 1)
	Create           // 如果源有目标没有的文件则在目标创建文件
	Replace          // 如果源比目标新就替换目标的文件
	Delete           // 如果目标有源没有的文件则删除目标的文件
)

func (c CopyType) Exists(ct CopyType) bool {
	return c&ct == ct
}

// Copy
func Copy(dst, src Filehub, path string, ct CopyType, forkSize int) error {
	if ct == None {
		return nil
	}

	dds, err := DiffHub(dst, src, path)
	if err != nil {
		return err
	}

	f := fork.NewFork(forkSize)

	// 如果源有目标没有的文件则在目标创建文件
	if ct.Exists(Create) {
		for _, v := range dds {
			vsrc := v.Src
			vdst := v.Dst
			if vsrc != nil && vdst == nil {
				f.Push(func() {
					p := vsrc.Path()
					d, t, err := src.Get(p)
					if err != nil {
						ffmt.Mark(err)
						return
					}

					p, err = dst.Put(p, d, t)
					if err != nil {
						ffmt.Mark(err)
						return
					}
				})
			}
		}
	}

	// 如果源比目标新就替换目标的文件
	if ct.Exists(Replace) {
		for _, v := range dds {
			vsrc := v.Src
			vdst := v.Dst
			if vsrc != nil && vdst != nil &&
				vsrc.ModTime().After(vdst.ModTime()) {
				f.Push(func() {
					p := vsrc.Path()
					d, t, err := src.Get(p)
					if err != nil {
						ffmt.Mark(err)
						return
					}

					p, err = dst.Put(p, d, t)
					if err != nil {
						ffmt.Mark(err)
						return
					}
				})
			}
		}
	}

	// 如果目标有源没有的文件则删除目标的文件
	if ct.Exists(Delete) {
		for _, v := range dds {
			vsrc := v.Src
			vdst := v.Dst
			if vsrc == nil && vdst != nil {
				f.Push(func() {
					err := dst.Del(vdst.Path())
					if err != nil {
						ffmt.Mark(err)
						return
					}
				})
			}
		}
	}

	f.Join()

	return nil
}

// DiffHub 比较
func DiffHub(dst, src Filehub, path string) ([]*DiffInfo, error) {
	fd, err := dst.List(path)
	if err != nil {
		return nil, err
	}

	fs, err := src.List(path)
	if err != nil {
		return nil, err
	}

	return Diff(fd, fs), nil
}

// Diff 排序后比较文件名差异
func Diff(fd, fs []FileInfo) (dds []*DiffInfo) {
	sort.SliceStable(fd, func(i, j int) bool {
		return fd[i].Path() < fd[j].Path()
	})
	sort.SliceStable(fs, func(i, j int) bool {
		return fs[i].Path() < fs[j].Path()
	})

	si := 0
	di := 0

	for di != len(fd) && si != len(fs) {
		vd := fd[di]
		vs := fs[si]
		vdp := vd.Path()
		vsp := vs.Path()
		if vdp < vsp {
			vs = nil
		} else if vdp > vsp {
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
			di++
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
	return fmt.Sprintf("%v =:= %v", d.Dst, d.Src)
}
