package gkbxsrv

import (
	log "github.com/faelmori/gkbxsrv/logz"
	kbxsrv "github.com/faelmori/gkbxsrv/services"
	vs "github.com/faelmori/gkbxsrv/version"
)

var (
	fsSvc   kbxsrv.FilesystemService
	cnfgSvc kbxsrv.ConfigService
	vSvc    vs.VersionService
	dbSvc   kbxsrv.DatabaseService
	crtSvc  kbxsrv.CertService
	brkrSvc *kbxsrv.Broker
	dkSvc   kbxsrv.DockerSrv
)

func initializeServicesDefault() {
	vSvc = vs.NewVersionService()
	cv := vSvc.GetCurrentVersion()
	log.Logger.Info("Starting gkbxsrv...", map[string]interface{}{
		"context": "main",
		"action":  "start",
		"version": cv,
	})

	//Broker service
	_, brkrErr := kbxsrv.NewBrokerService(true)
	if brkrErr != nil {
		log.Logger.Error("Error creating broker service", map[string]interface{}{
			"context": "main",
			"action":  "newBrokerService",
			"error":   brkrErr,
			"version": cv,
		})
		return
	}

	//Filesystem service
	fs := kbxsrv.NewFileSystemService("gkbxsrv")
	fsSvc = *fs

	//Config service
	cnfgSvc = kbxsrv.NewConfigService(fsSvc.GetConfigFilePath(), fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	loadCfgErr := cnfgSvc.LoadConfig()
	if loadCfgErr != nil {
		log.Logger.Error("Error loading configuration", map[string]interface{}{
			"context": "main",
			"action":  "loadConfig",
			"error":   loadCfgErr,
			"version": cv,
		})

		return
	}

	//Docker service
	dkSvc = kbxsrv.NewDockerService()
	if dkSvcInsChk := dkSvc.IsDockerInstalled(); !dkSvcInsChk {
		log.Logger.Warn("Docker is not installed", map[string]interface{}{
			"context": "main",
			"action":  "isDockerInstalled",
			"version": cv,
		})
		insDockerErr := dkSvc.InstallDocker()
		if insDockerErr != nil {
			log.Logger.Error("Error installing Docker", map[string]interface{}{
				"context": "main",
				"action":  "installDocker",
				"error":   insDockerErr,
				"version": cv,
			})
			return
		}
	}

	//Database service
	dbSvc = kbxsrv.NewDatabaseService(cnfgSvc.GetConfigPath())

	//Certificate service
	cs := kbxsrv.NewCertService(fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	crtSvc = *cs
}

func GetFilesystemService(configFile string) kbxsrv.FilesystemService {
	if fsSvc == nil {
		fs := kbxsrv.NewFileSystemService(configFile)
		fsSvc = *fs
	}
	return fsSvc
}

func GetConfigService(configFile string) kbxsrv.ConfigService {
	if cnfgSvc == nil {
		fs := GetFilesystemService(configFile)
		cnfgSvc = kbxsrv.NewConfigService(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
	}
	cnfgSvcErr := cnfgSvc.LoadConfig()
	if cnfgSvcErr != nil {
		log.Logger.Error("Error loading configuration", map[string]interface{}{
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

func GetDatabaseService(configFile string) kbxsrv.DatabaseService {
	if dbSvc == nil {
		cnfgSvc := GetConfigService(configFile)
		dbSvc = kbxsrv.NewDatabaseService(cnfgSvc.GetConfigPath())
	}
	return dbSvc
}

func GetCertService() kbxsrv.CertService {
	if crtSvc == nil {
		fs := GetFilesystemService("gkbxsrv")
		cs := kbxsrv.NewCertService(fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
		crtSvc = *cs
	}
	return crtSvc
}

func GetBrokerService() *kbxsrv.Broker {
	if brkrSvc == nil {
		_, brkrErr := kbxsrv.NewBrokerService(true)
		if brkrErr != nil {
			log.Logger.Error("Error creating broker service", map[string]interface{}{
				"context": "main",
				"action":  "newBrokerService",
				"error":   brkrErr,
			})
		}
	}
	return brkrSvc
}

func GetDockerService() kbxsrv.DockerSrv {
	if dkSvc == nil {
		dkSvc = kbxsrv.NewDockerService()
	}
	return dkSvc
}

func GetServices(configFile string) map[string]interface{} {
	return map[string]interface{}{
		"fileSystem":  GetFilesystemService(configFile),
		"config":      GetConfigService(configFile),
		"version":     GetVersionService(),
		"database":    GetDatabaseService(configFile),
		"certificate": GetCertService(),
		"broker":      GetBrokerService(),
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
