package services

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/pebbe/zmq4"
	"github.com/spf13/viper"
	"log"
	"sync"
)

const brokerEndpoint = "tcp://*:5555"

var once sync.Once

type Broker interface {
	HandleRouter()
	HandleSub()
}
type BrokerImpl struct {
	context *zmq4.Context
	pub     *zmq4.Socket
	sub     *zmq4.Socket
	router  *zmq4.Socket
}

var brokerInstance *BrokerImpl

func StartBroker() {
	once.Do(func() {
		context, _ := zmq4.NewContext()
		pub, _ := context.NewSocket(zmq4.PUB)
		sub, _ := context.NewSocket(zmq4.SUB)
		router, _ := context.NewSocket(zmq4.ROUTER)

		defer pub.Close()
		defer sub.Close()
		defer router.Close()

		_ = pub.Bind("tcp://*:5556")
		_ = sub.Bind("tcp://*:5557")
		_ = router.Bind(brokerEndpoint)

		sub.SetSubscribe("")

		log.Println("ZeroMQ BrokerImpl iniciado com autenticação JWT")
		brokerInstance = &BrokerImpl{
			context: context,
			pub:     pub,
			sub:     sub,
			router:  router,
		}

		go brokerInstance.HandleRouter()
		go brokerInstance.HandleSub()
	})
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

		log.Printf("Mensagem recebida de [%s]: %s", identity, message)

		// Autenticação
		if !ValidateJWT(message) {
			log.Printf("Mensagem rejeitada por falha na autenticação [%s]", identity)
			continue
		}

		_, sendErr := b.pub.Send(message, 0)
		if sendErr != nil {
			log.Printf("Erro ao enviar mensagem: %v", sendErr)
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

func NewBroker() *BrokerImpl {
	if brokerInstance == nil {
		StartBroker()
	}
	return brokerInstance
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
