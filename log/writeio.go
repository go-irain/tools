package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type WriteIO struct {

	// 日志保存的路径
	// 最好abs路径
	path string

	// maxsize单个日志文件最大byte 单位MB
	maxsize uint64
	// maxcoutnum 最大日志文件数量 0不记录
	maxcoutnum int

	// 当前文件写入的未知
	nowfilesize uint64
	w           io.Writer

	lock sync.Mutex
}

func NewWriteIO(path string, maxsize uint64, maxcoutnum int) *WriteIO {
	return &WriteIO{
		path:       path,
		maxsize:    maxsize,
		maxcoutnum: maxcoutnum,
	}
}

func (w *WriteIO) setOutputHandler(size uint64) error {
	filename := w.path + "last.log"
	f, erro := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if erro != nil {
		return erro
	}
	w.w = f
	w.nowfilesize = size
	return nil
}

// 切割文件
func (w *WriteIO) splitFile() error {
	// 获取日志目录下的文件列表
	// 并按照名字排序
	fmt.Println(w.path)
	fs, _ := filepath.Glob(w.path + "*.log")
	fmt.Println(fs)
	fslen := len(fs)
	if fslen > 1 {
		sort.Strings(fs)
	}

	// 保证日志文件总数为maxcoutnum
	// 超过则会删除比较早的log文件
	if fslen >= w.maxcoutnum {
		sliceend := fslen - w.maxcoutnum + 1
		for _, v := range fs[0:sliceend] {
			os.Remove(v)
		}
		fs = fs[sliceend:]
		fslen = len(fs)
	}
	if fslen > 0 {
		if w.w == nil {
			if finfo, erro := os.Stat(w.path + "last.log"); erro == nil {
				oldsize := uint64(finfo.Size())
				if oldsize < w.maxsize {
					// 老文件大小不够不创建而是追加
					return w.setOutputHandler(oldsize)
				}
			}
		} else {
			if fd, ok := w.w.(*os.File); ok {
				fd.Close()
			}
		}
		var lastindex int
		if fslen > 1 {
			// 如果日志文件大于1个则把当前最后一个重新按照编号命名
			lastfilename := filepath.Base(fs[fslen-2])
			lastfilename = strings.TrimRight(lastfilename, ".log")
			parts := strings.Split(lastfilename, "_")
			var shortindexname string
			if len(parts) == 1 {
				shortindexname = parts[0]
			} else {
				shortindexname = parts[len(parts)-1]
			}
			last2index := strings.TrimLeft(shortindexname, "0")
			la, _ := strconv.Atoi(last2index)
			lastindex = la + 1
		}
		os.Rename(w.path+"last.log", fmt.Sprintf("%s%08d.log", w.path, lastindex))
	}
	// 新文件
	return w.setOutputHandler(0)
}

func (w *WriteIO) write(msg []byte) error {
	w.lock.Lock()
	if w.w == nil || w.nowfilesize > w.maxsize {
		fmt.Println("--------------------------------------------------")
		if erro := w.splitFile(); erro != nil {
			w.lock.Unlock()
			return erro
		}
	}
	size, erro := w.w.Write(msg)
	w.nowfilesize += uint64(size)
	w.lock.Unlock()
	return erro
}
