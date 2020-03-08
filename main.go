package main

import (
	"flag"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"gitlab.com/seknox/trasa/trasadbproxy/proxy"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	logLevel := flag.String("l", "trace", "Set log level")
	logOutputToFile := flag.Bool("f", false, "Write to file")

	flag.Parse()
	level, err := logger.ParseLevel(*logLevel)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(level, false)

	if *logOutputToFile {
		f, err := os.OpenFile("/var/log/trasadbproxy.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		logger.SetOutput(f)
	} else {
		logger.SetOutput(os.Stdout)
	}

	logger.SetLevel(level)
	logger.SetReportCaller(true)

	logger.SetFormatter(&logger.TextFormatter{
		ForceColors:   false,
		DisableColors: false,
		//ForceQuote:                false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "",
		DisableSorting:            true,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		//PadLevelText:              false,
		QuoteEmptyFields: false,
		FieldMap:         nil,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return filepath.Base(frame.Function), fmt.Sprintf(`%s:%d`, filepath.Base(frame.File), frame.Line)
		},
	})

	proxy.StartListner()
}
