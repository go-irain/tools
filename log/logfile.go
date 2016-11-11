package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func (log *Log) setOutputHandler(size uint64) error {
	filename := log.path + "last.log"
	f, erro := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if erro != nil {
		return erro
	}
	log.w = f
	log.nowfilesize = size
	return nil
}

// 切割文件
func (log *Log) splitFile() error {
	// 获取日志目录下的文件列表
	// 并按照名字排序
	fs, _ := filepath.Glob(log.path + "*.log")
	fslen := len(fs)
	if fslen > 1 {
		sort.Strings(fs)
	}

	// 保证日志文件总数为maxcoutnum
	// 超过则会删除比较早的log文件
	if fslen >= log.maxcoutnum {
		var diff int
		if log.maxcoutnum >= fslen {
			diff = log.maxcoutnum - fslen
		} else {
			diff = fslen - log.maxcoutnum
		}
		sliceend := diff + 1
		for _, v := range fs[0:sliceend] {
			os.Remove(v)
		}
		fs = fs[sliceend:]
		fslen = len(fs)
	}

	if fslen > 0 {
		if log.w == nil {
			if finfo, erro := os.Stat(log.path + "last.log"); erro == nil {
				oldsize := uint64(finfo.Size())
				if oldsize < log.maxsize {
					// 老文件大小不够不创建而是追加
					return log.setOutputHandler(oldsize)
				}
			}
		} else {
			if fd, ok := log.w.(*os.File); ok {
				fd.Close()
			}
		}

		var lastindex int
		if fslen > 1 {
			// 如果日志文件大于1个则把当前最后一个重新按照编号命名
			lastfilename := filepath.Base(fs[fslen-2])
			last2index := strings.TrimLeft(strings.Split(lastfilename, ".")[0], "0")
			la, _ := strconv.Atoi(last2index)
			lastindex = la + 1
		}
		os.Rename(log.path+"last.log", fmt.Sprintf("%s%08d.log", log.path, lastindex))
	}
	// 新文件
	return log.setOutputHandler(0)
}
