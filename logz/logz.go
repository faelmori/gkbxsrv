package logz

import "github.com/faelmori/logz"

var (
	logzCfg logz.Config
	logger  logz.Logger
)

func getLogger() logz.Logger {
	if logger == nil {
		logger = logz.NewLogger("GoSpyder")
		logzCfg = logger.GetConfig()
		logzCfg.SetLevel("INFO")
		logger.SetConfig(logzCfg)
	}
	return logger
}

var Logger = getLogger()
