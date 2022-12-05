# Istio-operator chart

[Istio-operator](https://github.com/banzaicloud/istio-operator/tree/release-1.15) is a Kubernetes operator to deploy and manage [Istio](https://istio.io/) resources for a Kubernetes cluster.

## Prerequisites

- Helm3
- Kubernetes 1.21.0 - 1.24.x

## Installing the chart

To install the chart:

```bash
❯ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
❯ helm install --create-namespace --namespace=istio-system istio-operator banzaicloud-stable/istio-operator
```

## Uninstalling the Chart

To uninstall/delete the `istio-operator` release:

```bash
❯ helm uninstall istio-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Banzaicloud Istio Operator chart and their default values.

| Parameter                                      | Description                                                                                                                                                                                         | Default                                                                                  |
|------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------|
| `image.repository`                             | Operator container image repository                                                                                                                                                                 | `ghcr.io/banzaicloud/istio-operator`                                                     |
| `image.tag`                                    | Operator container image tag                                                                                                                                                                        | `v2.15.3`                                                                           |
| `image.pullPolicy`                             | Operator container image pull policy                                                                                                                                                                | `IfNotPresent`                                                                           |
| `replicaCount`                                 | Operator deployment replica count                                                                                                                                                                   | `1`                                                                                      |
| `extraArgs`                                    | Operator deployment arguments                                                                                                                                                                       | `[]`                                                                                     |
| `resources`                                    | CPU/Memory resource requests/limits (YAML)                                                                                                                                                          | Memory: `256Mi`, CPU: `200m`                                                             |
| `podAnnotations`                               | Operator deployment pod annotations (YAML)                                                                                                                                                          | sidecar.istio.io/inject: `"false"`                                                       |
| `podSecurityContext`                           | Operator deployment pod security context (YAML)                                                                                                                                                     | `fsGroup: 1337`                                                                          |
| `securityContext`                              | Operator deployment security context (YAML)                                                                                                                                                         | runAsUser: `1337`, runAsGroup: `1337`, runAsNonRoot: `true`, capabilities: `drop: - ALL` |
| `nodeselector`                                 | Operator deployment node selector (YAML)                                                                                                                                                            | `{}`                                                                                     |
| `tolerations`                                  | Operator deployment tolerations                                                                                                                                                                     | `[]`                                                                                     |
| `affinity`                                     | Operator deployment affinity (YAML)                                                                                                                                                                 | `{}`                                                                                     |
| `imagePullSecrets`                             | Operator deployment image pull secrets                                                                                                                                                              | `[]`                                                                                     |
| `prometheusMetrics.enabled`                    | If true, use direct access for Prometheus metrics                                                                                                                                                   | `true`                                                                                   |
| `prometheusMetrics.authProxy.enabled`          | If true, use auth proxy for Prometheus metrics                                                                                                                                                      | `true`                                                                                   |
| `prometheusMetrics.authProxy.image.repository` | Auth proxy container image repository                                                                                                                                                               | `gcr.io/kubebuilder/kube-rbac-proxy`                                                     |
| `prometheusMetrics.authProxy.image.tag`        | Auth proxy container image tag                                                                                                                                                                      | `v0.8.0`                                                                                 |
| `prometheusMetrics.authProxy.image.pullPolicy` | Auth proxy container image pull policy                                                                                                                                                              | `IfNotPresent`                                                                           |
| `rbac.enabled`                                 | If true, create rbac service account and roles                                                                                                                                                      | `true`                                                                                   |
| `nameOverride`                                 | Name override for resource names                                                                                                                                                                    | `""`                                                                                     |
| `fullnameOverride`                             | Full name override for resource names                                                                                                                                                               | `""`                                                                                     |
| `useNamespaceResource`                         | If true, create namespace                                                                                                                                                                           | `false`                                                                                  |
| `leaderElection.enabled`                       | If true, leader election is enabled for the operator deployment                                                                                                                                     | `false`                                                                                  |
| `leaderElection.namespace`                     | Namespace for the leader election configmap                                                                                                                                                         | `istio-system`                                                                           |
| `leaderElection.nameOverride`                  | Name override for the leader election configmap                                                                                                                                                     | `""`                                                                                     |
| `apiServerEndpointAddress`                     | Endpoint address of the API server of the cluster the controller is running on                                                                                                                      | `""`                                                                                     |
| `clusterRegistry.clusterAPI.enabled`           | If true, [cluster registry](https://github.com/cisco-open/cluster-registry-controller/api) API is used from the cluster                                                                             | `false`                                                                                  |
| `clusterRegistry.resourceSyncRules.enabled`    | If true, the necessary ResourceSyncRule resources from the [cluster registry](https://github.com/cisco-open/cluster-registry-controller/api) API are automatically created for multi cluster setups | `false`                                                                                  |
