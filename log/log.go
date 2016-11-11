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
	// 最大日志文件数量 0不记录
	maxsize    uint64
	maxcoutnum int

	// 是否记录行号
	// 测试使用 正式环境不建议开启 尽量减少写日志影响正常功能
	// 随着代码版本的不同 记录的文件行号也不会准确
	// showFileline bool 使用globalLine

	nowfilesize uint64
	w           io.Writer
}

type LogLevel uint32

const (
	L_DEBUG LogLevel = iota
	L_INFO
	L_WARN
	L_ERROR
)

func (l LogLevel) String() string {
	switch l {
	case 0:
		return "DBUG"
	case 1:
		return "INFO"
	case 2:
		return "WARN"
	case 3:
		return "ERRO"
	default:
		return "UNKN"
	}
}

const (
	_         = iota
	KB uint64 = 1 << (10 * iota)
	MB
)

var golbalLogger *Log

// 默认显示等级
var golbalLevel LogLevel

// 默认显示行数
var golbalLine = true

// filter 等级过滤器 可以自定义一些处理
var golbalFilter = map[LogLevel]func(string){}

var timenowfunc = time.Now

func stringToSlice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

func (log *Log) write(str string) error {
	if log.w == nil || log.nowfilesize > log.maxsize {
		if erro := log.splitFile(); erro != nil {
			println("err", erro)
			return erro
		}
	}
	size, erro := log.w.Write(stringToSlice(str))
	log.nowfilesize += uint64(size)
	return erro
}

func (log *Log) output(calldepth int, level LogLevel, str string) {
	if level > L_ERROR {
		return
	}
	// 等级小于当前规定的等级不输出
	if level < golbalLevel {
		return
	}
	if golbalLine {
		_, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		str = fmt.Sprintf("%s#%d %s", file, line, str)
	}
	str = fmt.Sprintf("%s [%s] %s", timenowfunc().Format("2006/01/02 15:04:05"), level, str)
	if len(str) == 0 || str[len(str)-1] != '\n' {
		str = string(append([]byte(str), '\n'))
	}
	if filter, ok := golbalFilter[level]; ok {
		filter(str)
	}
	// 未设置输出目录则直接stdout
	if log == nil {
		print(str)
	} else {
		// push到队列中
		log.write(str)
	}

}

func SetFlag(level LogLevel, showline bool) {
	if level > L_ERROR {
		level = L_ERROR
	}
	golbalLevel = level
	golbalLine = showline
}

func SetFilter(levl LogLevel, hander func(msg string)) error {
	if _, ok := golbalFilter[levl]; ok {
		return errors.New("filter error:level is alreay set")
	}
	golbalFilter[levl] = hander
	return nil
}

// SetOutDir 只有设置了目录等属性 才会真正的写入文件
func SetOutDir(path string, maxsize int, maxcount int) error {

	if golbalLogger != nil {
		return fmt.Errorf("the log output path has been set up.")
	}
	if maxsize <= 0 || maxcount <= 0 {
		return nil
	}
	path, _ = filepath.Abs(path)
	println("logDir:", path)
	erro := os.MkdirAll(path, 0777)
	if erro != nil {
		return erro
	}
	fi, erro := os.Stat(path)
	if erro != nil && os.IsNotExist(erro) || erro == nil && !fi.IsDir() {
		return fmt.Errorf("log SetOutDir error:%s is not a dir", path)
	}
	if fi.Mode().Perm() < 0666 {
		return fmt.Errorf("log SetOutDir error:%s no permissions to read and write", path)
	}

	log := &Log{
		maxcoutnum: maxcount,
		maxsize:    uint64(maxsize) * MB,
		path:       filepath.Clean(path) + string(filepath.Separator),
	}
	golbalLogger = log
	return nil
}

var defaultCar = 2

// Debug 调试输出
func Debug(a ...interface{}) {
	golbalLogger.output(defaultCar, L_DEBUG, fmt.Sprintln(a...))
}

// Info 普通信息
func Info(a ...interface{}) {
	golbalLogger.output(defaultCar, L_INFO, fmt.Sprintln(a...))
}

// Warnning 警告
func Warnning(a ...interface{}) {
	golbalLogger.output(defaultCar, L_WARN, fmt.Sprintln(a...))
}

// Error 错误
func Error(a ...interface{}) {
	golbalLogger.output(defaultCar, L_ERROR, fmt.Sprintln(a...))
}

// Debugf 格式化debug输出
func Debugf(format string, a ...interface{}) {
	golbalLogger.output(defaultCar, L_DEBUG, fmt.Sprintf(format, a...))
}

// Infof 格式化info输出
func Infof(format string, a ...interface{}) {
	golbalLogger.output(defaultCar, L_INFO, fmt.Sprintf(format, a...))
}

// Warnningf 格式化warnning输出
func Warnningf(format string, a ...interface{}) {
	golbalLogger.output(defaultCar, L_WARN, fmt.Sprintf(format, a...))
}

// Errorf 格式化error输出
func Errorf(format string, a ...interface{}) {
	golbalLogger.output(defaultCar, L_ERROR, fmt.Sprintf(format, a...))
}
