package utils

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/toolkits/file"
)

var (
	LocalIp string
	Root    string
	PkiRoot string

	ScoutPkiRoot     string
	ScoutPubFilename string
	ScoutPriFilename string

	GeneralPkiRoot     string
	GeneralPubFilename string

	PluginParentDir string
	PluginDirFile   string
	Plugins         = make(map[string]interface{})

	TaskCacheDir string
	FileCacheDir string

	GeneralIgnoreCommands []string

	logger *log.Logger
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
	ScoutPubFilename = ScoutPkiRoot + "/scout.pub"
	ScoutPriFilename = ScoutPkiRoot + "/scout.pem"
	GeneralPubFilename = GeneralPkiRoot + "/general.pub"
	TaskCacheDir = path.Join(Root, "taskCache")
	FileCacheDir = path.Join(Root, "fileCache")

	PluginParentDir = Config().Plugin.Dir
	file.InsureDir(PluginParentDir)
	PluginDirFile = PluginParentDir + "/plugin"
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
	fileAndStdoutWriter := io.MultiWriter(writers...)
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

func Logger() *log.Logger {
	lock.RLock()
	defer lock.RUnlock()
	return logger
}

func InitLocalIp() {
	conn, err := net.DialTimeout("udp", "google.com:80", time.Second*10)
	if err != nil {
		log.Println("get local addr failed !")
	} else {
		LocalIp = strings.Split(conn.LocalAddr().String(), ":")[0]
		conn.Close()
	}
}
