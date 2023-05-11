# Open Service Mesh Edge Helm Chart

![Version: 1.0.2](https://img.shields.io/badge/Version-1.0.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.2](https://img.shields.io/badge/AppVersion-v1.0.2-informational?style=flat-square)

A Helm chart to install the [ecnet](https://github.com/flomesh-io/ErieCanal) control plane on Kubernetes.

## Prerequisites

- Kubernetes >= 1.19.0-0

## Get Repo Info

```console
helm repo add ecnet https://flomesh-io.github.io/ErieCanal
helm repo update
```

## Install Chart

```console
helm install [RELEASE_NAME] ecnet/ecnet
```

The command deploys `ecnet-controller` on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values ecnet/ecnet
```

The following table lists the configurable parameters of the ecnet chart and their default values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key | string | `"kubernetes.io/os"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator | string | `"In"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0] | string | `"linux"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[1].key | string | `"kubernetes.io/arch"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[1].operator | string | `"In"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[1].values[0] | string | `"amd64"` |  |
| ecnet.cleanup.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[1].values[1] | string | `"arm64"` |  |
| ecnet.cleanup.nodeSelector | object | `{}` |  |
| ecnet.cleanup.tolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.configResyncInterval | string | `"90s"` | Sets the resync interval for regular proxy broadcast updates, set to 0s to not enforce any resync |
| ecnet.controlPlaneTolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.controllerLogLevel | string | `"info"` | Controller log verbosity |
| ecnet.curlImage | string | `"curlimages/curl"` | Curl image for control plane init container |
| ecnet.ecnetBootstrap | object | `{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]},{"key":"kubernetes.io/arch","operator":"In","values":["amd64","arm64"]}]}]}},"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app","operator":"In","values":["ecnet-bootstrap"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}},"nodeSelector":{},"podLabels":{},"replicaCount":1,"resource":{"limits":{"cpu":"0.5","memory":"128M"},"requests":{"cpu":"0.3","memory":"128M"}},"tolerations":[]}` | ECNET bootstrap parameters |
| ecnet.ecnetBootstrap.podLabels | object | `{}` | ECNET bootstrap's pod labels |
| ecnet.ecnetBootstrap.replicaCount | int | `1` | ECNET bootstrap's replica count |
| ecnet.ecnetBootstrap.resource | object | `{"limits":{"cpu":"0.5","memory":"128M"},"requests":{"cpu":"0.3","memory":"128M"}}` | ECNET bootstrap's container resource parameters |
| ecnet.ecnetBootstrap.tolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.ecnetBridge | object | `{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]},{"key":"kubernetes.io/arch","operator":"In","values":["amd64","arm64"]}]}]}},"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app","operator":"In","values":["ecnet-controller"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}},"cni":{"hostCniBridgeEth":"cni0"},"kernelTracing":true,"resource":{"limits":{"cpu":"1.5","memory":"1G"},"requests":{"cpu":"0.5","memory":"256M"}},"tolerations":[]}` | ECNET bridge parameters |
| ecnet.ecnetBridge.tolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.ecnetController | object | `{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]},{"key":"kubernetes.io/arch","operator":"In","values":["amd64","arm64"]}]}]}},"podAntiAffinity":{"preferredDuringSchedulingIgnoredDuringExecution":[{"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app","operator":"In","values":["ecnet-controller"]}]},"topologyKey":"kubernetes.io/hostname"},"weight":100}]}},"autoScale":{"cpu":{"targetAverageUtilization":80},"enable":false,"maxReplicas":5,"memory":{"targetAverageUtilization":80},"minReplicas":1},"podLabels":{},"replicaCount":1,"resource":{"limits":{"cpu":"1.5","memory":"1G"},"requests":{"cpu":"0.5","memory":"128M"}},"tolerations":[]}` | ECNET controller parameters |
| ecnet.ecnetController.autoScale | object | `{"cpu":{"targetAverageUtilization":80},"enable":false,"maxReplicas":5,"memory":{"targetAverageUtilization":80},"minReplicas":1}` | Auto scale configuration |
| ecnet.ecnetController.autoScale.cpu.targetAverageUtilization | int | `80` | Average target CPU utilization (%) |
| ecnet.ecnetController.autoScale.enable | bool | `false` | Enable Autoscale |
| ecnet.ecnetController.autoScale.maxReplicas | int | `5` | Maximum replicas for autoscale |
| ecnet.ecnetController.autoScale.memory.targetAverageUtilization | int | `80` | Average target memory utilization (%) |
| ecnet.ecnetController.autoScale.minReplicas | int | `1` | Minimum replicas for autoscale |
| ecnet.ecnetController.podLabels | object | `{}` | ECNET controller's pod labels |
| ecnet.ecnetController.replicaCount | int | `1` | ECNET controller's replica count (ignored when autoscale.enable is true) |
| ecnet.ecnetController.tolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.ecnetName | string | `"ecnet"` | Identifier for the instance of an ecnet within a cluster |
| ecnet.ecnetNamespace | string | `""` | Namespace to deploy ECNET in. If not specified, the Helm release namespace is used. |
| ecnet.enforceSingleEcnet | bool | `true` | Enforce only deploying one ecnet in the cluster |
| ecnet.image | object | `{"name":{"ecnetBootstrap":"ecnet-bootstrap","ecnetBridge":"ecnet-bridge","ecnetBridgeInit":"ecnet-bridge-init","ecnetCRDs":"ecnet-crds","ecnetController":"ecnet-controller","ecnetPreinstall":"ecnet-preinstall"},"pullPolicy":"IfNotPresent","registry":"flomesh","tag":"1.0.2"}` | ECNET control plane image parameters |
| ecnet.image.name | object | `{"ecnetBootstrap":"ecnet-bootstrap","ecnetBridge":"ecnet-bridge","ecnetBridgeInit":"ecnet-bridge-init","ecnetCRDs":"ecnet-crds","ecnetController":"ecnet-controller","ecnetPreinstall":"ecnet-preinstall"}` | Image name defaults |
| ecnet.image.name.ecnetBootstrap | string | `"ecnet-bootstrap"` | ecnet-bootstrap's image name |
| ecnet.image.name.ecnetBridge | string | `"ecnet-bridge"` | ecnet-bridge's image name |
| ecnet.image.name.ecnetBridgeInit | string | `"ecnet-bridge-init"` | ecnet-bridge-init's image name |
| ecnet.image.name.ecnetCRDs | string | `"ecnet-crds"` | ecnet-crds' image name |
| ecnet.image.name.ecnetController | string | `"ecnet-controller"` | ecnet-controller's image name |
| ecnet.image.name.ecnetPreinstall | string | `"ecnet-preinstall"` | ecnet-preinstall's image name |
| ecnet.image.pullPolicy | string | `"IfNotPresent"` | Container image pull policy for control plane containers |
| ecnet.image.registry | string | `"flomesh"` | Container image registry for control plane images |
| ecnet.image.tag | string | `"1.0.2"` | Container image tag for control plane images |
| ecnet.imagePullSecrets | list | `[]` | `ecnet-controller` image pull secret |
| ecnet.localDNSProxy | object | `{"enable":true}` | Local DNS Proxy improves the performance of your computer by caching the responses coming from your DNS servers |
| ecnet.pluginChains.inbound-http[0].plugin | string | `"modules/inbound-tls-termination"` |  |
| ecnet.pluginChains.inbound-http[0].priority | int | `180` |  |
| ecnet.pluginChains.inbound-http[1].plugin | string | `"modules/inbound-http-routing"` |  |
| ecnet.pluginChains.inbound-http[1].priority | int | `170` |  |
| ecnet.pluginChains.inbound-http[2].plugin | string | `"modules/inbound-metrics-http"` |  |
| ecnet.pluginChains.inbound-http[2].priority | int | `160` |  |
| ecnet.pluginChains.inbound-http[3].plugin | string | `"modules/inbound-tracing-http"` |  |
| ecnet.pluginChains.inbound-http[3].priority | int | `150` |  |
| ecnet.pluginChains.inbound-http[4].plugin | string | `"modules/inbound-logging-http"` |  |
| ecnet.pluginChains.inbound-http[4].priority | int | `140` |  |
| ecnet.pluginChains.inbound-http[5].plugin | string | `"modules/inbound-throttle-service"` |  |
| ecnet.pluginChains.inbound-http[5].priority | int | `130` |  |
| ecnet.pluginChains.inbound-http[6].plugin | string | `"modules/inbound-throttle-route"` |  |
| ecnet.pluginChains.inbound-http[6].priority | int | `120` |  |
| ecnet.pluginChains.inbound-http[7].plugin | string | `"modules/inbound-http-load-balancing"` |  |
| ecnet.pluginChains.inbound-http[7].priority | int | `110` |  |
| ecnet.pluginChains.inbound-http[8].plugin | string | `"modules/inbound-http-default"` |  |
| ecnet.pluginChains.inbound-http[8].priority | int | `100` |  |
| ecnet.pluginChains.inbound-tcp[0].disable | bool | `false` |  |
| ecnet.pluginChains.inbound-tcp[0].plugin | string | `"modules/inbound-tls-termination"` |  |
| ecnet.pluginChains.inbound-tcp[0].priority | int | `130` |  |
| ecnet.pluginChains.inbound-tcp[1].disable | bool | `false` |  |
| ecnet.pluginChains.inbound-tcp[1].plugin | string | `"modules/inbound-tcp-routing"` |  |
| ecnet.pluginChains.inbound-tcp[1].priority | int | `120` |  |
| ecnet.pluginChains.inbound-tcp[2].disable | bool | `false` |  |
| ecnet.pluginChains.inbound-tcp[2].plugin | string | `"modules/inbound-tcp-load-balancing"` |  |
| ecnet.pluginChains.inbound-tcp[2].priority | int | `110` |  |
| ecnet.pluginChains.inbound-tcp[3].disable | bool | `false` |  |
| ecnet.pluginChains.inbound-tcp[3].plugin | string | `"modules/inbound-tcp-default"` |  |
| ecnet.pluginChains.inbound-tcp[3].priority | int | `100` |  |
| ecnet.pluginChains.outbound-http[0].plugin | string | `"modules/outbound-http-routing"` |  |
| ecnet.pluginChains.outbound-http[0].priority | int | `160` |  |
| ecnet.pluginChains.outbound-http[1].plugin | string | `"modules/outbound-metrics-http"` |  |
| ecnet.pluginChains.outbound-http[1].priority | int | `150` |  |
| ecnet.pluginChains.outbound-http[2].plugin | string | `"modules/outbound-tracing-http"` |  |
| ecnet.pluginChains.outbound-http[2].priority | int | `140` |  |
| ecnet.pluginChains.outbound-http[3].plugin | string | `"modules/outbound-logging-http"` |  |
| ecnet.pluginChains.outbound-http[3].priority | int | `130` |  |
| ecnet.pluginChains.outbound-http[4].plugin | string | `"modules/outbound-circuit-breaker"` |  |
| ecnet.pluginChains.outbound-http[4].priority | int | `120` |  |
| ecnet.pluginChains.outbound-http[5].plugin | string | `"modules/outbound-http-load-balancing"` |  |
| ecnet.pluginChains.outbound-http[5].priority | int | `110` |  |
| ecnet.pluginChains.outbound-http[6].plugin | string | `"modules/outbound-http-default"` |  |
| ecnet.pluginChains.outbound-http[6].priority | int | `100` |  |
| ecnet.pluginChains.outbound-tcp[0].plugin | string | `"modules/outbound-tcp-routing"` |  |
| ecnet.pluginChains.outbound-tcp[0].priority | int | `120` |  |
| ecnet.pluginChains.outbound-tcp[1].plugin | string | `"modules/outbound-tcp-load-balancing"` |  |
| ecnet.pluginChains.outbound-tcp[1].priority | int | `110` |  |
| ecnet.pluginChains.outbound-tcp[2].plugin | string | `"modules/outbound-tcp-default"` |  |
| ecnet.pluginChains.outbound-tcp[2].priority | int | `100` |  |
| ecnet.preinstall | object | `{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]},{"key":"kubernetes.io/arch","operator":"In","values":["amd64","arm64"]}]}]}}},"nodeSelector":{},"tolerations":[]}` | ECNET's preinstall hook parameters |
| ecnet.preinstall.tolerations | list | `[]` | Node tolerations applied to control plane pods. The specified tolerations allow pods to schedule onto nodes with matching taints. |
| ecnet.proxyImage | string | `"flomesh/pipy:0.90.1-30"` | Proxy image for Linux node workloads |
| ecnet.proxyLogLevel | string | `"error"` | Log level for the proxy. Non developers should generally never set this value. In production environments the LogLevel should be set to `error` |
| ecnet.proxyServerPort | int | `6060` | Remote destination port on which the Discovery Service listens for new connections from Sidecars. |
| ecnet.repoServer | object | `{"codebase":"","image":"flomesh/pipy-repo:0.90.1-30","ipaddr":"127.0.0.1","standalone":false}` | Pipy RepoServer |
| ecnet.repoServer.codebase | string | `""` | codebase is the folder used by ecnetController. |
| ecnet.repoServer.image | string | `"flomesh/pipy-repo:0.90.1-30"` | Image used for Pipy RepoServer |
| ecnet.repoServer.ipaddr | string | `"127.0.0.1"` | ipaddr of host/service where Pipy RepoServer is installed |
| ecnet.repoServer.standalone | bool | `false` | if false , Pipy RepoServer is installed within ecnetController pod. |
| ecnet.trustDomain | string | `"cluster.local"` | The trust domain to use as part of the common name when requesting new certificates. |

<!-- markdownlint-enable MD013 MD034 -->
<!-- markdownlint-restore -->