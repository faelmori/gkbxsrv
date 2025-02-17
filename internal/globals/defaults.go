package globals

const KeyringService = "kubex"
const KeyringKey = "kubex_tls_pass"

const DefaultCacheDir = "$HOME/.cache/kubex"
const DefaultRedisVolume = "$HOME/.kubex/volumes/redis"
const DefaultMongoVolume = "$HOME/.kubex/volumes/mongo"
const DefaultRabbitMQVolume = "$HOME/.kubex/volumes/rabbitmq"
const DefaultPostgresVolume = "$HOME/.kubex/volumes/postgresql"
const DefaultKubexDir = "$HOME/.kubex"
const DefaultVaultDir = "$HOME/.kubex/.vault"
const DefaultKbxDir = "$HOME/.kubex/kbx"
const DefaultGoSpyderDir = "$HOME/.kubex/gospyder"
const DefaultGoSpyderConfigDir = "$HOME/.kubex/gospyder/config"
const DefaultGoSpyderConfigPath = "$HOME/.kubex/gospyder/config/config.json"
const DefaultKeyPath = "$HOME/.kubex/kubex-key.pem"
const DefaultCertPath = "$HOME/.kubex/kubex-cert.pem"

const (
	KubexConfigStructureFlag = "kubex_config_structure"
	KubexCertificatesFlag    = "kubex_certificates"
	KubexRedisPasswordFlag   = "kubex_redis_password"
	KubexRefreshSecretFlag   = "kubex_refresh_secret"
	KubexDBPasswordFlag      = "kubex_db_password"
	KubexCacheSetupFlag      = "kubex_cache_setup"
	KubexServicesSetupFlag   = "kubex_services_setup"
	KubexVaultSetupFlag      = "kubex_vault_setup"
	KubexDepsSetupFlag       = "kubex_deps_setup"
)

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}
func (v *ValidationError) FieldError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) FieldsError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) ErrorOrNil() error {
	return v
}
