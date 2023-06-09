package k8s

import (
	"fmt"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/announcements"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/constants"
	ecnetinformers "github.com/flomesh-io/ErieCanal/pkg/ecnet/k8s/informers"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/messaging"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/service"
)

// NewKubernetesController returns a new kubernetes.Controller which means to provide access to locally-cached k8s resources
func NewKubernetesController(informerCollection *ecnetinformers.InformerCollection, msgBroker *messaging.Broker, selectInformers ...InformerKey) Controller {
	return newClient(informerCollection, msgBroker, selectInformers...)
}

func newClient(informerCollection *ecnetinformers.InformerCollection, msgBroker *messaging.Broker, selectInformers ...InformerKey) *client {
	// Initialize client object
	c := &client{
		informers: informerCollection,
		msgBroker: msgBroker,
	}

	// Initialize informers
	informerInitHandlerMap := map[InformerKey]func(){
		Namespaces:      c.initNamespaceMonitor,
		Services:        c.initServicesMonitor,
		ServiceAccounts: c.initServiceAccountsMonitor,
		Pods:            c.initPodMonitor,
		Endpoints:       c.initEndpointMonitor,
	}

	// If specific informers are not selected to be initialized, initialize all informers
	if len(selectInformers) == 0 {
		selectInformers = []InformerKey{Namespaces, Services, ServiceAccounts, Pods, Endpoints}
	}

	for _, informer := range selectInformers {
		informerInitHandlerMap[informer]()
	}

	return c
}

// Initializes Namespace monitoring
func (c *client) initNamespaceMonitor() {
	// Add event handler to informer
	nsEventTypes := EventTypes{
		Add:    announcements.NamespaceAdded,
		Update: announcements.NamespaceUpdated,
		Delete: announcements.NamespaceDeleted,
	}
	c.informers.AddEventHandler(ecnetinformers.InformerKeyNamespace, GetEventHandlerFuncs(nil, nsEventTypes, c.msgBroker))
}

// Function to filter K8s meta Objects by ECNET's isMonitoredNamespace
func (c *client) shouldObserve(obj interface{}) bool {
	object, ok := obj.(metav1.Object)
	if !ok {
		return false
	}
	return c.IsMonitoredNamespace(object.GetNamespace())
}

// Initializes Service monitoring
func (c *client) initServicesMonitor() {
	svcEventTypes := EventTypes{
		Add:    announcements.ServiceAdded,
		Update: announcements.ServiceUpdated,
		Delete: announcements.ServiceDeleted,
	}
	c.informers.AddEventHandler(ecnetinformers.InformerKeyService, GetEventHandlerFuncs(c.shouldObserve, svcEventTypes, c.msgBroker))
}

// Initializes Service Account monitoring
func (c *client) initServiceAccountsMonitor() {
	svcEventTypes := EventTypes{
		Add:    announcements.ServiceAccountAdded,
		Update: announcements.ServiceAccountUpdated,
		Delete: announcements.ServiceAccountDeleted,
	}
	c.informers.AddEventHandler(ecnetinformers.InformerKeyServiceAccount, GetEventHandlerFuncs(c.shouldObserve, svcEventTypes, c.msgBroker))
}

func (c *client) initPodMonitor() {
	podEventTypes := EventTypes{
		Add:    announcements.PodAdded,
		Update: announcements.PodUpdated,
		Delete: announcements.PodDeleted,
	}
	c.informers.AddEventHandler(ecnetinformers.InformerKeyPod, GetEventHandlerFuncs(c.shouldObserve, podEventTypes, c.msgBroker))
}

func (c *client) initEndpointMonitor() {
	eptEventTypes := EventTypes{
		Add:    announcements.EndpointAdded,
		Update: announcements.EndpointUpdated,
		Delete: announcements.EndpointDeleted,
	}
	c.informers.AddEventHandler(ecnetinformers.InformerKeyEndpoints, GetEventHandlerFuncs(c.shouldObserve, eptEventTypes, c.msgBroker))
}

// IsMonitoredNamespace returns a boolean indicating if the namespace is among the list of monitored namespaces
func (c client) IsMonitoredNamespace(namespace string) bool {
	return c.informers.IsMonitoredNamespace(namespace)
}

// ListMonitoredNamespaces returns all namespaces that the mesh is monitoring.
func (c client) ListMonitoredNamespaces() ([]string, error) {
	var namespaces []string

	for _, ns := range c.informers.List(ecnetinformers.InformerKeyNamespace) {
		namespace, ok := ns.(*corev1.Namespace)
		if !ok {
			log.Error().Err(errListingNamespaces).Msg("Failed to list monitored namespaces")
			continue
		}
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

// GetService retrieves the Kubernetes Services resource for the given MeshService
func (c client) GetService(svc service.MeshService) *corev1.Service {
	// client-go cache uses <namespace>/<name> as key
	svcIf, exists, err := c.informers.GetByKey(ecnetinformers.InformerKeyService, svc.NamespacedKey())
	if exists && err == nil {
		svc := svcIf.(*corev1.Service)
		return svc
	}
	return nil
}

// ListServices returns a list of services that are part of monitored namespaces
func (c client) ListServices() []*corev1.Service {
	var services []*corev1.Service

	for _, serviceInterface := range c.informers.List(ecnetinformers.InformerKeyService) {
		svc := serviceInterface.(*corev1.Service)

		if !c.IsMonitoredNamespace(svc.Namespace) {
			continue
		}
		services = append(services, svc)
	}
	return services
}

// ListServiceAccounts returns a list of service accounts that are part of monitored namespaces
func (c client) ListServiceAccounts() []*corev1.ServiceAccount {
	var serviceAccounts []*corev1.ServiceAccount

	for _, serviceInterface := range c.informers.List(ecnetinformers.InformerKeyServiceAccount) {
		sa := serviceInterface.(*corev1.ServiceAccount)

		if !c.IsMonitoredNamespace(sa.Namespace) {
			continue
		}
		serviceAccounts = append(serviceAccounts, sa)
	}
	return serviceAccounts
}

// GetNamespace returns a Namespace resource if found, nil otherwise.
func (c client) GetNamespace(ns string) *corev1.Namespace {
	nsIf, exists, err := c.informers.GetByKey(ecnetinformers.InformerKeyNamespace, ns)
	if exists && err == nil {
		ns := nsIf.(*corev1.Namespace)
		return ns
	}
	return nil
}

// ListPods returns a list of pods part of the mesh
// Kubecontroller does not currently segment pod notifications, hence it receives notifications
// for all k8s Pods.
func (c client) ListPods() []*corev1.Pod {
	var pods []*corev1.Pod

	for _, podInterface := range c.informers.List(ecnetinformers.InformerKeyPod) {
		pod := podInterface.(*corev1.Pod)
		if !c.IsMonitoredNamespace(pod.Namespace) {
			continue
		}
		pods = append(pods, pod)
	}
	return pods
}

// GetEndpoints returns the endpoint for a given service, otherwise returns nil if not found
// or error if the API errored out.
func (c client) GetEndpoints(svc service.MeshService) (*corev1.Endpoints, error) {
	ep, exists, err := c.informers.GetByKey(ecnetinformers.InformerKeyEndpoints, svc.NamespacedKey())
	if err != nil {
		return nil, err
	}
	if exists {
		return ep.(*corev1.Endpoints), nil
	}
	return nil, nil
}

// ListServiceIdentitiesForService lists ServiceAccounts associated with the given service
func (c client) ListServiceIdentitiesForService(svc service.MeshService) ([]service.K8sServiceAccount, error) {
	var svcAccounts []service.K8sServiceAccount

	k8sSvc := c.GetService(svc)
	if k8sSvc == nil {
		return nil, fmt.Errorf("Error fetching service %q: %s", svc, errServiceNotFound)
	}

	svcAccountsSet := mapset.NewSet()
	pods := c.ListPods()
	for _, pod := range pods {
		svcRawSelector := k8sSvc.Spec.Selector
		selector := labels.Set(svcRawSelector).AsSelector()
		// service has no selectors, we do not need to match against the pod label
		if len(svcRawSelector) == 0 {
			continue
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			podSvcAccount := service.K8sServiceAccount{
				Name:      pod.Spec.ServiceAccountName,
				Namespace: pod.Namespace, // ServiceAccount must belong to the same namespace as the pod
			}
			svcAccountsSet.Add(podSvcAccount)
		}
	}

	for svcAcc := range svcAccountsSet.Iter() {
		svcAccounts = append(svcAccounts, svcAcc.(service.K8sServiceAccount))
	}
	return svcAccounts, nil
}

// IsMetricsEnabled returns true if the pod in the mesh is correctly annotated for prometheus scrapping
func IsMetricsEnabled(pod *corev1.Pod) bool {
	prometheusScrapeAnnotation, ok := pod.Annotations[constants.PrometheusScrapeAnnotation]
	if !ok {
		return false
	}

	isScrapingEnabled, _ := strconv.ParseBool(prometheusScrapeAnnotation)
	return isScrapingEnabled
}

// ServiceToMeshServices translates a k8s service with one or more ports to one or more
// MeshService objects per port.
func ServiceToMeshServices(c Controller, svc corev1.Service) []service.MeshService {
	var meshServices []service.MeshService

	for _, portSpec := range svc.Spec.Ports {
		meshSvc := service.MeshService{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			Port:      uint16(portSpec.Port),
		}

		// attempt to parse protocol from port name
		// Order of Preference is:
		// 1. port.appProtocol field
		// 2. protocol prefixed to port name (e.g. tcp-my-port)
		// 3. default to http
		protocol := constants.ProtocolHTTP
		for _, p := range constants.SupportedProtocolsInMesh {
			if strings.HasPrefix(portSpec.Name, p+"-") {
				protocol = p
				break
			}
		}

		// use port.appProtocol if specified, else use port protocol
		meshSvc.Protocol = pointer.StringDeref(portSpec.AppProtocol, protocol)

		// The endpoints for the kubernetes service carry information that allows
		// us to retrieve the TargetPort for the MeshService.
		endpoints, _ := c.GetEndpoints(meshSvc)
		if endpoints != nil {
			meshSvc.TargetPort = GetTargetPortFromEndpoints(portSpec.Name, *endpoints)
		} else {
			log.Warn().Msgf("k8s service %s/%s does not have endpoints but is being represented as a MeshService", svc.Namespace, svc.Name)
		}

		if !IsHeadlessService(svc) || endpoints == nil {
			meshServices = append(meshServices, meshSvc)
			continue
		}

		// If there's not at least 1 subdomain-ed MeshService added,
		// add the entire headless service
		var added bool
		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				if address.Hostname == "" {
					continue
				}
				meshServices = append(meshServices, service.MeshService{
					Namespace:  svc.Namespace,
					Name:       fmt.Sprintf("%s.%s", address.Hostname, svc.Name),
					Port:       meshSvc.Port,
					TargetPort: meshSvc.TargetPort,
					Protocol:   meshSvc.Protocol,
				})
				added = true
			}
		}

		if !added {
			meshServices = append(meshServices, meshSvc)
		}
	}

	return meshServices
}

// GetTargetPortFromEndpoints returns the endpoint port corresponding to the given endpoint name and endpoints
func GetTargetPortFromEndpoints(endpointName string, endpoints corev1.Endpoints) (endpointPort uint16) {
	// Per https://pkg.go.dev/k8s.io/api/core/v1#ServicePort and
	// https://pkg.go.dev/k8s.io/api/core/v1#EndpointPort, if a service has multiple
	// ports, then ServicePort.Name must match EndpointPort.Name when considering
	// matching endpoints for the service's port. ServicePort.Name and EndpointPort.Name
	// can be unset when the service has a single port exposed, in which case we are
	// guaranteed to have the same port specified in the list of EndpointPort.Subsets.
	//
	// The logic below works as follows:
	// If the service has multiple ports, retrieve the matching endpoint port using
	// the given ServicePort.Name specified by `endpointName`.
	// Otherwise, simply return the only port referenced in EndpointPort.Subsets.
	for _, subset := range endpoints.Subsets {
		for _, port := range subset.Ports {
			if endpointName == "" || len(subset.Ports) == 1 {
				// ServicePort.Name is not passed or a single port exists on the service.
				// Both imply that this service has a single ServicePort and EndpointPort.
				endpointPort = uint16(port.Port)
				return
			}

			// If more than 1 port is specified
			if port.Name == endpointName {
				endpointPort = uint16(port.Port)
				return
			}
		}
	}
	return
}

// GetTargetPortForServicePort returns the TargetPort corresponding to the Port used by clients
// to communicate with it.
func (c client) GetTargetPortForServicePort(namespacedSvc types.NamespacedName, port uint16) (uint16, error) {
	// Lookup the k8s service corresponding to the given service name.
	// The k8s service is necessary to lookup the TargetPort from the Endpoint whose name
	// matches the name of the port on the k8s Service object.
	svcIf, exists, err := c.informers.GetByKey(ecnetinformers.InformerKeyService, namespacedSvc.String())
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("service %s not found in cache", namespacedSvc)
	}

	svc := svcIf.(*corev1.Service)
	var portName string
	for _, portSpec := range svc.Spec.Ports {
		if uint16(portSpec.Port) == port {
			portName = portSpec.Name
			break
		}
	}

	// Lookup the endpoint port (TargetPort) that matches the given service and 'portName'
	ep, exists, err := c.informers.GetByKey(ecnetinformers.InformerKeyEndpoints, namespacedSvc.String())
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("endpoint for service %s not found in cache", namespacedSvc)
	}
	endpoint := ep.(*corev1.Endpoints)

	for _, subset := range endpoint.Subsets {
		for _, portSpec := range subset.Ports {
			if portSpec.Name == portName {
				return uint16(portSpec.Port), nil
			}
		}
	}

	return 0, fmt.Errorf("error finding port name %s for endpoint %s", portName, namespacedSvc)
}
