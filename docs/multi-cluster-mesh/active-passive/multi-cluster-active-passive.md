# Single mesh multi cluster multiple networks with Istio operator
This guide will walk through the process of configuring an active-passive multi-cluster Istio mesh discribed in the official [Istio documentation](https://istio.io/latest/docs/setup/install/multicluster/primary-remote_multi-network/), but without manual steps to set up connection and trust between the clusters. While this guide discuss a two cluster setup, multiple remote clusters can be added the same way.

## Objective:
Set up two clusters (one active – one passive) in a single mesh multi network (each in different network) setup with Istio operator.

## Setup: 

### Install Cluster Registry and Istio Operator:

#### Active setup:
1. Install cluster registry controller in the `istio-system` namespace:
```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-active --set network.name=network1
```
2. Install istio operator in the `istio-system` namespace:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
3. Apply ACTIVE `IstioControlPlane` Custom Resource to the `istio-system` namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-passive/active-icp.yaml
```

#### Passive setup:
- Install cluster registry controller (replace Kubernetes API server endpoint):
```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-passive --set network.name=network2 --set controller.apiServerEndpointAddress=<KUBERNETES_API_SERVER_FOR_ACTIVE>
```
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get secret,cluster demo-active -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get secret,cluster demo-passive -o yaml | pbcopy     pbpaste | kubectl apply -f -
```
- Install istio operator in the `istio-system` namespace:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set image.repository=lac21/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
- Apply PASSIVE `IstioControlPlane` Custom Resource to the istio-system namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-passive/passive-icp.yaml
```

### Install distributed bookinfo:
#### Active:
Label the `default` namespace with the name and namespace of the Istio control plane. This will enable sidecar injection for the later deployed demo application. Deploy the demo application
```
kubectl label ns default istio.io/rev=icp-v113x.istio-system
kubectl apply -f docs/multi-cluster-mesh/active-passive/demoapp-1.yaml 
```
#### Passive:
The namespace labels are synchronized between the clusters by the Cluster Registry Controller, so only the demo application needs to be deployed.
```
kubectl apply -f docs/multi-cluster-mesh/active-passive/demoapp-2.yaml
```

#### Test connection using the ingress gateway on the ACTIVE cluster:
```
INGRESS_HOST=$(kubectl -n default get service demo-imgw -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
open http://$INGRESS_HOST/productpage
```
(the first two numbers should be divided by two, that’s how many requests went to reviews-v1 and reviews-v2 received, reviews v3 received the rest out of the 100)
```
for i in `seq 1 100`; do curl -s "http://${INGRESS_HOST}/productpage" |grep -i -e "</html>" -e color=\"; done | sort | uniq –c
```

Traffic split:
```
kubectl apply -f docs/multi-cluster-mesh/active-passive/demoapp-vs-dr.yaml 
```