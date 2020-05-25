package cmd

import (
	"fmt"

	"gitee.com/azhai/xorm-refactor/config"
	"github.com/azhai/gozzo-utils/filesystem"
)

const (
	VERSION = "1.0.2"
)

var (
	settings *config.Settings
	verbose  bool // 详细输出
)

// 逐个尝试，找出第一个存在的文件
func FindFirstFile(fileNames []string, minSize int64) string {
	if len(fileNames) == 0 {
		return ""
	}
	for _, f := range fileNames { // 找到第一个存在的文件
		size, exists := filesystem.FileSize(f)
		if exists && size >= minSize {
			return f
		}
	}
	return ""
}

func Prepare(configFiles []string) (*config.Settings, error) {
	fileName := FindFirstFile(configFiles, 0)
	if fileName == "" {
		err := fmt.Errorf("need reverse file")
		return nil, err
	}
	settings, err := config.ReadSettings(fileName)
	if settings != nil {
		verbose = settings.Application.Debug
	} else if err == nil {
		err = fmt.Errorf("settings is empty")
	}
	return settings, err
}

func Verbose() bool {
	return verbose
}
