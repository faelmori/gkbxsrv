package gkbxsrv

import (
	vs "github.com/faelmori/gkbxsrv/version"
	kbxApi "github.com/faelmori/kbxutils/api"
	kbxsrv "github.com/faelmori/kbxutils/utils/helpers"
	log "github.com/faelmori/logz"
	"os"
)

var (
	fsSvc   kbxsrv.FileSystemService
	cnfgSvc kbxsrv.ConfigService
	vSvc    vs.VersionService
	dbSvc   kbxsrv.IDatabaseService
	crtSvc  kbxsrv.ICertService
	//brkrSvc *kbxsrv.Broker
	dkSvc kbxsrv.DockerService
)

func initializeServicesDefault() {
	vSvc = vs.NewVersionService()
	cv := vSvc.GetCurrentVersion()
	log.Info("Starting gkbxsrv...", map[string]interface{}{
		"context": "main",
		"action":  "start",
		"version": cv,
	})

	//Broker service
	port := os.Getenv("GKBXSRV_BROKER_PORT")
	if port == "" {
		port = "5555"
	}
	_, brkrErr := kbxApi.NewBrokerService(true, port)
	if brkrErr != nil {
		log.Error("Error creating broker service", map[string]interface{}{
			"context": "main",
			"action":  "newBrokerService",
			"error":   brkrErr,
			"version": cv,
		})
		return
	}

	//Filesystem service
	fs := kbxApi.NewFileSystemService("gkbxsrv")
	fsSvc = *fs

	//Config service
	cnfgSvc = kbxApi.NewConfigService(fsSvc.GetConfigFilePath(), fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	loadCfgErr := cnfgSvc.LoadConfig()
	if loadCfgErr != nil {
		log.Error("Error loading configuration", map[string]interface{}{
			"context": "main",
			"action":  "loadConfig",
			"error":   loadCfgErr,
			"version": cv,
		})

		return
	}

	//Docker service
	dkSvc = kbxApi.NewDockerService()
	if dkSvcInsChk := dkSvc.IsDockerInstalled(); !dkSvcInsChk {
		log.Warn("Docker is not installed", map[string]interface{}{
			"context": "main",
			"action":  "isDockerInstalled",
			"version": cv,
		})
		insDockerErr := dkSvc.InstallDocker()
		if insDockerErr != nil {
			log.Error("Error installing Docker", map[string]interface{}{
				"context": "main",
				"action":  "installDocker",
				"error":   insDockerErr,
				"version": cv,
			})
			return
		}
	}

	//Database service
	dbSvc = kbxApi.NewDatabaseService(cnfgSvc.GetConfigPath())

	//Certificate service
	cs := kbxApi.NewCertService(fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	crtSvc = cs
}

func GetFilesystemService(configFile string) kbxsrv.FileSystemService {
	if fsSvc == nil {
		fs := kbxApi.NewFileSystemService(configFile)
		fsSvc = *fs
	}
	return fsSvc
}

func GetConfigService(configFile string) kbxsrv.ConfigService {
	if cnfgSvc == nil {
		fs := GetFilesystemService(configFile)
		cnfgSvc = kbxApi.NewConfigService(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
	}
	cnfgSvcErr := cnfgSvc.LoadConfig()
	if cnfgSvcErr != nil {
		log.Error("Error loading configuration", map[string]interface{}{
			"context": "main",
			"action":  "loadConfig",
			"error":   cnfgSvcErr,
		})
	}
	return cnfgSvc
}

func GetVersionService() vs.VersionService {
	if vSvc == nil {
		vSvc = vs.NewVersionService()
	}
	return vSvc
}

func GetDatabaseService(configFile string) kbxApi.DatabaseService {
	if dbSvc == nil {
		cnfgSvc := GetConfigService(configFile)
		dbSvc = kbxApi.NewDatabaseService(cnfgSvc.GetConfigPath())
	}
	return dbSvc
}

func GetCertService() kbxsrv.ICertService {
	if crtSvc == nil {
		fs := GetFilesystemService("gkbxsrv")
		cs := kbxApi.NewCertService(fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
		crtSvc = cs
	}
	return crtSvc
}

func NewBrokerService(port string) *kbxsrv.Broker {
	if brkrSvc == nil {
		_, brkrErr := kbxsrv.NewBrokerService(true, port)
		if brkrErr != nil {
			log.Error("Error creating broker service", map[string]interface{}{
				"context": "main",
				"action":  "newBrokerService",
				"error":   brkrErr,
			})
		}
	}
	return brkrSvc
}

func GetDockerService() kbxApi.DockerSrv {
	if dkSvc == nil {
		dkSvc = kbxApi.NewDockerService()
	}
	return dkSvc
}

func GetServices(configFile string) map[string]interface{} {
	port := os.Getenv("GKBXSRV_BROKER_PORT")
	if port == "" {
		port = "5555"
	}
	return map[string]interface{}{
		"fileSystem":  GetFilesystemService(configFile),
		"config":      GetConfigService(configFile),
		"version":     GetVersionService(),
		"database":    GetDatabaseService(configFile),
		"certificate": GetCertService(),
		"broker":      NewBrokerService(port),
		"docker":      GetDockerService(),
	}
}

func GetServicesDefault() {
	initializeServicesDefault()
}

func Version() string {
	v := GetVersionService()
	return v.GetCurrentVersion()
}

func VersionCheck() string {
	v := GetVersionService()
	status := ""
	if ok, okErr := v.IsLatestVersion(); !ok || okErr != nil {
		if okErr != nil {
			status = "Error checking latest version"
		} else {
			lt, ltErr := v.GetLatestVersion()
			if ltErr != nil {
				status = "Error checking latest version"
			} else {
				status = "Update available: " + lt
			}
		}
	} else {
		status = "Up to date (" + v.GetCurrentVersion() + ")"
	}
	return status
}

func GetConfigPath() string {
	if cnfgSvc == nil {
		return ""
	}
	return cnfgSvc.GetConfigPath()
}
