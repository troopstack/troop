package utils

import (
	"gopkg.in/ini.v1"
)

type IniParser struct {
	conf_reader *ini.File
}

var (
	IniParserConf *IniParser
)

func Load(config_file_name string) error {
	conf, err := ini.Load(config_file_name)
	if err != nil {
		IniParserConf.conf_reader = nil
		return err
	}
	IParser := IniParser{}
	IParser.conf_reader = conf
	IniParserConf = &IParser
	return nil
}

func GetBool(section string, key string) bool {
	if IniParserConf.conf_reader == nil {
		return false
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return false
	}

	return s.Key(key).MustBool()
}

func GetString(section string, key string) string {
	if IniParserConf.conf_reader == nil {
		return ""
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return ""
	}

	return s.Key(key).String()
}

func GetInt(section string, key string, defaultValue int) int {
	if IniParserConf.conf_reader == nil {
		return defaultValue
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return defaultValue
	}

	value_int, _ := s.Key(key).Int()

	return int(value_int)
}

func GetInt32(section string, key string) int32 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_int, _ := s.Key(key).Int()

	return int32(value_int)
}

func GetUint32(section string, key string) uint32 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_int, _ := s.Key(key).Uint()

	return uint32(value_int)
}

func GetInt64(section string, key string) int64 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_int, _ := s.Key(key).Int64()
	return value_int
}

func GetUint64(section string, key string) uint64 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_int, _ := s.Key(key).Uint64()
	return value_int
}

func GetFloat32(section string, key string) float32 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_float, _ := s.Key(key).Float64()
	return float32(value_float)
}

func GetFloat64(section string, key string) float64 {
	if IniParserConf.conf_reader == nil {
		return 0
	}

	s := IniParserConf.conf_reader.Section(section)
	if s == nil {
		return 0
	}

	value_float, _ := s.Key(key).Float64()
	return value_float
}
