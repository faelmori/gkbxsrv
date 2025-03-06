package logz

import (
	"github.com/faelmori/logz"
	"os"
	"path/filepath"
)

var (
	logzCfg logz.Config
	logger  logz.Logger
)

func getLogger() logz.Logger {
	if logger == nil {
		logger = logz.NewLogger("GoSpyder")
		logzCfg = logger.GetConfig()
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "INFO"
		}
		logzCfg.SetLevel(logz.LogLevel(logLevel))
		lf := os.Getenv("LOG_FORMAT")
		if lf == "" {
			lf = "JSON"
		}
		logzCfg.SetFormat(logz.LogFormat(lf))
		lo := os.Getenv("LOG_OUTPUT")
		if lo == "" {
			lo = filepath.Join(os.Getenv("HOME"), ".kubex/logz/gkbxsrv.log")
		}
		logzCfg.SetOutput(lo)
		logger.SetConfig(logzCfg)
	}
	return logger
}

var Logger = getLogger()
