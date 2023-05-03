// Package server implements data plane server.
package server

import (
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/catalog"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/configurator"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/logger"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/workerpool"
)

var (
	log = logger.New("ecnet-bridge")
)

// Server implements the Aggregate Discovery Services
type Server struct {
	catalog        catalog.MeshCataloger
	ecnetNamespace string
	cfg            configurator.Configurator
	ready          bool
	workQueues     *workerpool.WorkerPool
	kubeController k8s.Controller
	msgBroker      *messaging.Broker
	dnsEndpoints   map[string]string
	dnsResolves    map[string]uint8
}
