package setting

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gitee.com/azhai/xorm-refactor/setting/dialect"
	"github.com/azhai/gozzo-utils/filesystem"
	json "github.com/goccy/go-json"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/yaml.v3"
	"xorm.io/xorm"
)

const (
	DEFAULT_DIR_MODE  = 0o755
	DEFAULT_FILE_MODE = 0o644
)

type IReverseConfig interface {
	GetReverseTarget(name string) ReverseTarget
	GetConnConfigMap(keys ...string) map[string]ConnConfig
	GetConnConfig(key string) (ConnConfig, bool)
}

type ConnConfig struct {
	DriverName  string             `json:"driver_name" yaml:"driver_name"`
	ReadOnly    bool               `json:"read_only" yaml:"read_only"`
	TablePrefix string             `json:"table_prefix" yaml:"table_prefix"`
	LogFile     string             `json:"log_file" yaml:"log_file"`
	Params      dialect.ConnParams `json:"params" yaml:"params"`
}

func (c ConnConfig) ConnectXorm(verbose bool) (*xorm.Engine, error) {
	d := dialect.GetDialectByName(c.DriverName)
	dsn := d.ParseDSN(c.Params)
	engine, err := xorm.NewEngine(c.DriverName, dsn)
	if err == nil {
		engine.ShowSQL(verbose)
	}
	return engine, err
}

func (c ConnConfig) ConnectRedis(verbose bool) (redis.Conn, error) {
	d := new(dialect.Redis)
	addr := d.ParseDSN(c.Params)
	return redis.Dial("tcp", addr, d.GetOptions()...)
}

type Configure struct {
	Debug         bool                  `json:"debug" yaml:"debug"`
	Connections   map[string]ConnConfig `json:"connections" yaml:"connections"`
	ReverseTarget ReverseTarget         `json:"reverse_target" yaml:"reverse_target"`
}

func ReadSettingsFrom(fileType, fileName string, cfg interface{}) error {
	var err error
	switch fileType {
	case "json", "Json", "JSON":
		var content []byte
		content, err = ioutil.ReadFile(fileName)
		if err == nil {
			err = json.Unmarshal(content, &cfg)
		}
	case "yml", "yaml", "Yaml", "YAML":
		var fp *os.File
		fp, err = os.Open(fileName)
		if err == nil {
			err = yaml.NewDecoder(fp).Decode(cfg)
		}
	}
	return err
}

func ReadSettingsExt(fileName string, cfg interface{}) (string, error) {
	pos := strings.LastIndex(fileName, ".")
	fileExt := strings.ToLower(fileName[pos+1:])
	if fileExt == "yml" || fileExt == "yaml" || fileExt == "json" {
		return fileExt, ReadSettingsFrom(fileExt, fileName, cfg)
	}
	size, exists := filesystem.FileSize(fileName + ".yml")
	if exists || size > 0 {
		return "yml", ReadSettingsFrom("yml", fileName+".yml", cfg)
	}
	size, exists = filesystem.FileSize(fileName + ".yaml")
	if exists || size > 0 {
		return "yaml", ReadSettingsFrom("yaml", fileName+".yaml", cfg)
	}
	size, exists = filesystem.FileSize(fileName + ".json")
	if exists || size > 0 {
		return "json", ReadSettingsFrom("json", fileName+".json", cfg)
	}
	return "", fmt.Errorf("Unknow settings file")
}

func ReadSettings(fileName, nameSpace string) (*Configure, error) {
	cfg, ext, err := new(Configure), "", fmt.Errorf("settings is empty")
	if fileName == "" {
		fileName, ext = "./settings.yml", "yml"
	} else {
		ext, err = ReadSettingsExt(fileName, &cfg)
	}
	if err != nil {
		cfg.ReverseTarget = DefaultMixinReverseTarget(nameSpace)
	}
	if cfg.Connections == nil {
		dbFileName := strings.Replace(fileName, "settings."+ext, "databases", 1)
		_, err = ReadSettingsExt(dbFileName, &cfg.Connections)
	}
	if err == nil && len(cfg.Connections) > 0 {
		cfg.RemovePrivates()
	}
	return cfg, err
}

func SaveSettingsTo(fileName string, cfg interface{}) error {
	wt, err := os.Open(fileName)
	if err == nil {
		err = yaml.NewEncoder(wt).Encode(cfg)
	}
	return err
}

func Settings2Bytes(cfg interface{}) []byte {
	buf := new(bytes.Buffer)
	err := yaml.NewEncoder(buf).Encode(cfg)
	if err == nil {
		return buf.Bytes()
	}
	return nil
}

func (cfg Configure) GetReverseTarget(name string) ReverseTarget {
	if name == "*" || name == cfg.ReverseTarget.Language {
		return cfg.ReverseTarget
	}
	return ReverseTarget{OutputDir: "/dev/null"}
}

func (cfg Configure) RemovePrivates() {
	for key := range cfg.Connections {
		if strings.HasPrefix(key, "_") {
			delete(cfg.Connections, key)
		}
	}
}

func (cfg Configure) GetConnConfigMap(keys ...string) map[string]ConnConfig {
	if len(keys) == 0 {
		return cfg.Connections
	}
	result := make(map[string]ConnConfig)
	for _, k := range keys {
		if c, ok := cfg.Connections[k]; ok {
			result[k] = c
		}
	}
	return result
}

func (cfg Configure) GetConnConfig(key string) (ConnConfig, bool) {
	if c, ok := cfg.Connections[key]; ok {
		return c, true
	}
	return ConnConfig{}, false
}
