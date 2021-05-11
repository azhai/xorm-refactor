package cmd

import (
	"fmt"

	"gitee.com/azhai/xorm-refactor/setting"
	"github.com/azhai/gozzo-utils/filesystem"
)

const (
	VERSION = "1.1.0"
)

var (
	settings *setting.Configure
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

func Prepare(fileName, nameSpace string) (*setting.Configure, error) {
	settings, err := setting.ReadSettings(fileName, nameSpace)
	if settings != nil {
		verbose = settings.Debug
	} else if err == nil {
		err = fmt.Errorf("settings is empty")
	}
	return settings, err
}

func Verbose() bool {
	return verbose
}
