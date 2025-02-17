package services

type Cache struct {
	Enabled          bool   `json:"enabled"`
	Setup            bool   `json:"setup"`
	CacheDir         string `json:"cache_dir"`
	SetupFlagPath    string `json:"setup_flag_path"`
	DepsFlagPath     string `json:"deps_flag_path"`
	ServicesFlagPath string `json:"services_flag_path"`
	VaultFlagPath    string `json:"vault_flag_path"`
}
