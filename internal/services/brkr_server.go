package services

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/internal/models"
	"github.com/faelmori/logz"
	"github.com/goccy/go-json"
	"github.com/pebbe/zmq4"
	"os"
	"sync"
	"time"
)

const (
	HeartbeatInterval = 2500 * time.Millisecond // Interval between heartbeats
)

type BrokerImpl struct {
	context     *zmq4.Context
	frontend    *zmq4.Socket // FRONTEND (ROUTER) with clients
	backend     *zmq4.Socket // BACKEND (DEALER) with workers
	services    map[string]*Service
	workers     map[string]*Worker
	waiting     []*Worker
	mu          sync.Mutex
	heartbeatAt time.Time
	brokerInfo  *BrokerInfoLock
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

func NewBrokerConn(port string) (*zmq4.Socket, error) {
	ctx, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("error creating ZMQ context: %v", err)
	}

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

	if hostBindErr := frontend.Bind(`tcp://0.0.0.0:` + port); hostBindErr != nil {
		return nil, fmt.Errorf("error binding FRONTEND (ROUTER): %v", hostBindErr)
	}

	return frontend, nil
}
func NewBroker(verbose bool) (*BrokerImpl, error) {
	ctx, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("error creating ZMQ context: %v", err)
	}

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

	if hostBindErr := frontend.Bind(`tcp://0.0.0.0:5555`); hostBindErr != nil {
		return nil, fmt.Errorf("error binding FRONTEND (ROUTER): %v", hostBindErr)
	}

	backend, err := ctx.NewSocket(zmq4.DEALER)
	if err != nil {
		return nil, fmt.Errorf("error creating BACKEND (DEALER): %v", err)
	}
	if bindErr := backend.Bind("inproc://backend"); bindErr != nil {
		return nil, fmt.Errorf("error binding BACKEND (DEALER): %v", bindErr)
	}

	broker := &BrokerImpl{
		brokerInfo:  NewBrokerInfo(randomName(), "5555"),
		context:     ctx,
		frontend:    frontend,
		backend:     backend,
		services:    make(map[string]*Service),
		workers:     make(map[string]*Worker),
		waiting:     []*Worker{},
		heartbeatAt: time.Now().Add(HeartbeatInterval),
		verbose:     verbose,
	}

	if broker.brokerInfo == nil {
		logz.Error("Error creating broker", nil)
		return nil, fmt.Errorf("error creating broker: Empty broker info")
	}
	data, marshalErr := json.Marshal(broker.brokerInfo.GetBrokerInfo())
	if marshalErr != nil {
		logz.Error("Error marshalling broker info", map[string]interface{}{
			"error": marshalErr,
		})
		return nil, marshalErr
	}
	if writeErr := os.WriteFile(broker.brokerInfo.GetPath(), data, 0644); writeErr != nil {
		logz.Error("Error writing broker file", map[string]interface{}{
			"error": writeErr,
		})
		return nil, writeErr
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
	logz.Info("Starting proxy between FRONTEND and BACKEND...", nil)
	err := zmq4.Proxy(b.frontend, b.backend, nil)
	if err != nil {
		logz.Error("Error in proxy between FRONTEND and BACKEND", map[string]interface{}{
			"error": err,
		})
	}
}
func (b *BrokerImpl) workerTask() {
	worker, err := b.context.NewSocket(zmq4.DEALER)
	if err != nil {
		logz.Error("Error creating socket for worker", map[string]interface{}{
			"error": err,
		})
		return
	}

	if connErr := worker.Connect("inproc://backend"); connErr != nil {
		logz.Error("Error connecting worker to BACKEND", map[string]interface{}{
			"context":  "gkbxsrv",
			"showDate": true,
			"action":   "workerTask",
			"error":    connErr.Error(),
		})
		return
	}

	for {
		msg, _ := worker.RecvMessage(0)
		if len(msg) < 2 {
			logz.Debug("Malformed message received in WORKER", nil)
			continue
		}

		id, msg := splitMessage(msg)

		payload := msg[len(msg)-1]
		deserializedModel, deserializedModelErr := models.NewModelRegistryFromSerialized([]byte(payload))
		if deserializedModelErr != nil {
			logz.Error("Error deserializing payload in WORKER", map[string]interface{}{
				"context": "workerTask",
				"payload": payload,
				"error":   deserializedModelErr.Error(),
			})
			continue
		}

		logz.Debug("Payload deserialized in WORKER", map[string]interface{}{
			"context": "workerTask",
			"payload": deserializedModel.ToModel(),
		})

		tp, tpErr := deserializedModel.GetType()
		if tpErr != nil {
			logz.Error("Error getting payload type in WORKER", map[string]interface{}{
				"context":           "workerTask",
				"tp":                tp,
				"error":             tpErr,
				"deserializedModel": deserializedModel,
			})
			continue
		}

		logz.Debug("Payload type in WORKER", map[string]interface{}{
			"context": "workerTask",
			"tp":      tp.Name(),
			"payload": deserializedModel.ToModel(),
		})

		if tp.Name() == "PingImpl" {
			response := fmt.Sprintf(`{"type":"ping","data":{"ping":"%v"}}`, "pong")
			if _, workerSendMessageErr := worker.SendMessage(id, response); workerSendMessageErr != nil {
				logz.Error("Error sending response to BACKEND in WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
					"error":    workerSendMessageErr,
				})
			} else {
				logz.Debug("Response sent to BACKEND in WORKER", map[string]interface{}{
					"context":  "workerTask",
					"response": response,
				})
			}
		} else {
			logz.Debug("Unknown command in WORKER", map[string]interface{}{
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
				logz.Warn(fmt.Sprintf("Expired worker: %s", id), nil)
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
	logz.Info("Broker stopped", nil)
}
