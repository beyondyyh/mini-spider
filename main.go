package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/baidu/go-lib/log"
	"github.com/baidu/go-lib/log/log4go"
)

const (
	Version            = "v1.0"
	SpiderConfFileName = "spider.conf"
)

var (
	confPath = flag.String("c", "./conf", "root path of configuration")
	logPath  = flag.String("l", "./log", "dir path of log")
	help     = flag.Bool("h", false, "show help")
	version  = flag.Bool("v", false, "show version")
)

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	if *version {
		fmt.Println(Version)
		return
	}

	err := initLog("INFO", *logPath)
	if err != nil {
		fmt.Printf("initLog(): %s\n", err.Error())
		Exit(-1)
	}

	config, err := ConfigLoad(filepath.Join(*confPath, SpiderConfFileName))
	if err != nil {
		log.Logger.Error("ConfigLoad(): %s", err.Error())
		Exit(-1)
	}

	seeds, err := SeedsLoad(config.UrlListFile)
	if err != nil {
		log.Logger.Error("SeedsLoad(): %s", err.Error())
		Exit(-1)
	}

	spider := NewScheduler(config, seeds)
	spider.Start()

	Exit(0)
}

func initLog(level, path string) error {
	// log4go.SetLogBufferLength(1000)
	log4go.SetLogWithBlocking(false)
	if err := log.Init("mini-spider", level, path, true, "D", 5); err != nil {
		return fmt.Errorf("err in log.Init(): %s", err.Error())
	}
	return nil
}

func Exit(code int) {
	log.Logger.Close()
	time.Sleep(100 * time.Millisecond)
	os.Exit(code)
}
