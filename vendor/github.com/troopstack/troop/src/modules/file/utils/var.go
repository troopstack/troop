package utils

import (
	"fmt"
	"log"
	"os"
)

var (
	Root     string
	FileRoot string
)

func InitRootDir() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		log.Fatalln("GetWD fail:", err)
	}
	FileRoot = Root + "/files"
}

func InitLog(level string) {
	fileName := Config().Log.Logfile
	logFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("open file error !")
	}
	prefix := fmt.Sprintf("[%s] ", level)
	if level != "debug" {
		log.Println("logging on", fileName)
		log.SetOutput(logFile) // 设置输出流
	}
	log.SetPrefix(prefix) // 日志前缀
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func InitDir() {
	dirs := []string{FileRoot}
	for _, value := range dirs {
		err := CreateDir(value)
		if err != nil {
			panic(err)
		}
	}
}
