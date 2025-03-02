package logz

import (
	"github.com/faelmori/logz"
)

var (
	logzCfg logz.LogzConfig
	logger  logz.LogzLogger
)

func init() {
	if logger == nil {
		logger = logz.NewLogger("GoSpyder")
		logzCfg = logger.GetConfig()
		logzCfg.SetLevel("INFO")
		logzCfg.SetFormat("json")
		logger.SetConfig(logzCfg)
	}
}

// Logger returns the global logger instance.
func Logger() logz.LogzLogger { return logger }
