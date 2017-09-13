package filehub

import (
	"fmt"
	"sort"

	"github.com/wzshiming/ffmt"
	"github.com/wzshiming/fork"
)

type CopyType uint32

const (
	None    = CopyType(0)
	_       = CopyType(1 << (32 - iota - 1))
	Create  // 如果源有目标没有的文件则在目标创建文件
	Replace // 如果源比目标新或目标就替换目标的文件
	Delete  // 如果目标有源没有的文件则删除目标的文件
)

func (c CopyType) Exists(ct CopyType) bool {
	return c&ct == ct
}

// Copy
func Copy(dst, src Filehub, ct CopyType, forkSize int) error {
	if ct == None {
		return nil
	}
	dds, err := DiffHub(dst, src)
	if err != nil {
		return err
	}

	f := fork.NewFork(forkSize)

	for _, v := range dds {
		func(v *DiffInfo) {
			f.Push(func() {
				if v.Src != nil {
					if (v.Dst == nil && ct.Exists(Create)) ||
						(v.Src.ModTime().After(v.Dst.ModTime()) && ct.Exists(Replace)) {
						p := v.Src.Path()
						d, t, err := src.Get(p)
						if err != nil {
							ffmt.Mark(err)
							return
						}
						dst.Put(p, d, t)
					}
				} else {
					if ct.Exists(Delete) {
						err := dst.Del(v.Dst.Path())
						if err != nil {
							ffmt.Mark(err)
							return
						}
					}
				}
			})

		}(v)
	}
	f.Join()

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
