package services

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/logz"
	"github.com/pebbe/zmq4"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	HeartbeatInterval = 2500 * time.Millisecond // Intervalo entre heartbeats
)

// BrokerImpl gerencia a comunicação entre clientes e workers.
type BrokerImpl struct {
	context     *zmq4.Context
	frontend    *zmq4.Socket // FRONTEND (ROUTER) para clientes
	backend     *zmq4.Socket // BACKEND (DEALER) para workers
	services    map[string]*Service
	workers     map[string]*Worker
	waiting     []*Worker
	mu          sync.Mutex
	heartbeatAt time.Time
	verbose     bool
}

// Service representa um serviço com suas requisições e workers
type Service struct {
	name     string
	requests [][]string
	waiting  []*Worker
}

// Worker representa um worker com identidade, vencimento e referência ao broker
type Worker struct {
	identity string
	service  *Service
	expiry   time.Time
	broker   *BrokerImpl
}

// NewBroker configura o contexto e os sockets, e inicia o pool de workers e o proxy.
func NewBroker(verbose bool) (*BrokerImpl, error) {
	ctx, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("error creating ZMQ context: %v", err)
	}

	// Criação do FRONTEND (ROUTER) para comunicação com clientes
	frontend, err := ctx.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, fmt.Errorf("error creating FRONTEND (ROUTER): %v", err)
	}
	frontend.SetRouterMandatory(1)
	frontend.SetRouterHandover(true)
	frontend.SetRcvtimeo(5 * time.Second)
	frontend.SetSndtimeo(5 * time.Second)
	if err := frontend.Bind("tcp://0.0.0.0:5555"); err != nil {
		return nil, fmt.Errorf("error binding FRONTEND (ROUTER): %v", err)
	}

	// Criação do BACKEND (DEALER) para comunicação com workers – usaremos inproc e proxy
	backend, err := ctx.NewSocket(zmq4.DEALER)
	if err != nil {
		return nil, fmt.Errorf("error creating BACKEND (DEALER): %v", err)
	}
	if err := backend.Bind("inproc://backend"); err != nil {
		return nil, fmt.Errorf("error binding BACKEND (DEALER): %v", err)
	}

	broker := &BrokerImpl{
		context:     ctx,
		frontend:    frontend,
		backend:     backend,
		services:    make(map[string]*Service),
		workers:     make(map[string]*Worker),
		waiting:     []*Worker{},
		heartbeatAt: time.Now().Add(HeartbeatInterval),
		verbose:     verbose,
	}

	// Inicia um pool de workers
	for i := 0; i < 5; i++ {
		go broker.workerTask()
	}

	// Inicia o proxy para conectar FRONTEND e BACKEND
	go broker.startProxy()

	// Opcional: iniciar gerenciamento de heartbeats
	// go broker.handleHeartbeats()

	return broker, nil
}

// startProxy usa o built-in Proxy para interligar FRONTEND e BACKEND
func (b *BrokerImpl) startProxy() {
	logz.Logger.Info("Starting proxy between FRONTEND and BACKEND...", nil)
	err := zmq4.Proxy(b.frontend, b.backend, nil)
	if err != nil {
		logz.Logger.Error("Error in proxy between FRONTEND and BACKEND", map[string]interface{}{"error": err})
	}
}

// A função processCommand usa o ModelRegistry para interpretar o payload e executar
// um comando dinâmico. Aqui, simulamos chamadas aos métodos básicos do repositório.
func (b *BrokerImpl) processCommand(cmd, payload string) string {
	// Desserializa o payload usando ModelRegistry
	registry, err := models.NewModelRegistryFromSerialized([]byte(payload))
	if err != nil {
		logz.Logger.Error("Error deserializing payload in processCommand", map[string]interface{}{
			"payload": payload, "error": err,
		})
		return fmt.Sprintf(`{"error":"deserialization error: %v"}`, err)
	}

	command := strings.ToLower(registry.GetCommand())
	if command == "" {
		command = "findall"
	}

	modelInstance := registry.ToModel()

	repo, err := GetRepoForModel(modelInstance)
	if err != nil {
		logz.Logger.Error("Error obtaining repository for model", map[string]interface{}{
			"error": err,
		})
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}

	// Executa o comando dinamicamente
	result, err := repo.ExecuteCommand(command, modelInstance)
	if err != nil {
		logz.Logger.Error("Error executing command in repository", map[string]interface{}{
			"command": command, "error": err,
		})
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}

	return fmt.Sprintf(`{"status":"success","message":"Command '%s' executed","data":%v}`, command, result)
}

func (b *BrokerImpl) processCommandDynamic(payloadStr string, repo interface{}) string {
	// Cria uma instância do ModelRegistry para obter o comando
	registry, err := models.NewModelRegistryFromSerialized([]byte(payloadStr))
	if err != nil {
		logz.Logger.Error("Error deserializing payload in processCommandDynamic", map[string]interface{}{
			"payload": payloadStr,
			"error":   err,
		})
		return fmt.Sprintf(`{"error":"deserialization error: %v"}`, err)
	}

	command := registry.GetCommand()
	if command == "" {
		command = "findAll"
	}
	// Obtém a instância do modelo com base no payload
	modelInstance := registry.ToModel()

	// Chama dinamicamente o método do repositório
	results, err := callRepositoryMethod(repo, command, modelInstance)
	if err != nil {
		logz.Logger.Error("Error calling repository method", map[string]interface{}{
			"command": command,
			"error":   err,
		})
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}

	// Aqui você pode assumir que o método retornou pelo menos um resultado,
	// e formatar a resposta (essa parte varia de acordo com seu design)
	// Por exemplo, se o método tem assinatura (User, error)
	if len(results) >= 2 {
		if !results[1].IsNil() {
			errVal := results[1].Interface().(error)
			return fmt.Sprintf(`{"error":"%v"}`, errVal)
		}

		data := results[0].Interface()
		return fmt.Sprintf(`{"status":"success","message":"%s executed","data":%v}`, command, data)
	}

	return fmt.Sprintf(`{"error":"Unexpected number of results"}`)
}

// workerTask simula um worker que processa mensagens do BACKEND.
func (b *BrokerImpl) workerTask() {
	worker, err := b.context.NewSocket(zmq4.DEALER)
	if err != nil {
		logz.Logger.Error("Error creating socket for worker", map[string]interface{}{"error": err})
		return
	}
	// Comentado defer worker.Close() para persistir o socket durante o ciclo do worker
	if err := worker.Connect("inproc://backend"); err != nil {
		logz.Logger.Error("Error connecting worker to BACKEND", map[string]interface{}{"error": err})
		return
	}

	for {
		msg, err := worker.RecvMessage(0)
		if err != nil {
			logz.Logger.Error("Error receiving message in worker", map[string]interface{}{"error": err})
			continue
		}
		if len(msg) < 2 {
			logz.Logger.Debug("Malformed message received in WORKER", nil)
			continue
		}

		// Separa a identidade (que será usada para encaminhar a resposta) do payload
		id, rest := splitMessage(msg)
		payload := rest[len(rest)-1]

		// Usa ModelRegistry para desserializar o payload
		registry, err := models.NewModelRegistryFromSerialized([]byte(payload))
		if err != nil {
			logz.Logger.Error("Error deserializing payload in WORKER", map[string]interface{}{
				"payload": payload, "error": err,
			})
			continue
		}

		logz.Logger.Debug("Payload deserialized in WORKER", map[string]interface{}{
			"payload": registry.ToModel(),
		})

		tp, err := registry.GetType()
		if err != nil {
			logz.Logger.Error("Error getting payload type in WORKER", map[string]interface{}{
				"error": err,
			})
			continue
		}
		logz.Logger.Debug("Payload type in WORKER", map[string]interface{}{
			"tp":      tp.Name(),
			"payload": registry.ToModel(),
		})

		// Processa o comando baseado no tipo e/ou comando
		switch strings.ToLower(tp.Name()) {
		case "pingimpl":
			// Responde ao comando "ping"
			response := fmt.Sprintf(`{"type":"ping","data":{"ping":"%v"}}`, "pong")
			if _, err := worker.SendMessage(id, response); err != nil {
				logz.Logger.Error("Error sending ping response to BACKEND in WORKER", map[string]interface{}{
					"response": response, "error": err,
				})
			} else {
				logz.Logger.Debug("Ping response sent to BACKEND in WORKER", map[string]interface{}{
					"response": response,
				})
			}
		default:
			// Se não for ping, processa como comando dinâmico:
			cmdResponse := b.processCommand(registry.GetCommand(), payload)
			if _, err := worker.SendMessage(id, cmdResponse); err != nil {
				logz.Logger.Error("Error sending command response to BACKEND in WORKER", map[string]interface{}{
					"response": cmdResponse, "error": err,
				})
			} else {
				logz.Logger.Debug("Command response sent to BACKEND in WORKER", map[string]interface{}{
					"response": cmdResponse,
				})
			}
		}
	}
}

// splitMessage separa os primeiros frames (identidade) do restante da mensagem.
func splitMessage(recPayload []string) (id, rest []string) {
	if len(recPayload) > 1 && recPayload[1] == "" {
		id = recPayload[:2]
		rest = recPayload[2:]
	} else {
		id = recPayload[:1]
		rest = recPayload[1:]
	}
	return
}

// Suponha que command seja uma string (por exemplo, "create") e que o repositório tenha um método Create.
func callRepositoryMethod(repo interface{}, command string, args ...interface{}) ([]reflect.Value, error) {
	// Converte o comando para o nome do método (primeira letra maiúscula)
	methodName := strings.Title(strings.ToLower(command))

	// Obtém o valor reflect do repositório
	repoValue := reflect.ValueOf(repo)

	// Obter o método pelo nome
	method := repoValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("método %s não encontrado", methodName)
	}

	// Prepara os argumentos para a chamada: cada argumento em um reflect.Value
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// Chama o método
	results := method.Call(in)
	return results, nil
}
