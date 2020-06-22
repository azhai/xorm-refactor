package config

import (
	"github.com/azhai/gozzo-utils/logging"
	"xorm.io/xorm/log"
)

// 将 xorm 的日志级别转为字符串
func GetLevelString(lvl log.LogLevel) string {
	var level string
	switch lvl {
	case log.LOG_DEBUG:
		level = "debug"
	case log.LOG_INFO:
		level = "info"
	case log.LOG_WARNING:
		level = "warn"
	case log.LOG_ERR:
		level = "error"
	default:
		level = "fatal"
	}
	return level
}

type SqlLogger struct {
	level   log.LogLevel
	showSQL bool
	*logging.Logger
}

func NewSqlLogger(filename string) *SqlLogger {
	cfg := logging.DefaultConfig
	if filename != "" {
		cfg.OutputMap = map[string][]string{
			":": {filename},
		}
	}
	logger := &SqlLogger{Logger: cfg.BuildSugar()}
	logger.SetLevel(log.LOG_INFO)
	logger.ShowSQL()
	return logger
}

// Level implement ILogger
func (s *SqlLogger) Level() log.LogLevel {
	return s.level
}

// SetLevel implement ILogger
func (s *SqlLogger) SetLevel(l log.LogLevel) {
	s.level = l
	return
}

// ShowSQL implement ILogger
func (s *SqlLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (s *SqlLogger) IsShowSQL() bool {
	return s.showSQL
}
