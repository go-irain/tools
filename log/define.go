package log

import "fmt"

// 日志大小单位
const (
	_        = iota
	KB int64 = 1 << (10 * iota)
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
	case L_DEBUG:
		return "DBUG"
	case L_INFO:
		return "INFO"
	case L_WARN:
		return "WARN"
	case L_ERROR:
		return "ERRO"
	default:
		return "UNKN"
	}
}

// 默认自动创建一个全局的log 直接使用
var golbalLogger *Log

func init() {
	golbalLogger = NewLog()
}

// 以下方法位全局映射

// SetTags 设置标签
func SetTags(tagnames ...string) error {
	return golbalLogger.SetTags(tagnames...)
}

// SetShowLineNumber 是否显示行号
func SetShowLineNumber(showline bool) {
	golbalLogger.SetShowLineNumber(showline)
}

// SetLevel 设置日志等级
func SetLevel(level Level) {
	golbalLogger.SetLevel(level)
}

// SetHook 过滤气 可以对指定的等级log进行自定义处理
func SetHook(levl Level, hander func(tag, msg string)) error {
	return golbalLogger.SetHook(levl, hander)
}

// SetOutDirConfig 设置输出目录 默认输出到控制台
// maxsize 单个文件最大 单位MB maxcount 文件夹最多保存多少个文件
func SetOutDirConfig(path string, maxsize int, maxcount int) {
	golbalLogger.SetOutDirConfig(path, maxsize, maxcount)
}

// TagDebug 调试输出
func TagDebug(tag string, a ...interface{}) {
	golbalLogger.Output(tag, L_DEBUG, 2, fmt.Sprintln(a...))
}

// TagInfo 普通信息
func TagInfo(tag string, a ...interface{}) {
	golbalLogger.Output(tag, L_INFO, 2, fmt.Sprintln(a...))
}

// TagWarnning 警告
func TagWarnning(tag string, a ...interface{}) {
	golbalLogger.Output(tag, L_WARN, 2, fmt.Sprintln(a...))
}

// TagError 错误
func TagError(tag string, a ...interface{}) {
	golbalLogger.Output(tag, L_ERROR, 2, fmt.Sprintln(a...))
}

// TagDebugf 格式化debug输出
func TagDebugf(tag, format string, a ...interface{}) {
	golbalLogger.Output(tag, L_DEBUG, 2, fmt.Sprintf(format, a...))
}

// TagInfof 格式化info输出
func TagInfof(tag, format string, a ...interface{}) {
	golbalLogger.Output(tag, L_INFO, 2, fmt.Sprintf(format, a...))
}

// TagWarnningf 格式化warnning输出
func TagWarnningf(tag, format string, a ...interface{}) {
	golbalLogger.Output(tag, L_WARN, 2, fmt.Sprintf(format, a...))
}

// TagErrorf 格式化error输出
func TagErrorf(tag, format string, a ...interface{}) {
	golbalLogger.Output(tag, L_ERROR, 2, fmt.Sprintf(format, a...))
}

// Debug 调试输出
func Debug(a ...interface{}) {
	golbalLogger.TagDebug(golbalLogger.defaultTagName, a...)
}

// Info 普通信息
func Info(a ...interface{}) {
	golbalLogger.TagInfo(golbalLogger.defaultTagName, a...)
}

// Warnning 警告
func Warnning(a ...interface{}) {
	golbalLogger.TagWarnning(golbalLogger.defaultTagName, a...)
}

// Error 错误
func Error(a ...interface{}) {
	golbalLogger.TagError(golbalLogger.defaultTagName, a...)
}

// Debugf 格式化debug输出
func Debugf(format string, a ...interface{}) {
	golbalLogger.TagDebugf(golbalLogger.defaultTagName, format, a...)
}

// Infof 格式化info输出
func Infof(format string, a ...interface{}) {
	golbalLogger.TagInfof(golbalLogger.defaultTagName, format, a...)
}

// Warnningf 格式化warnning输出
func Warnningf(format string, a ...interface{}) {
	golbalLogger.TagWarnningf(golbalLogger.defaultTagName, format, a...)
}

// Errorf 格式化error输出
func Errorf(format string, a ...interface{}) {
	golbalLogger.TagErrorf(golbalLogger.defaultTagName, format, a...)
}
