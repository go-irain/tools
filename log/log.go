// afocus 175228533@qq.com

package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Log struct {
	// 日志级别
	// 小于level级别的日志将不再记录
	//level uint32 使用globalLevel

	// 日志保存的路径
	// 最好abs路径
	path string

	// maxsize单个日志文件最大byte 单位MB
	maxsize int64
	// maxcoutnum 最大日志文件数量 0不记录
	maxcoutnum int

	// 是否记录行号
	// 测试使用 正式环境不建议开启 尽量减少写日志影响正常功能
	// 随着代码版本的不同 记录的文件行号也不会准确
	showFileline bool

	// 日志等级
	level Level

	// 负责写文件
	// 一个日志可能会写不同的文件名
	tags           map[string]io.Writer
	defaultTagName string

	bufpool *sync.Pool

	// filter 等级过滤器 可以自定义一些处理
	filters map[Level]func(string, string)

	lock sync.Mutex
}

// NewLog 初始化一个新的log
func NewLog() *Log {
	l := &Log{
		filters:        make(map[Level]func(string, string)),
		showFileline:   true,
		level:          L_DEBUG,
		tags:           make(map[string]io.Writer),
		defaultTagName: (strings.Split(filepath.Base(os.Args[0]), "."))[0],
		bufpool: &sync.Pool{New: func() interface{} {
			return bytes.NewBuffer([]byte{})
		}},
	}
	l.SetTags(l.defaultTagName)
	return l
}

var logpaths = []string{}
var timenowfunc = time.Now

// SetTags 设置标签 作用是将日志输出到指定前缀的文件中
func (log *Log) SetTags(tagnames ...string) error {
	log.lock.Lock()
	defer log.lock.Unlock()
	for _, name := range tagnames {
		name = strings.TrimSpace(name)
		if len(name) == 0 {
			return errors.New("SetTags params is empty")
		}
		if _, ok := log.tags[name]; ok {
			return errors.New("SetTags params subname is exist:" + name)
		}
		if log.path == "" {
			log.tags[name] = os.Stdout
		} else {
			log.tags[name] = NewWriteIO(log.path, name+"_", log.maxsize, log.maxcoutnum)
		}
	}
	return nil
}

// SetLevel 设置日志等级
func (log *Log) SetLevel(level Level) {
	log.lock.Lock()
	defer log.lock.Unlock()
	if level > L_ERROR {
		level = L_ERROR
	}
	log.level = level
}

// SetShowLineNumber 是否显示行号
func (log *Log) SetShowLineNumber(show bool) {
	log.lock.Lock()
	defer log.lock.Unlock()
	log.showFileline = show
}

// SetHook 钩子 可以对指定的等级log进行自定义处理
func (log *Log) SetHook(levl Level, hander func(tag, msg string)) error {
	log.lock.Lock()
	defer log.lock.Unlock()
	if _, ok := log.filters[levl]; ok {
		return errors.New("filter error:level is alreay set")
	}
	log.filters[levl] = hander
	return nil
}

// SetOutDirConfig 设置输出目录 默认输出到控制台
// maxsize 单个日志文件的最大大小 单位MB maxcount 目录下最大日志文件数量 每个独立的tag日志 独立计算
func (log *Log) SetOutDirConfig(path string, maxsize int, maxcount int) {
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
	log.maxsize = int64(maxsize) * MB
	log.maxcoutnum = maxcount
	for name := range log.tags {
		log.tags[name] = NewWriteIO(log.path, name+"_", log.maxsize, log.maxcoutnum)
	}
}

func (log *Log) Output(tag string, level Level, calldepth int, str string) {
	if level > L_ERROR || level < log.level {
		return
	}
	buf := log.bufpool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.WriteString(timenowfunc().Format("2006/01/02 15:04:05"))
	buf.WriteString(fmt.Sprintf(" [%s] <%s> ", level, tag))
	if log.showFileline {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		length := len(file) - 1
		for i := length; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1 : length-2]
				break
			}
		}
		buf.WriteString(fmt.Sprintf("%s:%d ", file, line))
	}
	if len(str) == 0 || str[len(str)-1] != '\n' {
		str += "\n"
	}
	buf.WriteString(str)
	log.lock.Lock()
	filter, fok := log.filters[level]
	out, ok := log.tags[tag]
	if !ok {
		out = log.tags[log.defaultTagName]
	}
	_, erro := buf.WriteTo(out)
	log.lock.Unlock()
	if fok {
		filter(tag, str)
	}
	if erro != nil {
		println(erro.Error())
	}
	log.bufpool.Put(buf)
}

// TagDebug 调试输出
func (log *Log) TagDebug(tag string, a ...interface{}) {
	log.Output(tag, L_DEBUG, 3, fmt.Sprintln(a...))
}

// TagInfo 普通信息
func (log *Log) TagInfo(tag string, a ...interface{}) {
	log.Output(tag, L_INFO, 3, fmt.Sprintln(a...))
}

// TagWarnning 警告
func (log *Log) TagWarnning(tag string, a ...interface{}) {
	log.Output(tag, L_WARN, 3, fmt.Sprintln(a...))
}

// TagError 错误
func (log *Log) TagError(tag string, a ...interface{}) {
	log.Output(tag, L_ERROR, 3, fmt.Sprintln(a...))
}

// TagDebugf 格式化debug输出
func (log *Log) TagDebugf(tag, format string, a ...interface{}) {
	log.Output(tag, L_DEBUG, 3, fmt.Sprintf(format, a...))
}

// TagInfof 格式化info输出
func (log *Log) TagInfof(tag, format string, a ...interface{}) {
	log.Output(tag, L_INFO, 3, fmt.Sprintf(format, a...))
}

// TagWarnningf 格式化warnning输出
func (log *Log) TagWarnningf(tag, format string, a ...interface{}) {
	log.Output(tag, L_WARN, 3, fmt.Sprintf(format, a...))
}

// TagErrorf 格式化error输出
func (log *Log) TagErrorf(tag, format string, a ...interface{}) {
	log.Output(tag, L_ERROR, 3, fmt.Sprintf(format, a...))
}

// Debug 调试输出
func (log *Log) Debug(a ...interface{}) {
	log.TagDebug(log.defaultTagName, a...)
}

// Info 普通信息
func (log *Log) Info(a ...interface{}) {
	log.TagInfo(log.defaultTagName, a...)
}

// Warnning 警告
func (log *Log) Warnning(a ...interface{}) {
	log.TagWarnning(log.defaultTagName, a...)
}

// Error 错误
func (log *Log) Error(a ...interface{}) {
	log.TagError(log.defaultTagName, a...)
}

// Debugf 格式化debug输出
func (log *Log) Debugf(format string, a ...interface{}) {
	log.TagDebugf(log.defaultTagName, format, a...)
}

// Infof 格式化info输出
func (log *Log) Infof(format string, a ...interface{}) {
	log.TagInfof(log.defaultTagName, format, a...)
}

// Warnningf 格式化warnning输出
func (log *Log) Warnningf(format string, a ...interface{}) {
	log.TagWarnningf(log.defaultTagName, format, a...)
}

// Errorf 格式化error输出
func (log *Log) Errorf(format string, a ...interface{}) {
	log.TagErrorf(log.defaultTagName, format, a...)
}
