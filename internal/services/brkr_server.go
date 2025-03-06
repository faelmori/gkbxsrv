package services

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/gkbxsrv/logz"
	"github.com/pebbe/zmq4"
	"sync"
	"time"
)

const (
	HeartbeatLiveness = 3                       // Tentativas de heartbeat antes de expirar
	HeartbeatInterval = 2500 * time.Millisecond // Intervalo entre heartbeats
	HeartbeatExpiry   = HeartbeatInterval * HeartbeatLiveness
)

// Estrutura do Broker
type BrokerImpl struct {
	context     *zmq4.Context
	frontend    *zmq4.Socket // FRONTEND (ROUTER) para comunicação com clientes
	backend     *zmq4.Socket // BACKEND (DEALER) para comunicação com workers
	services    map[string]*Service
	workers     map[string]*Worker
	waiting     []*Worker
	mu          sync.Mutex
	heartbeatAt time.Time
	verbose     bool
}

// Estrutura para um Serviço
type Service struct {
	name     string
	requests [][]string
	waiting  []*Worker
}

// Estrutura para um Worker
type Worker struct {
	identity string
	service  *Service
	expiry   time.Time
	broker   *BrokerImpl
}

// Criação do Broker
func NewBroker(verbose bool) (*BrokerImpl, error) {
	ctx, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("erro ao criar contexto ZMQ: %v", err)
	}

	// Criação do FRONTEND (ROUTER)
	frontend, err := ctx.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar FRONTEND (ROUTER): %v", err)
	}
	frontend.SetRouterMandatory(1)
	frontend.SetRouterHandover(true)
	if err := frontend.Bind("tcp://0.0.0.0:5555"); err != nil {
		return nil, fmt.Errorf("erro ao vincular FRONTEND (ROUTER): %v", err)
	}

	// Criação do BACKEND (DEALER)
	backend, err := ctx.NewSocket(zmq4.DEALER)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar BACKEND (DEALER): %v", err)
	}
	if err := backend.Bind("inproc://backend"); err != nil {
		return nil, fmt.Errorf("erro ao vincular BACKEND (DEALER): %v", err)
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

	// Lança os workers
	for i := 0; i < 5; i++ {
		go broker.workerTask()
	}

	// Inicia o proxy
	go broker.startProxy()

	// Inicia o gerenciamento de heartbeats
	go broker.handleHeartbeats()

	return broker, nil
}

// Proxy para conectar FRONTEND (ROUTER) e BACKEND (DEALER)
func (b *BrokerImpl) startProxy() {
	logz.Logger.Info("Iniciando proxy entre FRONTEND e BACKEND...", nil)
	err := zmq4.Proxy(b.frontend, b.backend, nil)
	if err != nil {
		logz.Logger.Error("Erro no proxy entre FRONTEND e BACKEND", map[string]interface{}{
			"error": err,
		})
	}
}

// Lógica do Worker (simulado)
func (b *BrokerImpl) workerTask() {
	worker, err := b.context.NewSocket(zmq4.DEALER)
	if err != nil {
		logz.Logger.Error("Erro ao criar socket para worker", map[string]interface{}{
			"error": err,
		})
		return
	}
	defer worker.Close()

	if err := worker.Connect("inproc://backend"); err != nil {
		logz.Logger.Error("Erro ao conectar worker ao BACKEND", map[string]interface{}{
			"error": err,
		})
		return
	}

	for {
		msg, err := worker.RecvMessage(0)
		if err != nil {
			logz.Logger.Error("Erro ao receber mensagem no WORKER", map[string]interface{}{
				"error": err,
			})
			continue
		}

		logz.Logger.Info("Mensagem recebida no WORKER", map[string]interface{}{
			"context": "workerTask",
			"frames":  len(msg),
			"message": msg,
		})

		// Valida se a mensagem contém pelo menos 2 frames
		if len(msg) < 2 {
			logz.Logger.Warn("Mensagem malformada recebida no WORKER", nil)
			continue
		}

		// Processa o payload
		payload := msg[len(msg)-1]
		deserializedModel, deserializedModelErr := models.NewModelRegistryFromSerialized([]byte(payload))
		if deserializedModelErr != nil {
			logz.Logger.Error("Erro ao desserializar payload no WORKER", map[string]interface{}{
				"context": "workerTask",
				"payload": payload,
				"error":   deserializedModelErr.Error(),
			})
			continue
		}

		logz.Logger.Info("Payload desserializado no WORKER", map[string]interface{}{
			"context": "workerTask",
			"payload": deserializedModel.ToModel(),
		})

		tp, tpErr := deserializedModel.GetType()
		if tpErr != nil {
			logz.Logger.Error("Erro ao obter tipo do payload no WORKER", map[string]interface{}{
				"context":           "workerTask",
				"tp":                tp,
				"error":             tpErr,
				"deserializedModel": deserializedModel,
			})
			continue
		}

		logz.Logger.Info("Tipo do payload no WORKER", map[string]interface{}{
			"context": "workerTask",
			"tp":      tp.Name(),
			"payload": deserializedModel.ToModel(),
		})

		if tp.Name() == "PingImpl" {
			response := fmt.Sprintf(`{"type":"ping","data":{"ping":"%v"}}`, "pong")
			if _, workerSendMessageErr := worker.SendMessage(response); workerSendMessageErr != nil {
				logz.Logger.Error("Erro ao enviar resposta ao BACKEND no WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
					"error":    workerSendMessageErr,
				})
			} else {
				logz.Logger.Info("Resposta enviada ao BACKEND no WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
				})
			}
		} else {
			logz.Logger.Warn("Comando desconhecido no WORKER", map[string]interface{}{
				"context": "workerTask",
				"type":    tp.Name(),
				"payload": deserializedModel.ToModel(),
			})
		}
	}
}

// Gerenciamento de heartbeats
func (b *BrokerImpl) handleHeartbeats() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for id, worker := range b.workers {
			if now.After(worker.expiry) {
				logz.Logger.Warn(fmt.Sprintf("Worker expirado: %s", id), nil)
				delete(b.workers, id)
			}
		}
		b.mu.Unlock()
	}
}

// Fechamento do broker
func (b *BrokerImpl) Stop() {
	b.frontend.Close()
	b.backend.Close()
	b.context.Term()
	logz.Logger.Info("Broker encerrado", nil)
}
