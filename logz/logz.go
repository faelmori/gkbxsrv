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

var Logger = getLogger()

func getLogger() logz.Logger {
	if logger == nil {
		logger = logz.NewLogger("GKbxSRV")
		logzCfg = logger.GetConfig()
		logLevel := os.Getenv("LOG_LEVEL")
		if logLevel == "" {
			logLevel = "DEBUG"
		}
		logzCfg.SetLevel(logz.LogLevel(logLevel))
		lf := os.Getenv("LOG_FORMAT")
		if lf == "" {
			lf = "JSON"
		}
		logzCfg.SetFormat(logz.LogFormat(lf))
		lo := os.Getenv("LOG_OUTPUT")
		var hmErr error
		if lo == "" {
			hm := os.Getenv("HOME")
			if hm == "" {
				hm, hmErr = os.UserCacheDir()
				if hmErr != nil {
					hm = "/tmp"
				}
			}
			lo = filepath.Join(hm, ".kubex/logz/gkbxsrv.log")
		}
		logzCfg.SetOutput(lo)
		logger.SetConfig(logzCfg)
	}
	return logger
}
