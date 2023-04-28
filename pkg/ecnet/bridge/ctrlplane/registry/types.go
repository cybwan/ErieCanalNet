// Package registry implements handler's methods.
package registry

import (
	"sync"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/bridge/ctrlplane"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
)

var (
	lock sync.Mutex
)

// ProxyRegistry keeps track of Sidecar proxies
// from the control plane.
type ProxyRegistry struct {
	msgBroker  *messaging.Broker
	cacheProxy *ctrlplane.Proxy

	// Fire a inform to update proxies
	UpdateProxies func()
}
