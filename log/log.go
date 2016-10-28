// afocus 175228533@qq.com

package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	// 最大日志文件数量 0不记录
	maxsize    uint64
	maxcoutnum int

	// 是否记录行号
	// 测试使用 正式环境不建议开启 尽量减少写日志影响正常功能
	// 随着代码版本的不同 记录的文件行号也不会准确
	// showFileline bool 使用globalLine

	nowfilesize uint64
	buf         []byte
	msg         chan string
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

// LStr level string name
// var LStr = []string{
// 	" [DBUG] ",
// 	" [INFO] ",
// 	" [WARN] ",
// 	" [ERRO] ",
// }

const (
	_         = iota
	KB uint64 = 1 << (10 * iota)
	MB
)

var golbalLogger *Log

var golbalLevel LogLevel
var golbalLine = true

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

var timenowfunc = time.Now

func (log *Log) write(str string) error {
	if log.w == nil || log.nowfilesize > log.maxsize {
		if erro := log.splitFile(); erro != nil {
			println("err", erro)
			return erro
		}
	}

	t := timenowfunc()
	log.buf = log.buf[:0]

	year, month, day := t.Date()
	itoa(&log.buf, year, 4)
	log.buf = append(log.buf, '-')
	itoa(&log.buf, int(month), 2)
	log.buf = append(log.buf, '-')
	itoa(&log.buf, day, 2)
	log.buf = append(log.buf, ' ')

	hour, min, sec := t.Clock()
	itoa(&log.buf, hour, 2)
	log.buf = append(log.buf, ':')
	itoa(&log.buf, min, 2)
	log.buf = append(log.buf, ':')
	itoa(&log.buf, sec, 2)

	log.buf = append(log.buf, str...)
	if len(str) == 0 || str[len(str)-1] != '\n' {
		log.buf = append(log.buf, '\n')
	}
	size, erro := log.w.Write(log.buf)
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
	str = fmt.Sprintf(" [%s] %s", level, str)
	//str = LStr[int(level)] + str
	// 未设置输出目录则直接stdout
	if log == nil {
		if len(str) == 0 || str[len(str)-1] != '\n' {
			str = string(append([]byte(str), '\n'))
		}
		print(time.Now().Format("2006-01-02 15:04:05") + str)
		return
	}
	// push到队列中
	log.msg <- str
}

func SetFlag(level LogLevel, showline bool) {
	if level > L_ERROR {
		level = L_ERROR
	}
	golbalLevel = level
	golbalLine = showline
}

// Flush
// 设置SetOutDir后务必调用Flush 保证数据全部写入文件
// log.SetOutDir(....)
// defer log.Flush()
func Flush() {
	if golbalLogger != nil {
		time.Sleep(time.Millisecond * 1)
		// 关闭chan 停止写入
		close(golbalLogger.msg)
		for {
			// 检查缓存的消息是否全部写入完毕
			// 写入完毕则退出
			if len(golbalLogger.msg) == 0 {
				return
			}
		}
	}
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
		msg:        make(chan string, 512),
	}
	go func() {
		for v := range log.msg {
			if err := log.write(v); err != nil {
				println(err.Error())
			}
		}
	}()
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
