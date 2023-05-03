package server

import (
	"fmt"

	"github.com/cilium/ebpf/rlimit"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/bridge/dataplane/helpers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/catalog"
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
		dnsEndpoints:   make(map[string]string),
		dnsResolves:    make(map[string]uint8),
	}
	return &server
}

// Start starts the ecnet bridge server
func (s *Server) Start(kernelTracing bool, bridgeEth string) error {
	if err := helpers.LoadProgs(kernelTracing, bridgeEth); err != nil {
		log.Fatal().Msgf("failed to load ebpf programs: %v", err)
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal().Msgf("remove memlock error: %v", err)
	}
	if err := helpers.InitLoadPinnedMap(); err != nil {
		return fmt.Errorf("failed to load ebpf maps: %v", err)
	}
	if err := helpers.AttachProgs(); err != nil {
		return fmt.Errorf("failed to attach ebpf programs: %v", err)
	}

	// Start broadcast listener thread
	go s.broadcastListener()

	s.ready = true

	return nil
}

// Stop stops the ecnet bridge server
func (s *Server) Stop() error {
	if err := helpers.UnLoadProgs(); err != nil {
		return fmt.Errorf("unload failed: %v", err)
	}
	return nil
}
