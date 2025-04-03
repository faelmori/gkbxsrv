package services

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/globals"
	"github.com/faelmori/gkbxsrv/internal/utils"
	"github.com/faelmori/gkbxsrv/version"
	"github.com/goccy/go-json"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	setupFlagsMap = map[string]string{
		globals.KubexConfigStructureFlag: base64.StdEncoding.EncodeToString([]byte(globals.KubexConfigStructureFlag)),
		globals.KubexCertificatesFlag:    base64.StdEncoding.EncodeToString([]byte(globals.KubexCertificatesFlag)),
		globals.KubexRedisPasswordFlag:   base64.StdEncoding.EncodeToString([]byte(globals.KubexRedisPasswordFlag)),
		globals.KubexRefreshSecretFlag:   base64.StdEncoding.EncodeToString([]byte(globals.KubexRefreshSecretFlag)),
		globals.KubexDBPasswordFlag:      base64.StdEncoding.EncodeToString([]byte(globals.KubexDBPasswordFlag)),
		globals.KubexCacheSetupFlag:      base64.StdEncoding.EncodeToString([]byte(globals.KubexCacheSetupFlag)),
		globals.KubexServicesSetupFlag:   base64.StdEncoding.EncodeToString([]byte(globals.KubexServicesSetupFlag)),
		globals.KubexVaultSetupFlag:      base64.StdEncoding.EncodeToString([]byte(globals.KubexVaultSetupFlag)),
		globals.KubexDepsSetupFlag:       base64.StdEncoding.EncodeToString([]byte(globals.KubexDepsSetupFlag)),
	}
)

type FileSystemService interface {
	ExistsConfigFile() bool
	ExistsKubexDirs() (map[string]bool, error)
	ExistsHostKubexDirs() (map[string]bool, error)
	CreateKubexUserStructure() error
	GetConfigFilePath() string
	GetDefaultCacheDir() string
	GetDefaultRedisVolumeDir() string
	GetDefaultMongoVolumeDir() string
	GetDefaultRabbitMQVolumeDir() string
	GetDefaultPostgresVolumeDir() string
	GetDefaultKubexDir() string
	GetDefaultVaultDir() string
	GetDefaultKbxDir() string
	GetDefaultRootDir() string
	GetDefaultConfigDir() string
	GetDefaultKeyPath() string
	GetDefaultCertPath() string
	WriteToFile(filePath string, data interface{}, fileType *string) error
	GetFromFile(filePath string, fileType *string, data interface{}) (object interface{}, err error)
	SetSetupCacheFlag(setupFlag string) error
	CheckSetupCacheFlag(setupFlag string) bool
	SanitizePath(parts ...string) string
	hostDefaultDirs() ([]string, error)
	kubexDefaultDirs() ([]string, error)
}

type FileSystemServiceImpl struct {
	configFilePath string
	cacheDir       string
	redisVolume    string
	mongoVolume    string
	rabbitMQVolume string
	postgresVolume string
	kubexDir       string
	vaultDir       string
	kbxDir         string
	goSpiderDir    string
	rootDir        string
	configDir      string
	keyPath        string
	certPath       string
}

func (f *FileSystemServiceImpl) SanitizePath(parts ...string) string {
	var cleanedParts = make([]string, 0)
	for _, p := range parts {
		cleanedParts = append(cleanedParts, filepath.Clean(strings.TrimSpace(strings.ToValidUTF8(os.ExpandEnv(p), ""))))
	}
	fullPath := strings.Join(cleanedParts, string(filepath.Separator))
	return fullPath
}
func (f *FileSystemServiceImpl) CheckSetupCacheFlag(setupFlag string) bool {
	if realSetupFlag, ok := setupFlagsMap[setupFlag]; ok {
		setupFlag = realSetupFlag
	} else {
		log.Fatal(fmt.Sprintf("❌ This is not a Kubex setup: %s\n", setupFlag))
	}
	cacheFlagPath := filepath.Join(f.cacheDir, strings.Join([]string{setupFlag, ".flag"}, version.Version()))
	if _, statErr := os.Stat(cacheFlagPath); os.IsNotExist(statErr) {
		return false
	}
	return true
}
func (f *FileSystemServiceImpl) SetSetupCacheFlag(setupFlag string) error {
	cacheFlagPath := filepath.Join(f.cacheDir, strings.Join([]string{setupFlag, ".flag"}, version.Version()))
	if _, statErr := os.Stat(cacheFlagPath); os.IsNotExist(statErr) {
		if _, createErr := os.Create(cacheFlagPath); createErr != nil {
			return fmt.Errorf("❌ Erro ao criar flag de cache: %v", createErr)
		}
	}

	return nil
}
func (f *FileSystemServiceImpl) hostDefaultDirs() ([]string, error) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("❌ Erro ao obter o diretório do usuário: %v", homeErr)
	}
	return []string{
		filepath.Join(home, ".cache", "kubex"),
		filepath.Join(home, ".kubex", ".vault"),
		filepath.Join(home, ".kubex", "volumes", "postgresql"),
		filepath.Join(home, ".kubex", "volumes", "mongo"),
		filepath.Join(home, ".kubex", "volumes", "redis"),
		filepath.Join(home, ".kubex", "volumes", "rabbitmq"),
		filepath.Join(home, ".kubex", "gospider"),
		filepath.Join(home, ".kubex", "kbx"),
		filepath.Join(home, ".kubex"),
	}, nil
}
func (f *FileSystemServiceImpl) kubexDefaultDirs() ([]string, error) {
	fmt.Println("🔍 Verificando diretórios padrão...")
	var dirs []string
	dirs = make([]string, 0)
	for index, dir := range dirs {
		dirs[index] = f.SanitizePath(dir)
		//fmt.Println("📁 Diretório:", dirs[index])
	}
	return dirs, nil
}
func (f *FileSystemServiceImpl) ExistsKubexDirs() (map[string]bool, error) {
	dirs, dirsErr := f.kubexDefaultDirs()
	if dirsErr != nil {
		return nil, fmt.Errorf("❌ Erro ao obter diretórios (o): %v - %s", dirsErr, strings.Join(dirs, ", "))
	}
	exists := make(map[string]bool)
	for index, dir := range dirs {
		if _, statErr := os.Stat(dirs[index]); os.IsNotExist(statErr) {
			exists[dir] = false
		} else {
			exists[dir] = true
		}
	}
	return exists, nil
}
func (f *FileSystemServiceImpl) ExistsConfigFile() bool {
	if _, statErr := os.Stat(f.configFilePath); os.IsNotExist(statErr) {
		return false
	}
	return true
}
func (f *FileSystemServiceImpl) ExistsHostKubexDirs() (map[string]bool, error) {
	dirs, dirsErr := f.hostDefaultDirs()
	if dirsErr != nil {
		return nil, dirsErr
	}
	exists := make(map[string]bool)
	for _, dir := range dirs {
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			exists[dir] = false
		} else {
			exists[dir] = true
		}
	}
	return exists, nil
}
func (f *FileSystemServiceImpl) CreateKubexUserStructure() error {
	dirsMap, dirsErr := f.ExistsKubexDirs()
	if dirsErr != nil {
		return fmt.Errorf("❌ Erro ao verificar diretórios (v): %v", dirsErr)
	}
	for dir, exists := range dirsMap {
		if exists {
			fmt.Println("✅ Diretório já existe:", dir)
			continue
		}
		if filepath.Dir(dir) == dir {
			if err := utils.EnsureDir(dir, 0755, []string{}); err != nil {
				log.Fatalf("❌ Erro ao criar diretórios (d): %v - %s", err, dir)
				//os.Exit(1)
			}
		} else {
			if err := utils.EnsureFile(dir, 0755, []string{}); err != nil {
				log.Fatalf("❌ Erro ao criar diretórios (f): %v - %s", err, dir)
				//os.Exit(1)
			}
		}
	}
	fmt.Println("✅ Estrutura de diretórios criada!")
	return nil
}
func (f *FileSystemServiceImpl) GetFromFile(filePathArg string, fileType *string, data interface{}) (object interface{}, err error) {
	filePath := f.SanitizePath(filePathArg)
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("❌ Arquivo não encontrado: %s", filePath)
	}
	file, fileErr := os.ReadFile(filePath)
	if fileErr != nil {
		return nil, fmt.Errorf("❌ Erro ao ler arquivo: %v", fileErr)
	}
	switch *fileType {
	case `json`:
		if jsonErr := json.Unmarshal(file, &data); jsonErr != nil {
			return nil, fmt.Errorf("❌ Erro ao deserializar dados: %v", jsonErr)
		}
	case `xml`:
		if xmlErr := xml.Unmarshal(file, &data); xmlErr != nil {
			return nil, fmt.Errorf("❌ Erro ao deserializar dados: %v", xmlErr)
		}
	case `yaml`:
		if yamlErr := yaml.Unmarshal(file, &data); yamlErr != nil {
			return nil, fmt.Errorf("❌ Erro ao deserializar dados: %v", yamlErr)
		}
	default:
		return file, nil
	}
	return data, nil
}
func (f *FileSystemServiceImpl) WriteToFile(filePathArg string, data interface{}, fileType *string) error {
	filePath := f.SanitizePath(filePathArg)
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		if _, createErr := os.Create(filePath); createErr != nil {
			return fmt.Errorf("❌ Erro ao criar arquivo: %v", createErr)
		}
	}
	var content []byte
	var contentErr error
	if fileType != nil {
		switch *fileType {
		case `json`:
			content, contentErr = json.Marshal(data)
		case `xml`:
			content, contentErr = xml.Marshal(data)
		case `yaml`:
			content, contentErr = yaml.Marshal(data)
		default:
			content = data.([]byte)
		}
		if contentErr != nil {
			return fmt.Errorf("❌ Erro ao serializar dados: %v", contentErr)
		}
	} else {
		content = data.([]byte)
	}
	if writeErr := os.WriteFile(filePath, content, 0644); writeErr != nil {
		return fmt.Errorf("❌ Erro ao escrever no arquivo: %v", writeErr)
	}
	return nil
}
func (f *FileSystemServiceImpl) GetConfigFilePath() string           { return f.configFilePath }
func (f *FileSystemServiceImpl) GetDefaultCacheDir() string          { return f.cacheDir }
func (f *FileSystemServiceImpl) GetDefaultRedisVolumeDir() string    { return f.redisVolume }
func (f *FileSystemServiceImpl) GetDefaultMongoVolumeDir() string    { return f.mongoVolume }
func (f *FileSystemServiceImpl) GetDefaultRabbitMQVolumeDir() string { return f.rabbitMQVolume }
func (f *FileSystemServiceImpl) GetDefaultPostgresVolumeDir() string { return f.postgresVolume }
func (f *FileSystemServiceImpl) GetDefaultKubexDir() string          { return f.kubexDir }
func (f *FileSystemServiceImpl) GetDefaultVaultDir() string          { return f.vaultDir }
func (f *FileSystemServiceImpl) GetDefaultKbxDir() string            { return f.kbxDir }
func (f *FileSystemServiceImpl) GetDefaultRootDir() string           { return f.rootDir }
func (f *FileSystemServiceImpl) GetDefaultConfigDir() string         { return f.configDir }
func (f *FileSystemServiceImpl) GetDefaultKeyPath() string           { return f.keyPath }
func (f *FileSystemServiceImpl) GetDefaultCertPath() string          { return f.certPath }

func NewFileSystemSrv(configFile string) FileSystemService {
	if configFile == "" {
		configFile = globals.DefaultGoSpiderConfigPath
	}
	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		log.Fatal(fmt.Sprintf("❌ Erro ao obter o diretório do usuário: %v", homeDirErr))
		return nil
	}

	return &FileSystemServiceImpl{
		configFilePath: strings.Replace(configFile, "$HOME", homeDir, -1),
		cacheDir:       strings.Replace(globals.DefaultCacheDir, "$HOME", homeDir, -1),
		redisVolume:    strings.Replace(globals.DefaultRedisVolume, "$HOME", homeDir, -1),
		mongoVolume:    strings.Replace(globals.DefaultMongoVolume, "$HOME", homeDir, -1),
		rabbitMQVolume: strings.Replace(globals.DefaultRabbitMQVolume, "$HOME", homeDir, -1),
		postgresVolume: strings.Replace(globals.DefaultPostgresVolume, "$HOME", homeDir, -1),
		kubexDir:       strings.Replace(globals.DefaultKubexDir, "$HOME", homeDir, -1),
		vaultDir:       strings.Replace(globals.DefaultVaultDir, "$HOME", homeDir, -1),
		kbxDir:         strings.Replace(globals.DefaultKbxDir, "$HOME", homeDir, -1),
		rootDir:        strings.Replace(globals.DefaultGoSpiderDir, "$HOME", homeDir, -1),
		configDir:      strings.Replace(globals.DefaultGoSpiderConfigDir, "$HOME", homeDir, -1),
		keyPath:        strings.Replace(globals.DefaultKeyPath, "$HOME", homeDir, -1),
		certPath:       strings.Replace(globals.DefaultCertPath, "$HOME", homeDir, -1),
	}
}
