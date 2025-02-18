package services

import "github.com/faelmori/gokubexfs/internal/globals"

type Server struct {
	Port         string `gorm:"not null" json:"port"`
	BindAddress  string `gorm:"not null" json:"bind_address"`
	ReadTimeout  int    `gorm:"not null" json:"read_timeout"`
	WriteTimeout int    `gorm:"not null" json:"write_timeout"`
}

type Database struct {
	Type             string      `gorm:"not null" json:"type"`
	Driver           string      `gorm:"not null" json:"driver"`
	ConnectionString string      `gorm:"omitempty" json:"connection_string"`
	Dsn              string      `gorm:"omitempty" json:"dsn"`
	Path             string      `gorm:"omitempty" json:"path"`
	Host             string      `gorm:"omitempty" json:"host"`
	Port             interface{} `gorm:"omitempty" json:"port"`
	Username         string      `gorm:"omitempty" json:"username"`
	Password         string      `gorm:"omitempty" json:"password"`
	Name             string      `gorm:"omitempty" json:"name"`
}
type JWT struct {
	RefreshSecret         string `gorm:"omitempty" json:"refresh_secret"`
	PrivateKey            string `gorm:"omitempty" json:"private_key"`
	PublicKey             string `gorm:"omitempty" json:"public_key"`
	ExpiresIn             int    `gorm:"omitempty" json:"expires_in"`
	IDExpirationSecs      int    `gorm:"omitempty" json:"id_expiration_secs"`
	RefreshExpirationSecs int    `gorm:"omitempty" json:"refresh_expiration_secs"`
}
type Redis struct {
	Enabled  bool   `gorm:"default:true" json:"enabled"`
	Addr     string `gorm:"omitempty" json:"addr"`
	Port     int    `gorm:"omitempty" json:"port"`
	Username string `gorm:"omitempty" json:"username"`
	Password string `gorm:"omitempty" json:"password"`
	DB       int    `gorm:"omitempty" json:"db"`
}
type RabbitMQ struct {
	Enabled        bool   `gorm:"default:true" json:"enabled"`
	Username       string `gorm:"omitempty" json:"username"`
	Password       string `gorm:"omitempty" json:"password"`
	Port           int    `gorm:"omitempty" json:"port"`
	ManagementPort int    `gorm:"omitempty" json:"management_port"`
}
type MongoDB struct {
	Enabled  bool   `json:"enabled;default:false"`
	Host     string `json:"host;default:localhost"`
	Port     int    `json:"port;default:22022"`
	Username string `json:"username;omitempty"`
	Password string `json:"password;omitempty"`
}

type Certificate struct {
	keyPath  string
	certPath string
	keyring  string
}
type Docker struct{}
type FileSystem struct {
	configFilePath string
	cacheDir       string
	redisVolume    string
	mongoVolume    string
	rabbitMQVolume string
	postgresVolume string
	kubexDir       string
	vaultDir       string
	kbxDir         string
	goSpyderDir    string
	rootDir        string
	configDir      string
	keyPath        string
	certPath       string
}
type Cache struct {
	Enabled          bool   `json:"enabled"`
	Setup            bool   `json:"setup"`
	CacheDir         string `json:"cache_dir"`
	SetupFlagPath    string `json:"setup_flag_path"`
	DepsFlagPath     string `json:"deps_flag_path"`
	ServicesFlagPath string `json:"services_flag_path"`
	VaultFlagPath    string `json:"vault_flag_path"`
}

var (
	ErrUsernameRequired = &globals.ValidationError{Field: "username", Message: "Username is required"}
	ErrPasswordRequired = &globals.ValidationError{Field: "password", Message: "Password is required"}
	ErrEmailRequired    = &globals.ValidationError{Field: "email", Message: "Email is required"}
	ErrDBNotProvided    = &globals.ValidationError{Field: "db", Message: "Database not provided"}
)
