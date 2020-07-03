package utils

import (
	"bufio"
	"errors"
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

type MySQLConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	DB                string
	Charset           string
	MaxIdleConnection int
	MaxOpenConnection int
}

type MQConfig struct {
	User             string
	Password         string
	Host             string
	Port             string
	VHost            string
	MaxConnectionNum int
	MaxChannelNum    int
}

type RpcConfig struct {
	Listen string
}

type HttpConfig struct {
	Listen string
	Token  string
}

type FileConfig struct {
	Address string
}

type LogConfig struct {
	Logfile string
}

type ScoutConfig struct {
	AutoAccept bool
}

type PluginConfig struct {
	Enabled bool
	Git     string
}

type IgnoreCommandConfig struct {
	Commands []string
}

type GlobalConfig struct {
	Debug         *DebugConfig
	MySQL         *MySQLConfig
	Rpc           *RpcConfig
	MQ            *MQConfig
	Http          *HttpConfig
	File          *FileConfig
	Log           *LogConfig
	Scout         *ScoutConfig
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

func MySQL() (*MySQLConfig, error) {
	user := Config().MySQL.User
	host := Config().MySQL.Host
	port := Config().MySQL.Port
	db := Config().MySQL.DB
	if user == "" || host == "" || port == "" || db == "" {
		return nil, errors.New("MySQL configuration is abnormal")
	}
	return Config().MySQL, nil
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
		log.Fatalf("try load config file[%s] error[%s]\n", cfg, err.Error())
	}

	lock.Lock()
	defer lock.Unlock()

	DebugConfig := DebugConfig{}
	DebugConfig.Enabled = GetBool("debug", "enabled")

	MySQLConfig := MySQLConfig{}
	MySQLConfig.User = GetString("mysql", "user")
	MySQLConfig.Password = GetString("mysql", "password")
	MySQLConfig.Host = GetString("mysql", "host")
	MySQLConfig.Port = GetString("mysql", "port")
	MySQLConfig.DB = GetString("mysql", "db")
	mysql_charset := GetString("mysql", "charset")
	if mysql_charset == "" {
		mysql_charset = "utf8"
	}
	MySQLConfig.Charset = mysql_charset
	MySQLConfig.MaxIdleConnection = GetInt("mysql", "max_idle_connection", 10)
	MySQLConfig.MaxOpenConnection = GetInt("mysql", "max_open_connection", 100)

	MQConfig := MQConfig{}
	MQConfig.User = GetString("rabbit_mq", "user")
	MQConfig.Password = GetString("rabbit_mq", "password")
	MQConfig.Host = GetString("rabbit_mq", "host")
	MQConfig.Port = GetString("rabbit_mq", "port")
	MQConfig.VHost = GetString("rabbit_mq", "vhost")
	MQConfig.MaxConnectionNum = GetInt("rabbit_mq", "max_connection_num", 10)
	MQConfig.MaxChannelNum = GetInt("rabbit_mq", "max_channel_num", 10)

	LogConfig := LogConfig{}
	LogConfig.Logfile = GetString("log", "logfile")

	RpcConfig := RpcConfig{}
	RpcConfig.Listen = GetString("rpc", "listen")

	HttpConfig := HttpConfig{}
	HttpConfig.Listen = GetString("http", "listen")
	HttpConfig.Token = GetString("http", "token")

	ScoutConfig := ScoutConfig{}
	ScoutConfig.AutoAccept = GetBool("scout", "auto_accept")

	FileConfig := FileConfig{}
	FileConfig.Address = GetString("file", "address")

	PluginConfig := PluginConfig{}
	PluginConfig.Enabled = true
	PluginConfig.Git = GetString("plugin", "git")

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
	GlobalConfig.MySQL = &MySQLConfig
	GlobalConfig.MQ = &MQConfig
	GlobalConfig.Log = &LogConfig
	GlobalConfig.Rpc = &RpcConfig
	GlobalConfig.Http = &HttpConfig
	GlobalConfig.Scout = &ScoutConfig
	GlobalConfig.File = &FileConfig
	GlobalConfig.Plugin = &PluginConfig
	GlobalConfig.IgnoreCommand = &IgnoreCommandConfig
	config = &GlobalConfig
	log.Println("read config file:", cfg, "successfully")
}
