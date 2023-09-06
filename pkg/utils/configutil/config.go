package configutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

func NewConfigFromFile(filePath string, target any) error {
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("file %s not exist", filePath))
	}
	// 如果文件大小为0，直接返回
	if fi.Size() == 0 {
		return nil
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("error opening configutil file: %s", err))
	}
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(target); err != nil {
			return errors.New(fmt.Sprintf("error parse configutil file: %s", err))
		}
	} else if strings.HasSuffix(filePath, ".json") {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(target); err != nil {
			return errors.New(fmt.Sprintf("error parse configutil file: %s", err))
		}
	}
	return nil
}
