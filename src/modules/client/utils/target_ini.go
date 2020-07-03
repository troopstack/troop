package utils

import (
	"bufio"
	"io"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

type TargetIniParser struct {
	conf_reader *ini.File
}

var (
	TargetIniParserConf *TargetIniParser
)

func TargetIniLoad(config_file_name string) error {
	f, _ := os.Open(config_file_name)
	data := ""
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "\\n", "", -1)
		if line != "" && !strings.HasPrefix(line, "[") && !strings.HasSuffix(line, "]") {
			line = strings.TrimLeft(line, ",")
			line = strings.TrimRight(line, ",")
			line = line + "="
			line = strings.Replace(line, ",", "=\n", -1)
		}
		data += line + "\n"
		if err != nil || io.EOF == err {
			break
		}
	}
	f.Close()
	conf, err := ini.Load([]byte(data))
	if err != nil {
		return err
	}
	IParser := TargetIniParser{}
	IParser.conf_reader = conf
	TargetIniParserConf = &IParser
	return nil
}

func removeDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func TargetIniGetAllKeys() (keys []string) {
	if TargetIniParserConf.conf_reader == nil {
		return
	}
	for _, v := range TargetIniParserConf.conf_reader.Sections() {
		keys = append(keys, v.KeyStrings()...)
	}
	keys = removeDuplicateElement(keys)
	return
}

func TargetIniGetKeys(section string) (keys []string) {
	if TargetIniParserConf.conf_reader == nil {
		return
	}
	section = strings.Replace(section, " ", "", -1)
	sections := strings.Split(section, ",")

	for i := range sections {
		s := TargetIniParserConf.conf_reader.Section(sections[i])
		if s == nil {
			continue
		}
		keys = append(keys, s.KeyStrings()...)
	}
	keys = removeDuplicateElement(keys)
	return
}
