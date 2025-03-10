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
	HeartbeatInterval = 2500 * time.Millisecond // Interval between heartbeats
)

type BrokerImpl struct {
	context     *zmq4.Context
	frontend    *zmq4.Socket // FRONTEND (ROUTER) for communication with clients
	backend     *zmq4.Socket // BACKEND (DEALER) for communication with workers
	services    map[string]*Service
	workers     map[string]*Worker
	waiting     []*Worker
	mu          sync.Mutex
	heartbeatAt time.Time
	verbose     bool
}

type Service struct {
	name     string
	requests [][]string
	waiting  []*Worker
}

type Worker struct {
	identity string
	service  *Service
	expiry   time.Time
	broker   *BrokerImpl
}

func NewBroker(verbose bool) (*BrokerImpl, error) {
	ctx, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("error creating ZMQ context: %v", err)
	}

	// Create FRONTEND (ROUTER)
	frontend, err := ctx.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, fmt.Errorf("error creating FRONTEND (ROUTER): %v", err)
	}
	frontendSetRouterMandatoryErr := frontend.SetRouterMandatory(1)
	if frontendSetRouterMandatoryErr != nil {
		return nil, frontendSetRouterMandatoryErr
	}
	frontendSetRouterHandoverErr := frontend.SetRouterHandover(true)
	if frontendSetRouterHandoverErr != nil {
		return nil, frontendSetRouterHandoverErr
	}
	if err := frontend.Bind("tcp://0.0.0.0:5555"); err != nil {
		return nil, fmt.Errorf("error binding FRONTEND (ROUTER): %v", err)
	}

	// Create BACKEND (DEALER)
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

	// Launch workers
	for i := 0; i < 5; i++ {
		go broker.workerTask()
	}

	// Start the proxy
	go broker.startProxy()

	// Start heartbeat management
	//go broker.handleHeartbeats()

	return broker, nil
}

func (b *BrokerImpl) startProxy() {
	logz.Logger.Info("Starting proxy between FRONTEND and BACKEND...", nil)
	err := zmq4.Proxy(b.frontend, b.backend, nil)
	if err != nil {
		logz.Logger.Error("Error in proxy between FRONTEND and BACKEND", map[string]interface{}{
			"error": err,
		})
	}
}

func (b *BrokerImpl) workerTask() {
	worker, err := b.context.NewSocket(zmq4.DEALER)
	if err != nil {
		logz.Logger.Error("Error creating socket for worker", map[string]interface{}{
			"error": err,
		})
		return
	}
	//defer worker.Close()

	if err := worker.Connect("inproc://backend"); err != nil {
		logz.Logger.Error("Error connecting worker to BACKEND", map[string]interface{}{
			"error": err,
		})
		return
	}

	for {
		msg, _ := worker.RecvMessage(0)
		if len(msg) < 2 {
			logz.Logger.Debug("Malformed message received in WORKER", nil)
			continue
		}

		id, msg := splitMessage(msg)

		payload := msg[len(msg)-1]
		deserializedModel, deserializedModelErr := models.NewModelRegistryFromSerialized([]byte(payload))
		if deserializedModelErr != nil {
			logz.Logger.Error("Error deserializing payload in WORKER", map[string]interface{}{
				"context": "workerTask",
				"payload": payload,
				"error":   deserializedModelErr.Error(),
			})
			continue
		}

		logz.Logger.Debug("Payload deserialized in WORKER", map[string]interface{}{
			"context": "workerTask",
			"payload": deserializedModel.ToModel(),
		})

		tp, tpErr := deserializedModel.GetType()
		if tpErr != nil {
			logz.Logger.Error("Error getting payload type in WORKER", map[string]interface{}{
				"context":           "workerTask",
				"tp":                tp,
				"error":             tpErr,
				"deserializedModel": deserializedModel,
			})
			continue
		}

		logz.Logger.Debug("Payload type in WORKER", map[string]interface{}{
			"context": "workerTask",
			"tp":      tp.Name(),
			"payload": deserializedModel.ToModel(),
		})

		if tp.Name() == "PingImpl" {
			response := fmt.Sprintf(`{"type":"ping","data":{"ping":"%v"}}`, "pong")
			if _, workerSendMessageErr := worker.SendMessage(id, response); workerSendMessageErr != nil {
				logz.Logger.Error("Error sending response to BACKEND in WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
					"error":    workerSendMessageErr,
				})
			} else {
				logz.Logger.Debug("Response sent to BACKEND in WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
				})
			}
		} else {
			logz.Logger.Debug("Unknown command in WORKER", map[string]interface{}{
				"context": "workerTask",
				"type":    tp.Name(),
				"payload": deserializedModel.ToModel(),
			})
		}
	}
}

func (b *BrokerImpl) handleHeartbeats() {
	ticker := time.NewTicker(HeartbeatInterval)
	//defer ticker.Stop()
	//defer b.mu.Unlock()

	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for id, worker := range b.workers {
			if now.After(worker.expiry) {
				logz.Logger.Warn(fmt.Sprintf("Expired worker: %s", id), nil)
				delete(b.workers, id)
			}
		}
		b.mu.Unlock()
	}

	b.mu.Unlock()
}

func (b *BrokerImpl) Stop() {
	_ = b.frontend.Close()
	_ = b.backend.Close()
	_ = b.context.Term()
	logz.Logger.Info("Broker stopped", nil)
}

func splitMessage(recPayload []string) (id, msg []string) {
	if recPayload[1] == "" {
		id = recPayload[:2]
		msg = recPayload[2:]
	} else {
		id = recPayload[:1]
		msg = recPayload[1:]
	}
	return
}
