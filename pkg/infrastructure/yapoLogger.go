package infrastructure

import (
	"github.com/Yapo/logger"
	"github.schibsted.io/Yapo/suggested-ads/pkg/interfaces/loggers"
)

// yapoLogger struct that implements the Logger interface using the Yapo/logger library
type yapoLogger struct{}

// MakeYapoLogger creates and sets up a yapo flavored Logger
func MakeYapoLogger(config *LoggerConf) (loggers.Logger, error) {
	var log yapoLogger
	err := log.init(config)
	return log, err
}

// Init intialize the logger
func (y *yapoLogger) init(config *LoggerConf) error {
	loggerConf := logger.LogConfig{
		Syslog: logger.SyslogConfig{
			Enabled:  config.SyslogEnabled,
			Identity: config.SyslogIdentity,
		},
		Stdlog: logger.StdlogConfig{
			Enabled: config.StdlogEnabled,
		},
	}
	if err := logger.Init(loggerConf); err != nil {
		return err
	}
	logger.SetLogLevel(config.LogLevel)
	return nil
}

// LogDebug logs a message at DEBUG level
func (y yapoLogger) Debug(format string, params ...interface{}) {
	logger.Debug(format, params...)
}

// LogInfo logs a message at INFO level
func (y yapoLogger) Info(format string, params ...interface{}) {
	logger.Info(format, params...)
}

// LogWarn logs a message at WARNING level
func (y yapoLogger) Warn(format string, params ...interface{}) {
	logger.Warn(format, params...)
}

// LogError logs a message at ERROR level
func (y yapoLogger) Error(format string, params ...interface{}) {
	logger.Error(format, params...)
}

// LogCrit logs a message at CRITICAL level
func (y yapoLogger) Crit(format string, params ...interface{}) {
	logger.Crit(format, params...)
}
