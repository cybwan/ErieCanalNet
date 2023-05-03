package registry

import (
	"sync"

	"github.com/google/uuid"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/bridge/ctrlplane"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
)

// NewProxyRegistry initializes a new empty *ProxyRegistry.
func NewProxyRegistry(msgBroker *messaging.Broker) *ProxyRegistry {
	return &ProxyRegistry{
		msgBroker: msgBroker,
	}
}

// RegisterProxy registers a newly connected proxy.
func (pr *ProxyRegistry) RegisterProxy() *ctrlplane.Proxy {
	lock.Lock()
	defer lock.Unlock()
	if pr.cacheProxy == nil {
		pr.cacheProxy = &ctrlplane.Proxy{Mutex: new(sync.RWMutex)}
		pr.cacheProxy.UUID, _ = uuid.NewUUID()
		pr.cacheProxy.Quit = make(chan bool)
	}
	return pr.cacheProxy
}

// GetConnectedProxy loads a connected proxy from the registry.
func (pr *ProxyRegistry) GetConnectedProxy() *ctrlplane.Proxy {
	lock.Lock()
	defer lock.Unlock()
	return pr.cacheProxy
}
