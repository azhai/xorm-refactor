package setting

import (
	"bytes"
	"os"

	"gitee.com/azhai/xorm-refactor/setting/dialect"
	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_DIR_MODE  = 0o755
	DEFAULT_FILE_MODE = 0o644
)

type IConnectConfig interface {
	GetConnConfigMap(keys ...string) map[string]ConnConfig
	GetConnConfig(key string) (ConnConfig, bool)
}

type IReverseConfig interface {
	GetReverseTargets() []ReverseTarget
	IConnectConfig
}

type AppConfig struct {
	Debug       bool `json:"debug" yaml:"debug"`
	PluralTable bool `json:"plural_table" yaml:"plural_table"`
}

type LogConfig struct {
	AccessFile string `json:"access_file" yaml:"access_file"`
	ErrorFile  string `json:"error_file" yaml:"error_file"`
	SqlFile    string `json:"sql_file" yaml:"sql_file"`
}

type PartConfig struct {
	TablePrefix   string   `json:"table_prefix" yaml:"table_prefix"`
	IncludeTables []string `json:"include_tables" yaml:"include_tables"`
	ExcludeTables []string `json:"exclude_tables" yaml:"exclude_tables"`
}

type ConnConfig struct {
	DriverName string                          `json:"driver_name" yaml:"driver_name"`
	ReadOnly   bool                            `json:"read_only" yaml:"read_only"`
	Params     dialect.ConnParams              `json:"params" yaml:"params"`
	PartConfig `json:",inline" yaml:",inline"` // 注意逗号不能少
}

func NewReverseSource(c ConnConfig) (*ReverseSource, dialect.Dialect) {
	d := dialect.GetDialectByName(c.DriverName)
	r := &ReverseSource{
		Database: d.Name(), // 其实也等于 c.DriverName
		ConnStr:  d.ParseDSN(c.Params),
	}
	if dr, ok := d.(*dialect.Redis); ok {
		r.OptStr = dr.Values.Encode()
	}
	return r, d
}

func ReverseSource2RedisDialect(r *ReverseSource) *dialect.Redis {
	d, err := dialect.NewRedis(r.ConnStr, "")
	if err != nil || r.Database != d.Name() {
		return nil
	}
	_ = d.ParseOptions(r.OptStr)
	return d
}

type DataSource struct {
	ConnKey      string
	ImporterPath string
	PartConfig
	*ReverseSource
}

func NewDataSource(c ConnConfig, name string) *DataSource {
	ds := &DataSource{ConnKey: name, PartConfig: c.PartConfig}
	var d dialect.Dialect
	ds.ReverseSource, d = NewReverseSource(c)
	if d != nil {
		ds.ImporterPath = d.ImporterPath()
	}
	return ds
}

func (ds DataSource) GetDriverName() string {
	if ds.ReverseSource != nil {
		return ds.ReverseSource.Database
	}
	return ""
}

type Configure struct {
	Application    AppConfig             `json:"application" yaml:"application"`
	Logging        LogConfig             `json:"logging" yaml:"logging"`
	Connections    map[string]ConnConfig `json:"connections" yaml:"connections"`
	ReverseTargets []ReverseTarget       `json:"reverse_targets" yaml:"reverse_targets"`
}

func ReadSettings(fileName string) (*Configure, error) {
	cfg := new(Configure)
	err := ReadSettingsFrom(fileName, &cfg)
	return cfg, err
}

func ReadSettingsFrom(fileName string, cfg interface{}) error {
	rd, err := os.Open(fileName)
	if err == nil {
		err = yaml.NewDecoder(rd).Decode(cfg)
	}
	return err
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

func (cfg Configure) GetReverseTargets() []ReverseTarget {
	return cfg.ReverseTargets
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
