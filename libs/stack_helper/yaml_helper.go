package stack_helper

import (
	"gopkg.in/yaml.v3"
	"os"
)

func ParseYAML(filename string) (map[string]map[string]string, error) {
	// 開啟檔案
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var content = make([]byte, stat.Size())
	f.Read(content)
	// 解析 yaml
	var out map[string]map[string]string
	err1 := yaml.Unmarshal(content, &out)
	if err != nil {
		return nil, err1
	}
	return out, nil
}
