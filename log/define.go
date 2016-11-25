package log

// 日志大小单位
const (
	_         = iota
	KB uint64 = 1 << (10 * iota)
	MB
)

// Level 日志等级
type Level uint32

const (
	L_DEBUG Level = iota // 调试
	L_INFO               // 信息
	L_WARN               // 警告
	L_ERROR              // 错误
)

func (level Level) String() string {
	switch level {
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

const defaultCar = 2

// 默认自动创建一个全局的log 直接使用
var golbalLogger *Log

func init() {
	golbalLogger = NewLog()
}

// 以下方法位全局映射

func AddSub(name string) error {
	return golbalLogger.AddSub(name)
}

func Sub(name string) *Output {
	return golbalLogger.Sub(name)
}

// SetFlag 设置是否显示行号，以及日志等级
func SetFlag(level Level, showline bool) {
	golbalLogger.SetFlag(level, showline)
}

// SetFilter 过滤气 可以对指定的等级log进行自定义处理
func SetFilter(levl Level, hander func(msg string)) error {
	return golbalLogger.SetFilter(levl, hander)
}

// SetOutDir 设置输出目录 默认输出到控制台
// maxsize 单个文件最大 单位MB maxcount 文件夹最多保存多少个文件
func SetOutDir(path string, maxsize int, maxcount int) {
	golbalLogger.SetOutDir(path, maxsize, maxcount)
}

// Debug 调试输出
func Debug(a ...interface{}) {
	golbalLogger.Debug(a...)
}

// Info 普通信息
func Info(a ...interface{}) {
	golbalLogger.Info(a...)
}

// Warnning 警告
func Warnning(a ...interface{}) {
	golbalLogger.Warnning(a...)
}

// Error 错误
func Error(a ...interface{}) {
	golbalLogger.Error(a...)
}

// Debugf 格式化debug输出
func Debugf(format string, a ...interface{}) {
	golbalLogger.Debugf(format, a...)
}

// Infof 格式化info输出
func Infof(format string, a ...interface{}) {
	golbalLogger.Infof(format, a...)
}

// Warnningf 格式化warnning输出
func Warnningf(format string, a ...interface{}) {
	golbalLogger.Warnningf(format, a...)
}

// Errorf 格式化error输出
func Errorf(format string, a ...interface{}) {
	golbalLogger.Errorf(format, a...)
}
