// afocus 175228533@qq.com

package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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

	// 负责写文件
	// 一个日志可能会写不同的文件名
	outputs           map[string]*Output
	defaultOutputname string

	// filter 等级过滤器 可以自定义一些处理
	filters map[Level]func(string)

	lock sync.Mutex
}

// NewLog 初始化一个新的log
func NewLog() *Log {
	l := &Log{
		filters:           make(map[Level]func(string)),
		showFileline:      true,
		level:             L_DEBUG,
		outputs:           make(map[string]*Output),
		defaultOutputname: (strings.Split(os.Args[0], "."))[0],
	}
	l.AddSub(l.defaultOutputname)
	return l
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

func (log *Log) AddSub(name string) error {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return errors.New("AddSub params subname is empty")
	}
	log.lock.Lock()
	defer log.lock.Unlock()
	if _, ok := log.outputs[name]; ok {
		return errors.New("AddSub params subname is exist:" + name)
	}
	path := log.path + name + "_"
	log.outputs[name] = &Output{
		loger: log,
		wio:   NewWriteIO(path, log.maxsize, log.maxcoutnum),
	}
	return nil
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

	for k, v := range log.outputs {
		v.updateConfig(k)
	}
}

func (log *Log) Sub(name string) *Output {
	if v, ok := log.outputs[name]; ok {
		return v
	}
	return log.outputs[log.defaultOutputname]
}

// Debug 调试输出
func (log *Log) Debug(a ...interface{}) {
	log.Sub(log.defaultOutputname).Debug(a...)
}

// Info 普通信息
func (log *Log) Info(a ...interface{}) {
	log.Sub(log.defaultOutputname).Info(a...)
}

// Warnning 警告
func (log *Log) Warnning(a ...interface{}) {
	log.Sub(log.defaultOutputname).Warnning(a...)
}

// Error 错误
func (log *Log) Error(a ...interface{}) {
	log.Sub(log.defaultOutputname).Error(a...)
}

// Debugf 格式化debug输出
func (log *Log) Debugf(format string, a ...interface{}) {
	log.Sub(log.defaultOutputname).Debugf(format, a...)
}

// Infof 格式化info输出
func (log *Log) Infof(format string, a ...interface{}) {
	log.Sub(log.defaultOutputname).Infof(format, a...)
}

// Warnningf 格式化warnning输出
func (log *Log) Warnningf(format string, a ...interface{}) {
	log.Sub(log.defaultOutputname).Warnningf(format, a...)
}

// Errorf 格式化error输出
func (log *Log) Errorf(format string, a ...interface{}) {
	log.Sub(log.defaultOutputname).Errorf(format, a...)
}

type Output struct {
	loger *Log
	wio   *WriteIO
}

func (o *Output) updateConfig(name string) {
	o.wio.path = o.loger.path + name + "_"
	o.wio.maxcoutnum = o.loger.maxcoutnum
	o.wio.maxsize = o.loger.maxsize
}

func (o *Output) print(level Level, str string) error {
	if level > L_ERROR || level < o.loger.level {
		return nil
	}

	buf := make([]byte, 0, len(str)+20)
	buf = append(buf, timenowfunc().Format("2006/01/02 15:04:05")...)
	buf = append(buf, fmt.Sprintf(" [%s] ", level)...)

	if o.loger.showFileline {
		_, file, line, ok := runtime.Caller(defaultCar)
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
		buf = append(buf, fmt.Sprintf("%s:L%d ", file, line)...)
	}
	msgbytes := stringToSlice(str)
	if len(msgbytes) == 0 || msgbytes[len(msgbytes)-1] != '\n' {
		msgbytes = append(msgbytes, '\n')
	}
	buf = append(buf, msgbytes...)
	if filter, ok := o.loger.filters[level]; ok {
		filter(string(buf))
	}

	if o.loger.path == "" {
		_, erro := os.Stderr.Write(buf)
		return erro
	}
	return o.wio.write(buf)
}

// Debug 调试输出
func (o *Output) Debug(a ...interface{}) {
	o.print(L_DEBUG, fmt.Sprintln(a...))
}

// Info 普通信息
func (o *Output) Info(a ...interface{}) {
	o.print(L_INFO, fmt.Sprintln(a...))
}

// Warnning 警告
func (o *Output) Warnning(a ...interface{}) {
	o.print(L_WARN, fmt.Sprintln(a...))
}

// Error 错误
func (o *Output) Error(a ...interface{}) {
	o.print(L_ERROR, fmt.Sprintln(a...))
}

// Debugf 格式化debug输出
func (o *Output) Debugf(format string, a ...interface{}) {
	o.print(L_DEBUG, fmt.Sprintf(format, a...))
}

// Infof 格式化info输出
func (o *Output) Infof(format string, a ...interface{}) {
	o.print(L_INFO, fmt.Sprintf(format, a...))
}

// Warnningf 格式化warnning输出
func (o *Output) Warnningf(format string, a ...interface{}) {
	o.print(L_WARN, fmt.Sprintf(format, a...))
}

// Errorf 格式化error输出
func (o *Output) Errorf(format string, a ...interface{}) {
	o.print(L_ERROR, fmt.Sprintf(format, a...))
}
