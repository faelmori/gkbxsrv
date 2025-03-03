package services

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/logz"
	"github.com/goccy/go-json"
	"log"
	"reflect"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/pebbe/zmq4"
	"github.com/spf13/viper"
)

const brokerEndpoint = "tcp://0.0.0.0:5555"

var once sync.Once
var brokerInstance *BrokerImpl

type Broker interface {
	HandleRouter()
	HandleSub()
	//handleMessage(msg []byte) map[string]interface{}
	GetDBService() IDatabaseService
	SetDBService(dbService IDatabaseService)
	GetServerCfg() ConfigService
	SetServerCfg(serverCfg ConfigService)
}
type BrokerImpl struct {
	context   *zmq4.Context
	pub       *zmq4.Socket
	sub       *zmq4.Socket
	router    *zmq4.Socket
	dbService IDatabaseService
	serverCfg ConfigService
}

func StartBroker(cfg ConfigService, dbService IDatabaseService) {
	once.Do(func() {
		context, _ := zmq4.NewContext()
		pub, _ := context.NewSocket(zmq4.PUB)
		sub, _ := context.NewSocket(zmq4.SUB)
		router, _ := context.NewSocket(zmq4.ROUTER)

		_ = pub.Bind("tcp://*:5556")
		_ = sub.Bind("tcp://*:5557")
		_ = router.Bind(brokerEndpoint)

		subscribeErr := sub.SetSubscribe("")
		if subscribeErr != nil {
			logz.Logger.Error("Erro ao configurar o subscriber", map[string]interface{}{"error": subscribeErr.Error()})
			return
		}

		log.Println("ZeroMQ BrokerImpl iniciado com autenticação JWT")

		brokerInstance = &BrokerImpl{
			context:   context,
			pub:       pub,
			sub:       sub,
			router:    router,
			serverCfg: cfg,
			dbService: dbService,
		}

		go brokerInstance.HandleRouter()
		go brokerInstance.HandleSub()
	})
}
func NewBroker(cfg ConfigService, dbService IDatabaseService) Broker {
	if brokerInstance == nil {
		StartBroker(cfg, dbService)
	}
	return brokerInstance
}

func (b *BrokerImpl) handleMessage(msg []byte) map[string]interface{} {
	logz.Logger.Info("Mensagem recebida no broker", map[string]interface{}{
		"raw_message": string(msg),
	})

	// Processar a mensagem para deduzir o modelo
	model, err := processMessage(msg)
	if err != nil {
		logz.Logger.Error("Erro ao processar mensagem", map[string]interface{}{
			"error": err.Error(),
		})
		return map[string]interface{}{"error": err.Error()}
	}

	logz.Logger.Info("Modelo deduzido com sucesso", map[string]interface{}{
		"model_type": fmt.Sprintf("%T", model),
		"model_data": model,
	})

	// Exemplo de switch para lidar com o tipo deduzido
	switch m := model.(type) {
	case *models.UserImpl:
		logz.Logger.Info("Processando usuário", map[string]interface{}{
			"username": m.GetUsername(),
		})
		d, e := b.dbService.GetDB()
		if e != nil {
			logz.Logger.Error("Erro ao obter conexão com o banco de dados", map[string]interface{}{"error": e.Error()})
			return map[string]interface{}{"error": e.Error()}
		}
		userRepo := models.NewUserRepo(d)
		user, userErr := userRepo.FindOne([]string{"username", m.GetUsername()})
		if userErr != nil {
			logz.Logger.Warn("Usuário não encontrado", map[string]interface{}{
				"username": m.GetUsername(),
			})
			return map[string]interface{}{"status": "not_found"}
		}

		logz.Logger.Info("Usuário encontrado com sucesso", map[string]interface{}{
			"user_id":  user.GetID(),
			"username": user.GetUsername(),
			"email":    user.GetEmail(),
		})

		return map[string]interface{}{
			"status": "found",
			"user": map[string]interface{}{
				"id":       user.GetID(),
				"username": user.GetUsername(),
				"email":    user.GetEmail(),
			},
		}
	default:
		logz.Logger.Error("Tipo de modelo não suportado", map[string]interface{}{
			"model_type": fmt.Sprintf("%T", model),
		})
		return map[string]interface{}{"error": "tipo de model não suportado"}
	}
}
func (b *BrokerImpl) HandleRouter() {
	for {
		identity, err := b.router.Recv(0)
		if err != nil {
			log.Printf("Erro ao receber identidade: %v", err)
			continue
		}

		message, err := b.router.Recv(0)
		if err != nil {
			log.Printf("Erro ao receber mensagem: %v", err)
			continue
		}

		response := b.handleMessage([]byte(message))
		responseData, _ := json.Marshal(response)

		_, sendErr := b.router.Send(identity, zmq4.SNDMORE)
		if sendErr != nil {
			logz.Logger.Error("Erro ao enviar identidade", map[string]interface{}{"error": sendErr.Error()})
			return
		}
		_, respErr := b.router.Send(string(responseData), 0)
		if respErr != nil {
			logz.Logger.Error("Erro ao enviar resposta", map[string]interface{}{"error": respErr.Error()})
			return
		}
	}
}
func (b *BrokerImpl) HandleSub() {
	for {
		msg, err := b.sub.Recv(0)
		if err != nil {
			log.Printf("Erro ao receber no subscriber: %v", err)
			continue
		}
		log.Println("Broadcast recebido:", msg)
	}
}
func (b *BrokerImpl) GetDBService() IDatabaseService {
	return b.dbService
}
func (b *BrokerImpl) SetDBService(dbService IDatabaseService) {
	b.dbService = dbService
}
func (b *BrokerImpl) GetServerCfg() ConfigService {
	return b.serverCfg
}
func (b *BrokerImpl) SetServerCfg(serverCfg ConfigService) {
	b.serverCfg = serverCfg
}

func ConnectToBroker() (*zmq4.Socket, error) {
	socket, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		return nil, fmt.Errorf("Erro ao conectar ao broker: %v", err)
	}
	err = socket.Connect(brokerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Erro ao conectar ao broker: %v", err)
	}
	return socket, nil
}
func ValidateJWT(tokenString string) bool {
	cfg := NewConfigSrv(fs.GetDefaultKeyPath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
	pubKey := loadPublicKey(cfg)

	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		log.Println("Token inválido")
		return false
	}
	return true
}
func loadPrivateKey() (*rsa.PrivateKey, error) {
	keyDataB64 := viper.GetString("jwt.private_key")
	if keyDataB64 == "" {
		return nil, fmt.Errorf("error reading private key file: %v", "jwt.private_key is empty")
	}
	keyData, err := base64.StdEncoding.DecodeString(keyDataB64)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}
	pwd, pwdErr := crt.RetrievePassword()
	if pwdErr != nil {
		return nil, fmt.Errorf("error reading private key file: %v", pwdErr)
	}
	privateKey, privateKeyErr := crt.DecryptPrivateKey(keyData, []byte(pwd))
	if privateKeyErr != nil {
		return nil, fmt.Errorf("error parsing private key: %v", privateKeyErr)
	}
	return privateKey, nil
}
func loadPublicKey(cfg ConfigService) *rsa.PublicKey {
	if readErr := viper.ReadInConfig(); readErr != nil {
		log.Fatalf("Error reading public key file: %v", readErr)
	}
	if !fs.ExistsConfigFile() {
		log.Fatalf("Error reading public key file: %v", "config file not found")
	}

	if !cfg.IsConfigLoaded() {
		if loadErr := cfg.LoadConfig(); loadErr != nil {
			log.Fatalf("Error loading config file: %v", loadErr)
		}
	}

	keyDataB64 := viper.GetString("jwt.public_key")
	if keyDataB64 == "" {
		log.Fatalf("Error reading public key file: %v", "jwt.public_key is empty")
	}

	keyData, err := base64.StdEncoding.DecodeString(keyDataB64)
	if err != nil {
		log.Fatalf("Error reading public key file: %v", err)
	}

	publicKey, publicKeyErr := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if publicKeyErr != nil {
		fmt.Println(publicKey)
		fmt.Println(publicKeyErr)
		log.Fatalf("Error parsing public key: %v", publicKeyErr)
	}

	return publicKey
}
func processMessage(msg []byte) (interface{}, error) {
	// Estrutura genérica da mensagem recebida
	var genericMessage struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	// Decodificar o JSON genérico
	if err := json.Unmarshal(msg, &genericMessage); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	// Verificar se o tipo está registrado no modelRegistry
	modelType, exists := models.ModelRegistry[genericMessage.Type]
	if !exists {
		return nil, fmt.Errorf("tipo não registrado: %s", genericMessage.Type)
	}

	// Criar uma instância do tipo registrado dinamicamente
	modelInstance := reflect.New(modelType).Interface()

	// Re-serializar os dados e desserializar no modelo
	dataBytes, err := json.Marshal(genericMessage.Data)
	if err != nil {
		return nil, fmt.Errorf("erro ao re-serializar os dados: %v", err)
	}

	if err := json.Unmarshal(dataBytes, modelInstance); err != nil {
		return nil, fmt.Errorf("erro ao desserializar dados no modelo: %v", err)
	}

	return modelInstance, nil
}
