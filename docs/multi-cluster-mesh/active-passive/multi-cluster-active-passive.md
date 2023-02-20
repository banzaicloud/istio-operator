# Single mesh multi cluster active-passive on different networks with Istio operator
This guide will walk through the process of configuring an active-passive multi-cluster Istio mesh discribed in the official [Istio documentation](https://istio.io/latest/docs/setup/install/multicluster/primary-remote_multi-network/), but without manual steps to set up connection and trust between the clusters. While this guide discuss a two cluster setup, multiple remote clusters can be added the same way. The cluster roles in the official Istio documentation `primary, remote` are called `active, passive` here.
## Setup:

### Install Cluster Registry:
For further information, chek out the [GitHub repo](https://github.com/cisco-open/cluster-registry-controller#quickstart).
#### Active setup:
Install cluster registry controller, the `controller.apiServerEndpointAddress` value should be set to the public API endpoint address of the cluster (see [Cluster registry docs](https://github.com/cisco-open/cluster-registry-controller/tree/master/deploy/charts/cluster-registry#installing-the-chart)). This is needed, because with certain Kubernetes distributions, the default value can be a private IP address.:
```
API_SERVER=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
helm repo add cluster-registry https://cisco-open.github.io/cluster-registry-controller
helm install --namespace=cluster-registry --create-namespace cluster-registry cluster-registry/cluster-registry --set localCluster.name=demo-active --set network.name=network1 --set controller.apiServerEndpointAddress=$API_SERVER
```
#### Passive setup:
Install cluster registry controller, the `controller.apiServerEndpointAddress` value should be set to the public API endpoint address of the cluster (see [Cluster registry docs](https://github.com/cisco-open/cluster-registry-controller/tree/master/deploy/charts/cluster-registry#installing-the-chart)). This is needed, because with certain Kubernetes distributions, the default value can be a private IP address.:
```
API_SERVER=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
helm repo add cluster-registry https://cisco-open.github.io/cluster-registry-controller
helm install --namespace=cluster-registry --create-namespace cluster-registry cluster-registry/cluster-registry --set localCluster.name=demo-passive --set network.name=network2 --set controller.apiServerEndpointAddress=$API_SERVER
```
#### Set up registry connection:
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get -n=cluster-registry secret,cluster demo-active -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get -n=cluster-registry secret,cluster demo-passive -o yaml | pbcopy     pbpaste | kubectl apply -f -
```
### Install Istio Operator:
#### Active setup:
1. Install istio operator in the `istio-system` namespace:
```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm install --namespace=istio-system --create-namespace istio-operator-v117x banzaicloud-stable/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true
```
2. Apply ACTIVE `IstioControlPlane` Custom Resource to the `istio-system` namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-passive/active-icp.yaml
```
#### Passive setup:
1. Install istio operator in the `istio-system` namespace:
```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm install --namespace=istio-system --create-namespace istio-operator-v117x banzaicloud-stable/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true
```
2. Apply PASSIVE `IstioControlPlane` Custom Resource to the istio-system namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-passive/passive-icp.yaml
```

### Install distributed bookinfo:
#### Active:
Label the `default` namespace with the name and namespace of the Istio control plane. This will enable sidecar injection for the later deployed demo application. Deploy the demo application:
```
kubectl label ns default istio.io/rev=icp-v117x.istio-system
kubectl apply -f docs/multi-cluster-mesh/active-passive/demoapp-1.yaml
```
#### Passive:
The namespace labels are synchronized between the clusters by the Istio Operator, so only the demo application needs to be deployed.
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
