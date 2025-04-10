package gkbxsrv

import (
	kbxsrv "github.com/faelmori/gkbxsrv/internal/services"
	"github.com/faelmori/gkbxsrv/services"
	vs "github.com/faelmori/gkbxsrv/version"
	"github.com/faelmori/kbxutils/factory"
	kbxApi "github.com/faelmori/kbxutils/utils/interfaces"

	log "github.com/faelmori/logz"
	"os"
)

var (
	fsSvc   kbxApi.FileSystemService
	cnfgSvc kbxApi.IConfigService
	vSvc    vs.VersionService
	dbSvc   kbxApi.IDatabaseService
	crtSvc  kbxApi.ICertService
	brkrSvc *kbxsrv.BrokerImpl
	dkSvc   kbxApi.IDockerSrv
)

func initializeServicesDefault() {
	vSvc = vs.NewVersionService()
	cv := vSvc.GetCurrentVersion()
	log.InfoCtx("Starting gkbxsrv...", map[string]interface{}{
		"context": "main",
		"action":  "start",
		"version": cv,
	})

	//Broker service
	port := os.Getenv("GKBXSRV_BROKER_PORT")
	if port == "" {
		port = "5555"
	}
	_, brkrErr := services.NewBrokerService(true, port)
	if brkrErr != nil {
		log.ErrorCtx("Error creating broker service", map[string]interface{}{
			"context": "main",
			"action":  "newBrokerService",
			"error":   brkrErr,
			"version": cv,
		})
		return
	}

	//Filesystem service
	fs := factory.NewFilesystemService("gkbxsrv")
	fsSvc = *fs

	//Config service
	cnfgSvc = factory.NewConfigService(fsSvc.GetConfigFilePath(), fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	loadCfgErr := cnfgSvc.LoadConfig()
	if loadCfgErr != nil {
		log.ErrorCtx("Error loading configuration", map[string]interface{}{
			"context": "main",
			"action":  "loadConfig",
			"error":   loadCfgErr,
			"version": cv,
		})

		return
	}

	//Docker service
	dkSvc = factory.NewDockerService()
	if dkSvcInsChk := dkSvc.IsDockerInstalled(); !dkSvcInsChk {
		log.WarnCtx("Docker is not installed", map[string]interface{}{
			"context": "main",
			"action":  "isDockerInstalled",
			"version": cv,
		})
		insDockerErr := dkSvc.InstallDocker()
		if insDockerErr != nil {
			log.ErrorCtx("Error installing Docker", map[string]interface{}{
				"context": "main",
				"action":  "installDocker",
				"error":   insDockerErr,
				"version": cv,
			})
			return
		}
	}

	//Database service
	dbSvc = factory.NewDatabaseService(cnfgSvc.GetConfigPath())

	//Certificate service
	cs := factory.NewCertService(fsSvc.GetDefaultKeyPath(), fsSvc.GetDefaultCertPath())
	crtSvc = cs
}

func GetFilesystemService(configFile string) kbxApi.FileSystemService {
	if fsSvc == nil {
		fs := factory.NewFilesystemService("")
		fsSvc = *fs
	}
	return fsSvc
}

func GetConfigService(configFile string) kbxApi.IConfigService {
	if cnfgSvc == nil {
		fs := GetFilesystemService(configFile)
		cnfgSvc = factory.NewConfigService(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
	}
	cnfgSvcErr := cnfgSvc.LoadConfig()
	if cnfgSvcErr != nil {
		log.ErrorCtx("Error loading configuration", map[string]interface{}{
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

func GetDatabaseService(configFile string) kbxApi.IDatabaseService {
	if dbSvc == nil {
		cnfgSvc := GetConfigService(configFile)
		dbSvc = factory.NewDatabaseService(cnfgSvc.GetConfigPath())
	}
	return dbSvc
}

func GetCertService() kbxApi.ICertService {
	if crtSvc == nil {
		fs := GetFilesystemService("gkbxsrv")
		cs := factory.NewCertService(fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
		crtSvc = cs
	}
	return crtSvc
}

func NewBrokerService(port string) *services.Broker {
	if brkrSvc == nil {
		_, brkrErr := services.NewBrokerService(true, port)
		if brkrErr != nil {
			log.ErrorCtx("Error creating broker service", map[string]interface{}{
				"context": "main",
				"action":  "newBrokerService",
				"error":   brkrErr,
			})
		}
	}
	return brkrSvc
}

func GetDockerService() kbxApi.IDockerSrv {
	if dkSvc == nil {
		dkSvc = factory.NewDockerService()
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
