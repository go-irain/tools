package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type WriteIO struct {
	prefix string
	// path = dirpath + prefix
	path string
	//
	maxsize  int64
	maxcount int
	// 当前文件写入的大小
	wsize int64
	out   io.Writer
}

const minsize = 10
const mincount = 3

func NewWriteIO(dir, prefix string, maxsize int64, maxcount int) *WriteIO {
	if maxsize < minsize {
		maxsize = minsize
	}
	if maxcount < mincount {
		maxcount = mincount
	}
	return &WriteIO{
		prefix:   prefix,
		path:     dir + prefix,
		maxcount: maxcount,
		maxsize:  maxsize,
	}
}

func (o *WriteIO) split() error {
	// 获取目录下指定前缀的所有日志文件
	fs, erro := filepath.Glob(o.path + "*.log")
	if erro != nil {
		return erro
	}
	fslen := len(fs)
	if fslen > 0 {
		// 获取最后一格文件(last.log)的路径
		lastpath := fs[fslen-1]
		o.out.(*os.File).Close()
		var index int
		if fslen > 1 {
			sort.Strings(fs)
			// 如果日志数量超过指定数量则删除最老的文件
			if delcount := fslen + 1 - o.maxcount; delcount > 0 {
				for _, p := range fs[:delcount] {
					os.Remove(p)
				}
			}
			op := filepath.Base(fs[fslen-2])
			index, _ = strconv.Atoi(op[len(o.prefix) : len(op)-4])
		}
		// 把当前文件重命名为含有小标的名字
		os.Rename(lastpath, fmt.Sprintf("%s%08d.log", o.path, index+1))
	}
	// 创建新的last文件
	return o.create()
}

func (o *WriteIO) create() error {
	f, erro := os.OpenFile(o.path+"last.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if erro == nil {
		var info os.FileInfo
		info, erro = f.Stat()
		if erro == nil {
			o.wsize = info.Size()
			o.out = f
		}
	}
	return erro
}

// Write 实现io.Write接口
// 主要添加了分割文件的功能
func (o *WriteIO) Write(data []byte) (int, error) {
	if o.out == nil {
		if erro := o.create(); erro != nil {
			return 0, erro
		}
	}
	size, erro := o.out.Write(data)
	if erro != nil {
		return 0, erro
	}
	o.wsize += int64(size)
	if o.wsize > o.maxsize {
		if erro = o.split(); erro != nil {
			return size, erro
		}
	}
	return size, nil
}
