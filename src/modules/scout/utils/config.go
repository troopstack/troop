package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/toolkits/file"
)

type DebugConfig struct {
	Enabled bool
}

type HostConfig struct {
	Hostname string
	Ip       string
	Tag      []string
}

type GeneralConfig struct {
	Addresses []string
	Timeout   int
	Token     string
}

type MQConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	VHost    string
}

type LogConfig struct {
	Logfile string
}

type PluginConfig struct {
	Dir           string
	Plugins       []string
	CustomPlugins []string
}

type IgnoreCommandConfig struct {
	Commands []string
}

type GlobalConfig struct {
	Debug         *DebugConfig
	Host          *HostConfig
	General       *GeneralConfig
	Log           *LogConfig
	MQ            *MQConfig
	Plugin        *PluginConfig
	IgnoreCommand *IgnoreCommandConfig
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

func MQ() (*MQConfig, error) {
	user := Config().MQ.User
	password := Config().MQ.Password
	host := Config().MQ.Host
	port := Config().MQ.Port
	if user == "" || password == "" || host == "" || port == "" {
		return nil, errors.New("RabbitMQ configuration is abnormal")
	}
	return Config().MQ, nil
}

func Hostname() (string, error) {
	hostname := Config().Host.Hostname
	if hostname != "" {
		return hostname, nil
	}

	if os.Getenv("SCOUT_HOSTNAME") != "" {
		hostname = os.Getenv("SCOUT_HOSTNAME")
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	return hostname, err
}

func IP() string {
	ip := Config().Host.Ip
	if ip != "" {
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}

	return ip
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

	HostConfig := HostConfig{}
	HostConfig.Hostname = GetString("host", "hostname")
	HostConfig.Ip = GetString("host", "ip")
	tags := GetString("host", "tag")
	HostConfig.Tag = strings.Split(tags, ",")

	GeneralConfig := GeneralConfig{}
	addresses := GetString("general", "addresses")
	GeneralConfig.Addresses = strings.Split(addresses, ",")
	GeneralConfig.Timeout = GetInt("general", "timeout", 1000)
	GeneralConfig.Token = GetString("general", "token")

	MQConfig := MQConfig{}
	MQConfig.User = GetString("rabbit_mq", "user")
	MQConfig.Password = GetString("rabbit_mq", "password")
	MQConfig.Host = GetString("rabbit_mq", "host")
	MQConfig.Port = GetString("rabbit_mq", "port")
	MQConfig.VHost = GetString("rabbit_mq", "vhost")

	LogConfig := LogConfig{}
	LogConfig.Logfile = GetString("log", "logfile")

	PluginConfig := PluginConfig{}
	PluginConfig.Dir = GetString("plugin", "dir")
	plugins := GetString("plugin", "plugins")
	pluginList := []string{"scout_manager"}
	if plugins != "" {
		plugins = strings.TrimSpace(plugins)
		customPlugins := strings.Split(plugins, ",")
		PluginConfig.CustomPlugins = customPlugins
		pluginList = append(pluginList, customPlugins...)
	}
	PluginConfig.Plugins = pluginList

	IgnoreCommandConfig := IgnoreCommandConfig{}
	ignoreCommandFile := GetString("ignore_command", "file")
	if ignoreCommandFile != "" {
		f, err := os.Open(ignoreCommandFile)
		if err != nil {
			log.Fatalf("file: %s not exist", ignoreCommandFile)
		}
		defer f.Close()
		rd := bufio.NewReader(f)
		for {
			line, err := rd.ReadString('\n')
			line = strings.TrimSpace(line)
			line = strings.Replace(line, "\\n", "", -1)
			IgnoreCommandConfig.Commands = append(IgnoreCommandConfig.Commands, line)
			if err != nil || io.EOF == err {
				break
			}
		}
	}

	GlobalConfig := GlobalConfig{}
	GlobalConfig.Debug = &DebugConfig
	GlobalConfig.Host = &HostConfig
	GlobalConfig.General = &GeneralConfig
	GlobalConfig.MQ = &MQConfig
	GlobalConfig.Log = &LogConfig
	GlobalConfig.Plugin = &PluginConfig
	GlobalConfig.IgnoreCommand = &IgnoreCommandConfig
	config = &GlobalConfig

	log.Println("read config file:", cfg, "successfully")
}
