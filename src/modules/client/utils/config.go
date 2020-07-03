package utils

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/toolkits/file"
)

type GeneralConfig struct {
	Addresses string
	Token     string
}

type GlobalConfig struct {
	General *GeneralConfig
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
		notExistText := fmt.Sprintln("config file:", cfg, "is not existent. maybe you need `mv config.example.ini config.ini`")
		if runtime.GOOS == "linux" {
			cfg = "/usr/local/troop-client/conf/config.ini"
			if !file.IsExist(cfg) {
				log.Fatalln(notExistText)
			}
		} else {
			log.Fatalln(notExistText)
		}
	}

	ConfigFile = cfg

	if err := Load(cfg); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", cfg, err.Error())
		return
	}

	lock.Lock()
	defer lock.Unlock()

	GeneralConfig := GeneralConfig{}
	GeneralConfig.Addresses = GetString("general", "addresses")
	GeneralConfig.Token = GetString("general", "token")

	GlobalConfig := GlobalConfig{}
	GlobalConfig.General = &GeneralConfig

	config = &GlobalConfig
}
