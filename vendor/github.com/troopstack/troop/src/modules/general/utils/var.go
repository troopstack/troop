package utils

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Root               string
	PkiRoot            string
	ScoutPkiRoot       string
	GeneralPkiRoot     string
	GeneralPubFilename string
	GeneralPriFilename string
)

func InitRootDir() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		log.Fatalln("GetWD fail:", err)
	}
	PkiRoot = Root + "/pki"
	ScoutPkiRoot = PkiRoot + "/scout"
	GeneralPkiRoot = PkiRoot + "/general"
	GeneralPubFilename = GeneralPkiRoot + "/general.pub"
	GeneralPriFilename = GeneralPkiRoot + "/general.pem"
}

func InitLog(level string) {
	fileName := Config().Log.Logfile
	logFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("open file error !")
	}
	prefix := fmt.Sprintf("[%s] ", level)

	log.Println("logging on", fileName)
	writers := []io.Writer{
		logFile,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...,
	)
	log.SetOutput(fileAndStdoutWriter) // 设置输出流

	log.SetPrefix(prefix) // 日志前缀
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func InitDir() {
	dirs := []string{PkiRoot, GeneralPkiRoot, ScoutPkiRoot}
	for _, value := range dirs {
		err := CreateDir(value)
		if err != nil {
			panic(err)
		}
	}
}
