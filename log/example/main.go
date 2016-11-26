package main

import (
	"fmt"

	"github.com/go-irain/tools/log"
)

func main() {
	// 设置日志显示最低等级 默认L_DEBUG
	log.SetLevel(log.L_DEBUG)
	// 设置是否显示行号 默认显示
	log.SetShowLineNumber(true)
	// 设置日志输出文件夹选项 默认直接输出到控制台
	// log.SetOutDirConfig("log", 100, 10)
	// 设置标签前缀 默认会自动添加应用的名称作为前缀
	log.SetTags("encode", "decode")
	// 设置钩子对error等级的日志进行处理
	log.SetHook(log.L_ERROR, func(tag, message string) {
		fmt.Println("hook >>>", tag, message)
	})

	// 各种等级的输出方法 默认输出到 tag为应用名称的log文件里
	// log.Debug("debug message")
	// log.Info("info message")
	// log.Warnning("warnning message")
	// log.Error("error message")

	// 指定输出到已注册的tag文件中 如果tag没有被注册 则输出到默认tag中
	// log.TagDebug("encode", "encode debug message")
	// log.TagInfo("encode", "encode info message")
	// log.TagWarnning("decode", "decode warnning message")
	// log.TagError("decode", "decode error message")

	// 格式化输出
	log.Error("my age is", 30)
	// output: 2010/10/11 12:00:01 [DBUG] <example> main:36 my age is 30
	log.Debugf("my name is %s", "afocus")
	// output: 2010/10/11 12:00:01 [DBUG] <example> main:38 my name is afocus
	log.TagDebug("request", "ip:127.0.0.1 method:POST body:hello,world")
	// output: 2010/10/11 12:00:01 [DBUG] <request> main:40 ip:127.0.0.1 method:POST body:hello,world

	// 创建新的log对象 不适用全局log对象
	newlog := log.NewLog()
	newlog.SetShowLineNumber(false)
	newlog.Info("new log")

}
