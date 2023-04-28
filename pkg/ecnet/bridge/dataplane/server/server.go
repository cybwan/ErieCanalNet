package server

import (
	"github.com/cilium/ebpf/rlimit"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/catalog"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/cni/controller/helpers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/configurator"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/workerpool"
)

const (
	// ServerType is the type identifier for the ctrl server
	ServerType = "ecnet-bridge"

	// workerPoolSize is the default number of workerpool workers (0 is GOMAXPROCS)
	workerPoolSize = 0
)

// NewBridgeServer creates a new bridge server
func NewBridgeServer(meshCatalog catalog.MeshCataloger, ecnetNamespace string, cfg configurator.Configurator, kubecontroller k8s.Controller, msgBroker *messaging.Broker) *Server {
	server := Server{
		catalog:        meshCatalog,
		ecnetNamespace: ecnetNamespace,
		cfg:            cfg,
		workQueues:     workerpool.NewWorkerPool(workerPoolSize),
		kubeController: kubecontroller,
		msgBroker:      msgBroker,
	}
	return &server
}

// Start starts the codebase push server
func (s *Server) Start(kernelTracing bool, bridgeEth string) error {
	if err := helpers.LoadProgs(kernelTracing, bridgeEth); err != nil {
		log.Fatal().Msgf("failed to load ebpf programs: %v", err)
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal().Msgf("remove memlock error: %v", err)
	}
	s.ready = true
	return nil
}
