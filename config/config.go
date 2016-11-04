package config

import (
	"bufio"
	"os"
	"strings"
)

type ConfigSection map[string][]string
type configFile map[string]ConfigSection

var defConfig *configFile

// 根据名字获取对应的配置文件
func SetConfig(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	//解析ini文件
	r := bufio.NewReader(f)
	var (
		line string
		sec  string
	)
	sections := make(configFile)
	for err == nil {
		line, err = r.ReadString('\n')
		line = strings.TrimSpace(line)
		//空行或者注释跳过 注释支持;和#开头的行
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		//判断配置块[name]
		if line[0] == '[' && line[len(line)-1] == ']' {
			sec = line[1 : len(line)-1]
			_, has := sections[sec]
			if !has {
				sections[sec] = make(ConfigSection)
			}
			continue
		}
		if sec == "" {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key, val := strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])
		if key == "" || val == "" {
			continue
		}
		if slice, has := sections[sec][key]; has {
			slice = append(slice, val)
			sections[sec][key] = slice
		} else {
			sections[sec][key] = []string{val}
		}
	}
	defConfig = &sections
	return nil
}

func GetSection(sec string) ConfigSection {
	if defConfig == nil {
		return nil
	}
	return (*defConfig)[sec]
}

func GetValueSlice(sec, key string) []string {
	if seco := GetSection(sec); seco == nil {
		return []string{}
	} else {
		return seco.GetValueSlice(key)
	}
}

func GetValue(sec, key string) string {
	s := GetValueSlice(sec, key)
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

func (m ConfigSection) GetValueSlice(key string) []string {
	if m[key] != nil {
		return m[key]
	}
	return []string{}
}

func (m ConfigSection) GetAllKeyMapString() map[string]string {
	newm := make(map[string]string)
	for k, v := range m {
		newm[k] = v[0]
	}
	return newm
}

func (m ConfigSection) GetValue(key string) string {
	s := m.GetValueSlice(key)
	if len(s) > 0 {
		return s[0]
	}
	return ""
}
