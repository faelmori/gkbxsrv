package services

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	glb "github.com/faelmori/gkbxsrv/internal/globals"
	"github.com/faelmori/gkbxsrv/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var crt ICertService
var Fs FileSystemService

type ConfigService interface {
	GetConfigPath() string
	GetSettings() (map[string]interface{}, error)
	GetSetting(key string) (interface{}, error)
	GetLogger() *log.Logger

	GetDatabaseConfig() glb.Database
	SetDatabaseConfig(dbConfig *glb.Database) error

	SetLogger()

	IsConfigWatchEnabled() bool
	IsConfigLoaded() bool

	WatchConfig(enable bool, event func(fsnotify.Event)) error
	SaveConfig() error
	ResetConfig() error
	LoadConfig() error
	SetupConfig() error

	calculateMD5Hash(filePath string) (string, error)
	getExistingMD5Hash() (string, error)
	saveMD5Hash() error
	compareMD5Hash() (bool, error)
	genCacheFlag(flagToMark string) error
	SetupConfigFromDbService() error
}

type ConfigServiceImpl struct {
	Logger   *log.Logger  `json:"-"`
	FilePath string       `json:"file_path"`
	KeyPath  string       `json:"key_path"`
	CertPath string       `json:"cert_path"`
	Server   glb.Server   `json:"server"`
	Database glb.Database `json:"database"`
	JWT      glb.JWT      `json:"jwt"`
	Redis    glb.Redis    `json:"redis"`
	RabbitMQ glb.RabbitMQ `json:"rabbitmq"`
	MongoDB  glb.MongoDB  `json:"mongodb"`
}

func (c *ConfigServiceImpl) calculateMD5Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	hasher := md5.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	hashBytes := hasher.Sum(nil)
	hashString := fmt.Sprintf("%x", hashBytes)

	return hashString, nil
}

func (c *ConfigServiceImpl) getExistingMD5Hash() (string, error) {
	cfgFilePath := Fs.GetConfigFilePath()
	hashFilePath := fmt.Sprintf("%s.md5", cfgFilePath)
	if _, err := os.Stat(hashFilePath); err == nil {
		hashFile, hashFileErr := os.ReadFile(hashFilePath)
		if hashFileErr != nil {
			return "", hashFileErr
		}
		return string(hashFile), nil
	}
	return "", nil
}

func (c *ConfigServiceImpl) saveMD5Hash() error {
	filePath := Fs.GetConfigFilePath()
	hash, hashErr := c.calculateMD5Hash(filePath)
	if hashErr != nil {
		return hashErr
	}
	hashFilePath := fmt.Sprintf("%s.md5", filePath)
	if _, err := os.Stat(hashFilePath); err == nil {
		_ = os.Remove(hashFilePath)
	}
	hashFile, hashFileErr := os.Create(hashFilePath)
	if hashFileErr != nil {
		return hashFileErr
	}
	_, writeErr := hashFile.WriteString(hash)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func (c *ConfigServiceImpl) compareMD5Hash() (bool, error) {
	existingHash, existingHashErr := c.getExistingMD5Hash()
	if existingHashErr != nil {
		return false, existingHashErr
	}
	newHash, newHashErr := c.calculateMD5Hash(c.FilePath)
	if newHashErr != nil {
		return false, newHashErr
	}
	if existingHash == newHash {
		return true, nil
	}
	return false, nil
}

func (c *ConfigServiceImpl) genCacheFlag(flagToMark string) error {
	if Fs == nil {
		Fs = NewFileSystemSrv(c.FilePath)
	}
	return Fs.SetSetupCacheFlag(flagToMark)
}

func (c *ConfigServiceImpl) SetLogger() {
	c.Logger = log.New(os.Stdout, "GoSpyder", 3)
}

func (c *ConfigServiceImpl) GetLogger() *log.Logger {
	return c.Logger
}

func (c *ConfigServiceImpl) GetConfigPath() string {
	return c.FilePath
}

func (c *ConfigServiceImpl) GetSettings() (map[string]interface{}, error) {
	return viper.AllSettings(), nil
}

func (c *ConfigServiceImpl) GetSetting(key string) (interface{}, error) {
	val := viper.Get(key)
	if val == nil {
		return nil, fmt.Errorf("setting not found")
	}
	return val, nil
}

func (c *ConfigServiceImpl) GetDatabaseConfig() glb.Database {
	var dbConfig glb.Database
	if c == nil {
		err := c.LoadConfig()
		if err != nil {
			return dbConfig
		}
	}
	dbConfig = c.Database
	return dbConfig
}

func (c *ConfigServiceImpl) SetDatabaseConfig(dbConfig *glb.Database) error {
	viper.Set("database", dbConfig)
	c.Database = *dbConfig
	go func(c *ConfigServiceImpl) {
		_ = c.SaveConfig()
	}(c)
	return nil
}

func (c *ConfigServiceImpl) IsConfigWatchEnabled() bool {
	return viper.GetBool("config.watch")
}

func (c *ConfigServiceImpl) IsConfigLoaded() bool {
	return viper.ConfigFileUsed() != ""
}

func (c *ConfigServiceImpl) WatchConfig(enable bool, event func(fsnotify.Event)) error {
	if enable {
		viper.WatchConfig()
		viper.OnConfigChange(event)
	} else {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {})
	}
	return nil
}

func (c *ConfigServiceImpl) SaveConfig() error {
	if Fs == nil {
		Fs = NewFileSystemSrv(c.FilePath)
	}

	if c.FilePath == "" {
		c.FilePath = Fs.GetConfigFilePath()
	}

	data, marshalIndentErr := json.MarshalIndent(c, "", "  ")
	if marshalIndentErr != nil {
		return fmt.Errorf("failed to marshal config: %v", marshalIndentErr)
	}
	if mkdirAllErr := utils.EnsureDir(filepath.Dir(c.FilePath), 0755, []string{}); mkdirAllErr != nil {
		return fmt.Errorf("failed to create directories: %v", mkdirAllErr)
	}

	configFile, configFileErr := os.Create(c.FilePath)
	if configFileErr != nil {
		return configFileErr
	}
	if _, writeFileErr := configFile.Write(data); writeFileErr != nil {
		return fmt.Errorf("failed to write config file: %v", writeFileErr)
	}
	chownErr := configFile.Chown(os.Getuid(), os.Getgid())
	if chownErr != nil {
		return chownErr
	}
	chModReadErr := configFile.Chmod(0600)
	if chModReadErr != nil {
		return chModReadErr
	}

	return nil
}

func (c *ConfigServiceImpl) ResetConfig() error {
	viper.Reset()
	setupErr := c.SetupConfig()
	if setupErr != nil {
		return fmt.Errorf("failed to setup config: %v", setupErr)
	}
	return nil
}

func (c *ConfigServiceImpl) LoadConfig() error {
	if c.Logger == nil {
		c.SetLogger()
	}
	viper.SetConfigFile(c.FilePath)
	viper.AutomaticEnv()
	if readConfigErr := viper.ReadInConfig(); readConfigErr != nil {
		return fmt.Errorf("❌ Erro ao carregar configuração: %v", readConfigErr)
	}
	viper.WatchConfig()
	return nil
}

func (c *ConfigServiceImpl) SetupConfig() error {
	dckr := NewDockerSrv()

	if blKeepCfg, blKeepCfgErr := c.compareMD5Hash(); blKeepCfg && blKeepCfgErr == nil {
		if loadConfigErr := c.LoadConfig(); loadConfigErr != nil {
			return loadConfigErr
		}
		if dbErr := dckr.SetupDatabaseServices(); dbErr != nil {
			return dbErr
		}
		return nil
	}

	if c.Logger == nil {
		c.SetLogger()
	}
	if createDirErr := Fs.CreateKubexUserStructure(); createDirErr != nil {
		return createDirErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_config_structure")

	prvKey, pubKey, crtErr := crt.GenSelfCert()
	if crtErr != nil {
		return crtErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_certificates")
	redisPass, redisPassErr := crt.GenerateRandomKey(10)
	if redisPassErr != nil {
		return redisPassErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_redis_password")
	refreshSecret, refreshSecretErr := crt.GenerateRandomKey(10)
	if refreshSecretErr != nil {
		return refreshSecretErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_refresh_secret")
	password, passwordErr := crt.GenerateRandomKey(10)
	if passwordErr != nil {
		return passwordErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_db_password")

	if c.FilePath == "" {
		Fs = NewFileSystemSrv(c.FilePath)
		c.FilePath = Fs.GetConfigFilePath()
	}

	newC := &ConfigServiceImpl{
		Server: glb.Server{
			Port:         "8000",
			BindAddress:  "0.0.0.0",
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Database: glb.Database{
			Type:             "postgresql",
			Driver:           "postgres",
			ConnectionString: fmt.Sprintf("postgres://kubex_adm:%s@localhost:5432/kubex_db", password),
			Dsn:              fmt.Sprintf("postgres://kubex_adm:%s@localhost:5432/kubex_db", password),
			Path:             os.ExpandEnv(`$HOME/.kubex/volumes/postgresql`),
			Host:             "localhost",
			Port:             "5432",
			Username:         "kubex_adm",
			Password:         password,
			Name:             "kubex_db",
		},
		JWT: glb.JWT{
			RefreshSecret:         refreshSecret,
			PrivateKey:            base64.StdEncoding.EncodeToString(prvKey),
			PublicKey:             base64.StdEncoding.EncodeToString(pubKey),
			ExpiresIn:             3600,
			IDExpirationSecs:      3600,
			RefreshExpirationSecs: 86400,
		},
		Redis: glb.Redis{
			Enabled:  true,
			Port:     6379,
			Addr:     "localhost:6379",
			Username: "kubex_adm",
			Password: redisPass,
			DB:       0,
		},
		RabbitMQ: glb.RabbitMQ{
			Enabled:        true,
			Username:       "guest",
			Password:       "guest",
			Port:           5672,
			ManagementPort: 15672,
		},
	}
	c.Server = newC.Server
	c.Database = newC.Database
	c.JWT = newC.JWT
	c.Redis = newC.Redis
	c.RabbitMQ = newC.RabbitMQ

	saveCfgErr := c.SaveConfig()
	if saveCfgErr != nil {
		return saveCfgErr
	}
	_ = c.genCacheFlag("kubex_config_structure")

	saveCfgHashErr := c.saveMD5Hash()
	if saveCfgHashErr != nil {
		return saveCfgHashErr
	}

	loadConfigErr := c.LoadConfig()
	if loadConfigErr != nil {
		return loadConfigErr
	}

	if dbErr := dckr.SetupDatabaseServices(); dbErr != nil {
		return dbErr
	}

	return nil
}

func (c *ConfigServiceImpl) SetConfigProperty(key string, value interface{}) error {
	viper.Set(key, value)
	return nil
}

func (c *ConfigServiceImpl) SetupConfigFromDbService() error {
	if blKeepCfg, blKeepCfgErr := c.compareMD5Hash(); blKeepCfg && blKeepCfgErr == nil {
		if loadConfigErr := c.LoadConfig(); loadConfigErr != nil {
			return loadConfigErr
		}
		if dbErr := NewDockerSrv().SetupDatabaseServices(); dbErr != nil {
			return dbErr
		}
		return nil
	}

	if c.Logger == nil {
		c.SetLogger()
	}
	if createDirErr := Fs.CreateKubexUserStructure(); createDirErr != nil {
		return createDirErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_config_structure")

	prvKey, pubKey, crtErr := crt.GenSelfCert()
	if crtErr != nil {
		return crtErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_certificates")
	redisPass, redisPassErr := crt.GenerateRandomKey(10)
	if redisPassErr != nil {
		return redisPassErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_redis_password")
	refreshSecret, refreshSecretErr := crt.GenerateRandomKey(10)
	if refreshSecretErr != nil {
		return refreshSecretErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_refresh_secret")
	password, passwordErr := crt.GenerateRandomKey(10)
	if passwordErr != nil {
		return passwordErr
	}
	_ = Fs.SetSetupCacheFlag("kubex_db_password")

	if c.FilePath == "" {
		Fs = NewFileSystemSrv(c.FilePath)
		c.FilePath = Fs.GetConfigFilePath()
	}

	newC := &ConfigServiceImpl{
		Server: glb.Server{
			Port:         "8000",
			BindAddress:  "0.0.0.0",
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Database: glb.Database{
			Type:             "postgresql",
			Driver:           "postgres",
			ConnectionString: fmt.Sprintf("postgres://kubex_adm:%s@127.0.0.1:5432/kubex_db", password),
			Dsn:              fmt.Sprintf("postgres://kubex_adm:%s@127.0.0.1:5432/kubex_db", password),
			Path:             os.ExpandEnv(`$HOME/.kubex/volumes/postgresql`),
			Host:             "127.0.0.1",
			Port:             "5432",
			Username:         "kubex_adm",
			Password:         password,
			Name:             "kubex_db",
		},
		JWT: glb.JWT{
			RefreshSecret:         refreshSecret,
			PrivateKey:            base64.StdEncoding.EncodeToString(prvKey),
			PublicKey:             base64.StdEncoding.EncodeToString(pubKey),
			ExpiresIn:             3600,
			IDExpirationSecs:      3600,
			RefreshExpirationSecs: 86400,
		},
		Redis: glb.Redis{
			Enabled:  true,
			Port:     6379,
			Addr:     "127.0.0.1:6379",
			Username: "kubex_adm",
			Password: redisPass,
			DB:       0,
		},
		RabbitMQ: glb.RabbitMQ{
			Enabled:        true,
			Username:       "guest",
			Password:       "guest",
			Port:           5672,
			ManagementPort: 15672,
		},
	}
	c.Server = newC.Server
	c.Database = newC.Database
	c.JWT = newC.JWT
	c.Redis = newC.Redis
	c.RabbitMQ = newC.RabbitMQ

	saveCfgErr := c.SaveConfig()
	if saveCfgErr != nil {
		return saveCfgErr
	}
	_ = c.genCacheFlag("kubex_config_structure")

	saveCfgHashErr := c.saveMD5Hash()
	if saveCfgHashErr != nil {
		return saveCfgHashErr
	}

	loadConfigErr := c.LoadConfig()
	if loadConfigErr != nil {
		return loadConfigErr
	}
	return nil
}

func NewConfigSrv(configPath, keyPath, certPath string) ConfigService {
	home, homeErr := utils.GetWorkDir()
	if homeErr != nil {
		fmt.Println("❌ Erro ao obter diretório de trabalho: ", homeErr)
		os.Exit(1)
	}
	home = filepath.Dir(home)

	if configPath == "" {
		configPath = strings.ReplaceAll(glb.DefaultGoSpyderConfigPath, "$HOME", home)
	} else {
		configPath = strings.ReplaceAll(configPath, "$HOME", home)
	}
	if keyPath == "" {
		keyPath = strings.ReplaceAll(glb.DefaultKeyPath, "$HOME", home)
	} else {
		keyPath = strings.ReplaceAll(keyPath, "$HOME", home)
	}
	if certPath == "" {
		certPath = strings.ReplaceAll(glb.DefaultCertPath, "$HOME", home)
	} else {
		certPath = strings.ReplaceAll(certPath, "$HOME", home)
	}

	if Fs == nil {
		Fs = NewFileSystemSrv(configPath)
	}
	if crt == nil {
		crt = NewCertService(keyPath, certPath)
	}

	return &ConfigServiceImpl{
		FilePath: configPath,
		KeyPath:  keyPath,
		CertPath: certPath,
	}
}
