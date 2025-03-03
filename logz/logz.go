package logz

import "github.com/faelmori/logz"

var (
	logzCfg logz.LogzConfig
	logger  logz.LogzLogger
)

func getLogger() logz.LogzLogger {
	if logger == nil {
		logger = logz.NewLogger("GoSpyder")
		logzCfg = logger.GetConfig()
		logzCfg.SetLevel("INFO")
		logzCfg.SetFormat("JSON")
		logzCfg.SetOutput("/home/user/.kubex/logz/gkbxsrv.log")
		logger.SetConfig(logzCfg)
	}
	return logger
}

var Logger = getLogger()
