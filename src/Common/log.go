package Common

import (
	"bytes"
	"fmt"
	"time"
)

type Log struct {
	LogLevel logLevel //日志等级
}

type logLevel int8

const (
	_ = iota
	DEBUGLEVEL = 1
	INFOLEVEL = 2
	ERRORLEVEL = 3
)

type logColor int8

const (
	_ = iota
	DEBUGCOLOR = 1
	INFOCOLOR = 2
	ERRORCOLOR = 3
)

var LogInstance *Log

func GetInstanceLog() *Log {
	if LogInstance == nil {
		LogInstance = &Log{DEBUGLEVEL}
	}
	return LogInstance
}

func InitLog() {
	GetInstanceLog()
}

func (this *Log) SetLogLevel(LogLevel logLevel) {
	this.LogLevel = LogLevel
}

func (this *Log) DoWrite(str string,color int8) { //颜色选项暂时不用了
	var buffer bytes.Buffer
	buffer.WriteString(ToString(time.Now().Unix()))
	buffer.WriteString("： ")
	buffer.WriteString(str)
	resultStr := buffer.String()
	fmt.Sprintln(resultStr)
}

func Debug(strs ...string) {
	Instance := GetInstanceLog()
	if Instance.LogLevel <= DEBUGLEVEL {
		result := ""
		for _,str := range strs{
			result += str
		}
		Instance.DoWrite("INFO " + result, DEBUGCOLOR)
	}
}

func Info(strs ...string) {
	Instance := GetInstanceLog()
	if Instance.LogLevel <= INFOLEVEL {
		result := ""
		for _,str := range strs{
			result += str
		}
		Instance.DoWrite("INFO " + result, INFOCOLOR)
	}
}

func Error(strs ...string) {
	Instance := GetInstanceLog()
	if Instance.LogLevel <= ERRORLEVEL {
		result := ""
		for _,str := range strs{
			result += str
		}
		Instance.DoWrite("ERROR " + result, ERRORCOLOR)
	}
}