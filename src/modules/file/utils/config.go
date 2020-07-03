package utils

import (
	"fmt"
	"log"
	"sync"

	"github.com/toolkits/file"
)

type DebugConfig struct {
	Enabled bool
}

type HttpConfig struct {
	Listen string
	Token  string
}

type LogConfig struct {
	Logfile string
}

type GlobalConfig struct {
	Debug *DebugConfig
	Http  *HttpConfig
	Log   *LogConfig
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv config.example.ini config.ini`")
	}

	ConfigFile = cfg

	if err := Load(cfg); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", cfg, err.Error())
		return
	}

	lock.Lock()
	defer lock.Unlock()

	DebugConfig := DebugConfig{}
	DebugConfig.Enabled = GetBool("debug", "enabled")

	LogConfig := LogConfig{}
	LogConfig.Logfile = GetString("log", "logfile")

	HttpConfig := HttpConfig{}
	HttpConfig.Listen = GetString("http", "listen")
	HttpConfig.Token = GetString("http", "token")

	GlobalConfig := GlobalConfig{}
	GlobalConfig.Debug = &DebugConfig
	GlobalConfig.Log = &LogConfig
	GlobalConfig.Http = &HttpConfig
	config = &GlobalConfig

	log.Println("read config file:", cfg, "successfully")
}
