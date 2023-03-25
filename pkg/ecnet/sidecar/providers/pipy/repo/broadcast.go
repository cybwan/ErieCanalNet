// Package repo implements broadcast's methods.
package repo

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"

	"github.com/flomesh-io/ErieCanal/pkg/ecnet/announcements"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/constants"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/identity"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/sidecar/providers/pipy"
	"github.com/flomesh-io/ErieCanal/pkg/ecnet/sidecar/providers/pipy/registry"
)

// Routine which fulfills listening to proxy broadcasts
func (s *Server) broadcastListener() {
	// Register for proxy config updates broadcast by the message broker
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	// Wait for two informer synchronization periods
	slidingTimer := time.NewTimer(time.Second * 20)
	defer slidingTimer.Stop()

	slidingTimerReset := func() {
		slidingTimer.Reset(time.Second * 5)
	}

	s.retryProxiesJob = slidingTimerReset
	s.proxyRegistry.UpdateProxies = slidingTimerReset

	reconfirm := true

	for {
		select {
		case <-proxyUpdateChan:
			// Wait for an informer synchronization period
			slidingTimer.Reset(time.Second * 5)
			// Avoid data omission
			reconfirm = true

		case <-slidingTimer.C:
			connectedProxies := make(map[string]*pipy.Proxy)
			disconnectedProxies := make(map[string]*pipy.Proxy)
			proxies := s.fireExistProxies()
			for _, proxy := range proxies {
				if proxy.PodMetadata == nil {
					if err := s.recordPodMetadata(proxy); err != nil {
						slidingTimer.Reset(time.Second * 5)
						continue
					}
				}
				if proxy.PodMetadata == nil || proxy.Addr == nil || len(proxy.GetAddr()) == 0 {
					slidingTimer.Reset(time.Second * 5)
					continue
				}
				connectedProxies[proxy.UUID.String()] = proxy
			}

			s.proxyRegistry.RangeConnectedProxy(func(key, value interface{}) bool {
				proxyUUID := key.(string)
				if _, exists := connectedProxies[proxyUUID]; !exists {
					disconnectedProxies[proxyUUID] = value.(*pipy.Proxy)
				}
				return true
			})

			if len(connectedProxies) > 0 {
				for _, proxy := range connectedProxies {
					newJob := func() *PipyConfGeneratorJob {
						return &PipyConfGeneratorJob{
							proxy:      proxy,
							repoServer: s,
							done:       make(chan struct{}),
						}
					}
					<-s.workQueues.AddJob(newJob())
				}
			}

			if reconfirm {
				reconfirm = false
				slidingTimer.Reset(time.Second * 10)
			}

			go func() {
				if len(disconnectedProxies) > 0 {
					for _, proxy := range disconnectedProxies {
						s.proxyRegistry.UnregisterProxy(proxy)
					}
				}
			}()
		}
	}
}

func (s *Server) fireExistProxies() []*pipy.Proxy {
	var allProxies []*pipy.Proxy
	allPods := s.kubeController.ListPods()
	for _, pod := range allPods {
		proxy, err := GetProxyFromPod(pod)
		if err != nil {
			continue
		}
		proxy = s.fireUpdatedPod(s.proxyRegistry, proxy)
		allProxies = append(allProxies, proxy)
	}
	return allProxies
}

func (s *Server) fireUpdatedPod(proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy) *pipy.Proxy {
	connectedProxy := proxyRegistry.GetConnectedProxy(proxy.UUID.String())
	if connectedProxy == nil {
		proxyPtr := &proxy
		callback := func(storedProxyPtr **pipy.Proxy) {
			proxyPtr = storedProxyPtr
		}
		s.informProxy(proxyPtr, callback)
		return *proxyPtr
	}
	return connectedProxy
}

func (s *Server) informProxy(proxyPtr **pipy.Proxy, callback func(**pipy.Proxy)) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if aggregatedErr := s.informTrafficPolicies(proxyPtr, &wg, callback); aggregatedErr != nil {
			log.Error().Err(aggregatedErr).Msgf("Pipy Aggregated Traffic Policies Error.")
		}
	}()
	wg.Wait()
}

// GetProxyFromPod infers and creates a Proxy data structure from a Pod.
// This is a temporary workaround as proxy is required and expected in any vertical call,
// however snapshotcache has no need to provide visibility on proxies whatsoever.
// All verticals use the proxy structure to infer the pod later, so the actual only mandatory
// data for the verticals to be functional is the common name, which links proxy <-> pod
func GetProxyFromPod(pod *v1.Pod) (*pipy.Proxy, error) {
	app, appFound := pod.Labels["app"]
	if !appFound {
		return nil, fmt.Errorf("app label not found for pod %s/%s, not a mesh pod", pod.Namespace, pod.Name)
	}
	if app != constants.ECNETControllerName {
		return nil, fmt.Errorf("not mcs bridger controller pod")
	}
	uuidString := pod.UID
	proxyUUID, err := uuid.Parse(string(uuidString))
	if err != nil {
		return nil, fmt.Errorf("Could not parse UUID label into UUID type (%s): %w", uuidString, err)
	}

	sa := pod.Spec.ServiceAccountName
	namespace := pod.Namespace

	return pipy.NewProxy(proxyUUID, identity.New(sa, namespace), nil), nil
}

// GetProxyUUIDFromPod infers and creates a Proxy UUID from a Pod.
func GetProxyUUIDFromPod(pod *v1.Pod) (string, error) {
	uuidString := pod.UID
	proxyUUID, err := uuid.Parse(string(uuidString))
	if err != nil {
		return "", fmt.Errorf("Could not parse UUID label into UUID type (%s): %w", uuidString, err)
	}
	return proxyUUID.String(), nil
}
