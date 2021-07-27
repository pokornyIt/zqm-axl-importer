package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
)

const (
	identificationId      = "id"
	identificationIdOrder = "00__id__"
)

func validLogLevel(level string) log.Level {
	switch strings.ToUpper(level) {
	case "FAT", "F", "FATAL":
		return log.FatalLevel
	case "ERR", "E", "ERROR":
		return log.ErrorLevel
	case "WAR", "W", "WARNING":
		return log.WarnLevel
	case "INF", "I", "INFO":
		return log.InfoLevel
	case "TRC", "T", "TRACE":
		return log.TraceLevel
	case "DEB", "D", "DEBUG":
		return log.DebugLevel
	default:
		return log.InfoLevel
	}
}

func prettyFile(f *runtime.Frame) (string, string) {
	skip := 1
	var routine string
	var file string
	var line int
	var callFunc string
	var ok bool
	var pc uintptr
	for {
		pc, file, line, ok = runtime.Caller(skip)
		if skip > 10 {
			ok = false
		}
		if !ok {
			file = "---"
			routine = "---"
			line = 0
			callFunc = "---"
			break
		}
		if strings.Contains(file, "logrus") || strings.Contains(runtime.FuncForPC(pc).Name(), "logrus") {
			skip++
			continue
		}
		slash := strings.LastIndex(file, "/")
		routine = file[slash+1:]
		callFunc = runtime.FuncForPC(pc).Name()
		slash = strings.LastIndex(callFunc, ".")
		callFunc = callFunc[slash+1:]
		break
	}
	format := " %s:%d"
	if config.Log.LogToFile() {
		format = "%s:%d"
	}
	return fmt.Sprintf("%s()", callFunc), fmt.Sprintf(format, routine, line)
}

func sortLogFields(i []string) {
	if len(i) < 2 {
		return
	}
	idx := -1
	for j, s := range i {
		if s == identificationId {
			idx = j
		}
	}
	if idx > -1 && idx < len(i) {
		i[idx] = identificationIdOrder
	}
	sort.Strings(i)
	idx = -1
	for j, s := range i {
		if s == identificationIdOrder {
			idx = j
		}
	}
	if idx > -1 && idx < len(i) {
		i[idx] = identificationId
	}
}

func initLog() {
	if config.Log.JSONFormat {
		jsonFormatter := new(log.JSONFormatter)
		jsonFormatter.TimestampFormat = DateTimeFormat
		jsonFormatter.CallerPrettyfier = prettyFile
		log.SetFormatter(jsonFormatter)
	} else {
		Formatter := new(log.TextFormatter)
		Formatter.TimestampFormat = DateTimeFormat
		Formatter.FullTimestamp = true
		Formatter.DisableLevelTruncation = false
		Formatter.ForceColors = !config.Log.LogToFile()
		Formatter.SortingFunc = sortLogFields
		Formatter.CallerPrettyfier = prettyFile
		log.SetFormatter(Formatter)
	}

	log.SetReportCaller(config.Log.LogProgramInfo)
	lvl := validLogLevel(config.Log.Level)
	log.SetLevel(lvl)
	if config.Log.LogToFile() {
		lJack := &lumberjack.Logger{
			Filename:   config.Log.FileName,
			MaxBackups: config.Log.MaxBackups,
			MaxAge:     config.Log.MaxAge,
			MaxSize:    config.Log.MaxSize,
			Compress:   true,
		}
		if config.Log.Quiet {
			log.SetOutput(lJack)
		} else {
			mWriter := io.MultiWriter(os.Stdout, lJack)
			log.SetOutput(mWriter)
		}
	} else {
		if config.Log.Quiet {
			log.SetLevel(log.PanicLevel)
		}
	}

	log.WithFields(log.Fields{
		"ApplicationName": applicationName,
		"RuntimeVersion":  runtime.Version(),
		"CPUs":            runtime.NumCPU(),
		"Arch":            runtime.GOARCH,
	}).Info("Application Initializing")
}

func VersionDetail() string {
	return fmt.Sprintf("Version details\r\n\tApplication Name: %s\r\n\tRuntime Version: %s\r\n\tCPUs: %d\r\n\tArchitectire: %s",
		applicationName, runtime.Version(), runtime.NumCPU(), runtime.GOARCH)
}
