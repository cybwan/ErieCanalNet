// Package main implements ecnet bridge.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/bridge/dataplane/server"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/catalog"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/configurator"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/constants"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/errcode"
	configClientset "github.com/flomesh-io/ErieCanal/pkg/ecnet/gen/client/config/clientset/versioned"
	multiclusterClientset "github.com/flomesh-io/ErieCanal/pkg/ecnet/gen/client/multicluster/clientset/versioned"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/health"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/httpserver"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s/events"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s/informers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service/endpoint"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service/multicluster"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service/providers/fsm"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service/providers/kube"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/signals"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/logger"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/version"
)

var (
	verbosity           string
	ecnetName           string // An ID that uniquely identifies an ECNET instance
	ecnetNamespace      string
	ecnetServiceAccount string
	ecnetConfigName     string
	ecnetVersion        string
	trustDomain         string
	kernelTracing       bool
	bridgeEth           string

	scheme = runtime.NewScheme()
)

var (
	flags = pflag.NewFlagSet(`ecnet-bridge`, pflag.ExitOnError)
	log   = logger.New("ecnet/main")
)

func init() {
	flags.StringVarP(&verbosity, "verbosity", "v", "info", "Set log verbosity level")
	flags.StringVar(&ecnetName, "ecnet-name", "", "ecnet name")
	flags.StringVar(&ecnetNamespace, "ecnet-namespace", "", "ecnet ctrlplane's namespace")
	flags.StringVar(&ecnetServiceAccount, "ecnet-service-account", "", "ecnet ctrlplane's service account")
	flags.StringVar(&ecnetConfigName, "ecnet-config-name", "ecnet-config", "Name of the ecnet Config")
	flags.StringVar(&ecnetVersion, "ecnet-version", "", "Version of ecnet")

	// TODO (#4502): Remove when we add full MRC support
	flags.StringVar(&trustDomain, "trust-domain", "cluster.local", "The trust domain to use as part of the common name when requesting new certificates")

	// Get some flags from commands
	flags.BoolVarP(&kernelTracing, "kernel-tracing", "d", false, "kernel tracing mode")
	flags.StringVar(&bridgeEth, "bridge-eth", "cni0", "bridge veth created by CNI")

	_ = clientgoscheme.AddToScheme(scheme)
	_ = admissionv1.AddToScheme(scheme)
}

func main() {
	log.Info().Msgf("Starting ecnet-bridge %s; %s; %s", version.Version, version.GitCommit, version.BuildDate)
	if err := parseFlags(); err != nil {
		log.Fatal().Err(err).Msg("Error parsing cmd line arguments")
	}
	if err := logger.SetLogLevel(verbosity); err != nil {
		log.Fatal().Err(err).Msg("Error setting log level")
	}

	// Initialize kube config and client
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating kube configs using in-cluster config")
	}
	kubeClient := kubernetes.NewForConfigOrDie(kubeConfig)
	configClient := configClientset.NewForConfigOrDie(kubeConfig)
	multiclusterClient := multiclusterClientset.NewForConfigOrDie(kubeConfig)

	// Initialize the generic Kubernetes event recorder and associate it with the ecnet-dataplane pod resource
	bridgePod, err := getECNETBridgePod(kubeClient)
	if err != nil {
		log.Fatal().Msg("Error fetching ecnet-ctrlplane pod")
	}
	eventRecorder := events.GenericEventRecorder()
	if err = eventRecorder.Initialize(bridgePod, kubeClient, ecnetNamespace); err != nil {
		log.Fatal().Msg("Error initializing generic event recorder")
	}

	k8s.SetTrustDomain(trustDomain)

	// This ensures CLI parameters (and dependent values) are correct.
	if err = validateCLIParams(); err != nil {
		events.GenericEventRecorder().FatalEvent(err, events.InvalidCLIParameters, "Error validating CLI parameters")
	}

	_, cancel := context.WithCancel(context.Background())
	stop := signals.RegisterExitHandlers(cancel)

	msgBroker := messaging.NewBroker(stop)

	informerCollection, err := informers.NewInformerCollection(ecnetName, stop,
		informers.WithKubeClient(kubeClient),
		informers.WithConfigClient(configClient, ecnetConfigName, ecnetNamespace),
		informers.WithMultiClusterClient(multiclusterClient),
	)
	if err != nil {
		events.GenericEventRecorder().FatalEvent(err, events.InitializationError, "Error creating informer collection")
	}

	// This component will be watching resources in the config.flomesh.io API group
	cfg := configurator.NewConfigurator(informerCollection, ecnetNamespace, ecnetConfigName, msgBroker)
	k8sClient := k8s.NewKubernetesController(informerCollection, msgBroker)
	mcController := multicluster.NewMultiClusterController(informerCollection, kubeClient, k8sClient, msgBroker)
	kubeProvider := kube.NewClient(k8sClient, cfg)
	mcProvider := fsm.NewClient(mcController, cfg)
	endpointsProviders := []endpoint.Provider{kubeProvider, mcProvider}
	serviceProviders := []service.Provider{kubeProvider, mcProvider}

	meshCatalog := catalog.NewMeshCatalog(
		k8sClient,
		mcController,
		stop,
		cfg,
		serviceProviders,
		endpointsProviders,
		msgBroker,
	)

	dataPlaneServer := server.NewBridgeServer(meshCatalog, ecnetNamespace, cfg, k8sClient, msgBroker)
	if err = dataPlaneServer.Start(kernelTracing, bridgeEth); err != nil {
		events.GenericEventRecorder().FatalEvent(err, events.InitializationError, "Error initializing proxy control server")
	}

	// Initialize ecnet's http service server
	httpServer := httpserver.NewHTTPServer(constants.ECNETHTTPServerPort)
	// Health/Liveness probes
	funcProbes := []health.Probes{dataPlaneServer}
	httpServer.AddHandlers(map[string]http.Handler{
		constants.ECNETControllerReadinessPath: health.ReadinessHandler(funcProbes, nil),
		constants.ECNETControllerLivenessPath:  health.LivenessHandler(funcProbes, nil),
	})
	// Version
	httpServer.AddHandler(constants.VersionPath, version.GetVersionHandler())

	// Start HTTP server
	err = httpServer.Start()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to start ECNET metrics/probes HTTP server")
	}

	// Start the global log level watcher that updates the log level dynamically
	go k8s.WatchAndUpdateLogLevel(msgBroker, stop)

	<-stop
	cancel()
	log.Info().Msgf("Stopping ecnet-bridge %s; %s; %s", version.Version, version.GitCommit, version.BuildDate)

	/*-----------------*/

	//// Initialize kube config and client
	//kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	//if err != nil {
	//	log.Fatal().Err(err).Msgf("Error creating kube config (kubeconfig=%s)", kubeConfigFile)
	//}
	//kubeClient := kubernetes.NewForConfigOrDie(kubeConfig)
	//
	//if err = helpers.LoadProgs(config.KernelTracing); err != nil {
	//	log.Fatal().Msgf("failed to load ebpf programs: %v", err)
	//}
	//
	//if err = rlimit.RemoveMemlock(); err != nil {
	//	log.Fatal().Msgf("remove memlock error: %v", err)
	//}
	//
	//stop := make(chan struct{}, 1)
	//if err = podwatcher.Run(kubeClient, stop); err != nil {
	//	log.Fatal().Err(err)
	//}
	//log.Info().Msgf("Stopping ecnet-bridge %s; %s; %s", version.Version, version.GitCommit, version.BuildDate)
}

func parseFlags() error {
	if err := flags.Parse(os.Args); err != nil {
		return err
	}
	_ = flag.CommandLine.Parse([]string{})
	return nil
}

// getECNETBridgePod returns the ecnet-dataplane pod.
// The pod name is inferred from the 'BRIDGE_POD_NAME' env variable which is set during deployment.
func getECNETBridgePod(kubeClient kubernetes.Interface) (*corev1.Pod, error) {
	podName := os.Getenv("BRIDGE_POD_NAME")
	if podName == "" {
		return nil, fmt.Errorf("BRIDGE_POD_NAME env variable cannot be empty")
	}

	pod, err := kubeClient.CoreV1().Pods(ecnetNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		// TODO(#3962): metric might not be scraped before process restart resulting from this error
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingBridgePod)).
			Msgf("Error retrieving ecnet-dataplane pod %s", podName)
		return nil, err
	}

	return pod, nil
}
