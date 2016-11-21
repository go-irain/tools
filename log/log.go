// afocus 175228533@qq.com

package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type Log struct {
	// 日志级别
	// 小于level级别的日志将不再记录
	//level uint32 使用globalLevel

	// 日志保存的路径
	// 最好abs路径
	path string

	// maxsize单个日志文件最大byte 单位MB
	maxsize uint64
	// maxcoutnum 最大日志文件数量 0不记录
	maxcoutnum int

	// 是否记录行号
	// 测试使用 正式环境不建议开启 尽量减少写日志影响正常功能
	// 随着代码版本的不同 记录的文件行号也不会准确
	showFileline bool

	// 日志等级
	level Level

	// filter 等级过滤器 可以自定义一些处理
	filters map[Level]func(string)

	// 当前文件写入的未知
	nowfilesize uint64
	w           io.Writer

	lock sync.Mutex
}

// NewLog 初始化一个新的log
func NewLog() *Log {
	return &Log{
		filters:      make(map[Level]func(string)),
		showFileline: true,
		level:        L_DEBUG,
	}
}

var logpaths = []string{}

var timenowfunc = time.Now

func stringToSlice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

func (log *Log) output(calldepth int, level Level, str string) error {
	if level > L_ERROR || level < log.level {
		return nil
	}
	if log.showFileline {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		short := file
		length := len(file) - 1
		for i := length; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1 : length-2]
				break
			}
		}
		file = short
		str = fmt.Sprintf("%s:L%d %s", file, line, str)
	}
	str = fmt.Sprintf("%s [%s] %s", timenowfunc().Format("2006/01/02 15:04:05"), level, str)
	if len(str) == 0 || str[len(str)-1] != '\n' {
		str = string(append([]byte(str), '\n'))
	}
	if filter, ok := log.filters[level]; ok {
		filter(str)
	}
	msgbytes := stringToSlice(str)
	if log.path == "" {
		_, erro := os.Stderr.Write(msgbytes)
		return erro
	}
	log.lock.Lock()
	if log.w == nil || log.nowfilesize > log.maxsize {
		if erro := log.splitFile(); erro != nil {
			log.lock.Unlock()
			return erro
		}
	}
	size, erro := log.w.Write(msgbytes)
	log.nowfilesize += uint64(size)
	log.lock.Unlock()
	return erro

}

// SetFlag 设置是否显示行号，以及日志等级
func (log *Log) SetFlag(level Level, showline bool) {
	if level > L_ERROR {
		level = L_ERROR
	}
	log.lock.Lock()
	log.level = level
	log.showFileline = showline
	log.lock.Unlock()
}

// SetFilter 过滤气 可以对指定的等级log进行自定义处理
func (log *Log) SetFilter(levl Level, hander func(msg string)) error {
	log.lock.Lock()
	defer log.lock.Unlock()
	if _, ok := log.filters[levl]; ok {
		return errors.New("filter error:level is alreay set")
	}
	log.filters[levl] = hander
	return nil
}

// SetOutDir 设置输出目录 默认输出到控制台
func (log *Log) SetOutDir(path string, maxsize int, maxcount int) {
	if maxsize <= 0 || maxcount <= 0 {
		return
	}
	log.lock.Lock()
	defer log.lock.Unlock()
	path, _ = filepath.Abs(path)
	println("log SetOutDir:", path)
	for i := range logpaths {
		if logpaths[i] == path {
			println("log SetOutDir error: logpath is exist ->" + path)
			os.Exit(1)
		}
	}
	erro := os.MkdirAll(path, 0777)
	if erro != nil {
		println("log SetOutDir error:" + erro.Error())
		os.Exit(1)
	}
	logpaths = append(logpaths, path)
	fi, erro := os.Stat(path)
	if erro != nil && os.IsNotExist(erro) || erro == nil && !fi.IsDir() {
		println("log SetOutDir error:path is not a dir")
		os.Exit(1)
	}
	if fi.Mode().Perm() < 0666 {
		println("log SetOutDir error:path no permissions to read and write")
		os.Exit(1)
	}
	log.path = filepath.Clean(path) + string(filepath.Separator)
	log.maxsize = uint64(maxsize) * MB
	log.maxcoutnum = maxcount

}

// Debug 调试输出
func (log *Log) Debug(a ...interface{}) {
	log.output(defaultCar, L_DEBUG, fmt.Sprintln(a...))
}

// Info 普通信息
func (log *Log) Info(a ...interface{}) {
	log.output(defaultCar, L_INFO, fmt.Sprintln(a...))
}

// Warnning 警告
func (log *Log) Warnning(a ...interface{}) {
	log.output(defaultCar, L_WARN, fmt.Sprintln(a...))
}

// Error 错误
func (log *Log) Error(a ...interface{}) {
	log.output(defaultCar, L_ERROR, fmt.Sprintln(a...))
}

// Debugf 格式化debug输出
func (log *Log) Debugf(format string, a ...interface{}) {
	log.output(defaultCar, L_DEBUG, fmt.Sprintf(format, a...))
}

// Infof 格式化info输出
func (log *Log) Infof(format string, a ...interface{}) {
	log.output(defaultCar, L_INFO, fmt.Sprintf(format, a...))
}

// Warnningf 格式化warnning输出
func (log *Log) Warnningf(format string, a ...interface{}) {
	log.output(defaultCar, L_WARN, fmt.Sprintf(format, a...))
}

// Errorf 格式化error输出
func (log *Log) Errorf(format string, a ...interface{}) {
	log.output(defaultCar, L_ERROR, fmt.Sprintf(format, a...))
}
