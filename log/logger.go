package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LevelDebug = 1
	LevelInfo  = 2
	LevelWarn  = 3
	LevelError = 4
	LevelFatal = 5
)

var (
	logLevel = LevelDebug
)

func SetLevel(level int) {
	logLevel = level
}

var cwdLen int

func init() {
	_, filename, _, _ := runtime.Caller(0)
	cwdLen = len(strings.ReplaceAll(filename, "utils/log/logger.go", ""))
}

func formatLog(level string, format string, a ...interface{}) string {
	prefix := "[SRV] " + time.Now().Format("2006/01/02 - 15:04:05")
	content := ""
	if format == "" {
		content = fmt.Sprint(a...)
	} else {
		content = fmt.Sprintf(format, a...)
	}
	suffix := ""
	if !strings.HasSuffix(content, "\n") {
		suffix = "\n"
	}
	_, filename, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s %s ./%s:%d %s%s", prefix, level, filename[cwdLen:], line, content, suffix)
}

func Debug(a ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("DEBUG", "", a...))

}
func Debugf(format string, a ...interface{}) {
	if logLevel > LevelDebug {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("DEBUG", format, a...))

}
func Info(a ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("INFO", "", a...))

}
func Infof(format string, a ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("INFO", format, a...))

}

func Warn(a ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("WARN", "", a...))

}
func Warnf(format string, a ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("WARN", format, a...))

}

func Error(a ...interface{}) {
	if logLevel > LevelError {
		return
	}
	fmt.Fprint(os.Stderr, formatLog("ERROR", "", a...))
}

//错误
func Errorf(format string, a ...interface{}) {
	if logLevel > LevelError {
		return
	}
	fmt.Fprint(os.Stderr, formatLog("ERROR", format, a...))
}

//严重错误,应用无法继续运行
func Fatal(a ...interface{}) {
	fmt.Fprint(os.Stderr, formatLog("FATAL", "", a...))
	os.Exit(1)
}

//发送到TG群日志
func TGInfof(format string, a ...interface{}) {
	if logLevel > LevelInfo {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("INFO", format, a...))
	SendToBot(format, a...)
}

func TGWarnf(format string, a ...interface{}) {
	if logLevel > LevelWarn {
		return
	}
	fmt.Fprint(os.Stdout, formatLog("WARN", format, a...))
	SendToBot(format, a...)
}

func TGErrorf(format string, a ...interface{}) {
	if logLevel > LevelError {
		return
	}
	fmt.Fprint(os.Stderr, formatLog("ERROR", format, a...))
	SendToBot(format, a...)
}

func TGFatal(format string, a ...interface{}) {
	fmt.Fprint(os.Stderr, formatLog("FATAL", format, a...))
	os.Exit(1)
	SendToBot(format, a...)
}
