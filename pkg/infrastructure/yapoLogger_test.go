package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYapoLoggerNotStarted(t *testing.T) {
	conf := LoggerConf{
		SyslogIdentity: "test",
		SyslogEnabled:  false,
		StdlogEnabled:  false,
		LogLevel:       0,
	}
	_, err := MakeYapoLogger(&conf)
	assert.Error(t, err)
}

func TestYapoLogger(t *testing.T) {
	conf := LoggerConf{
		SyslogIdentity: "test",
		SyslogEnabled:  false,
		StdlogEnabled:  true,
		LogLevel:       0,
	}
	logger, err := MakeYapoLogger(&conf)
	assert.NoError(t, err)
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warning")
	logger.Error("error")
	logger.Crit("critical")
}
