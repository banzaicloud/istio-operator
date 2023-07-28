# Single mesh multi cluster active-active on different networks with Istio operator
This guide will walk through the process of configuring an active-active multi-cluster Istio mesh discribed in the official [Istio documentation](https://istio.io/latest/docs/setup/install/multicluster/multi-primary_multi-network/), but without manual steps to set up connection and trust between the clusters. While this guide discuss a two cluster setup, multiple remote clusters can be added the same way. The cluster roles in the official Istio documentation `primary, remote` are called `active, passive` here.
## Setup:

### Install Cluster Registry:
For further information, chek out the [GitHub repo](https://github.com/cisco-open/cluster-registry-controller#quickstart).
#### Active-1 setup:
Install cluster registry controller, the `controller.apiServerEndpointAddress` value should be set to the public API endpoint address of the cluster (see [Cluster registry docs](https://github.com/cisco-open/cluster-registry-controller/tree/master/deploy/charts/cluster-registry#installing-the-chart)). This is needed, because with certain Kubernetes distributions, the default value can be a private IP address.:
```
API_SERVER=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
helm repo add cluster-registry https://cisco-open.github.io/cluster-registry-controller
helm install --namespace=cluster-registry --create-namespace cluster-registry cluster-registry/cluster-registry --set localCluster.name=demo-active-1 --set network.name=network1 --set controller.apiServerEndpointAddress=$API_SERVER
```
#### Active-2 setup:
Install cluster registry controller, the `controller.apiServerEndpointAddress` value should be set to the public API endpoint address of the cluster (see [Cluster registry docs](https://github.com/cisco-open/cluster-registry-controller/tree/master/deploy/charts/cluster-registry#installing-the-chart)). This is needed, because with certain Kubernetes distributions, the default value can be a private IP address.:
```
API_SERVER=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')
helm repo add cluster-registry https://cisco-open.github.io/cluster-registry-controller
helm install --namespace=cluster-registry --create-namespace cluster-registry cluster-registry/cluster-registry --set localCluster.name=demo-active-2 --set network.name=network2 --set controller.apiServerEndpointAddress=$API_SERVER
```
#### Set up registry connection:
On Active-1 cluster, save cluster resources at `/tmp/demo-active-1.yaml`
```
kubectl get -n=cluster-registry cluster demo-active-1 -o yaml > /tmp/demo-active-1.yaml
```
On Active-2 cluster, save cluster resources at `/tmp/demo-active-2.yaml`
```
kubectl get -n=cluster-registry cluster demo-active-2 -o yaml > /tmp/demo-active-2.yaml
```
On Active-1 cluster, set up connection to Active-2
```
kubectl apply -f /tmp/demo-active-2.yaml 
```
On Active-2 cluster, set up connection to Active-1
```
kubectl apply -f /tmp/demo-active-1.yaml 
```
### Install Istio Operator:
#### Active-1 setup:
1. Install istio operator in the `istio-system` namespace:
```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm install --namespace=istio-system --create-namespace istio-operator-v117x banzaicloud-stable/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true
```
2. Apply ACTIVE-1 `IstioControlPlane` Custom Resource to the `istio-system` namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-active/active-icp-1.yaml
```
#### Active-2 setup:
1. Install istio operator in the `istio-system` namespace:
```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm install --namespace=istio-system --create-namespace istio-operator-v117x banzaicloud-stable/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true
```
2. Apply ACTIVE-2 `IstioControlPlane` Custom Resource to the istio-system namespace:
```
kubectl -n=istio-system apply -f docs/multi-cluster-mesh/active-active/active-icp-2.yaml
```

### Install distributed bookinfo:
#### Active-1:
Label the `default` namespace with the name and namespace of the Istio control plane. This will enable sidecar injection for the later deployed demo application. Deploy the demo application:
```
kubectl label ns default istio.io/rev=icp-v117x.istio-system
kubectl apply -f docs/multi-cluster-mesh/active-active/demoapp-1.yaml
```
#### Active-2:
The namespace labels are synchronized between the clusters by the Istio Operator, so only the demo application needs to be deployed.
```
kubectl apply -f docs/multi-cluster-mesh/active-active/demoapp-2.yaml
```

#### Test connection using the ingress gateway on the ACTIVE-1 cluster:
```
INGRESS_HOST=$(kubectl -n default get service demo-imgw -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
open http://$INGRESS_HOST/productpage
```
*Note: on AWS clusters, run: 
```
INGRESS_HOST=$(kubectl -n default get service demo-imgw -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
open http://$INGRESS_HOST/productpage
``` 

(the first two numbers should be divided by two, that’s how many requests went to reviews-v1 and reviews-v2 received, reviews v3 received the rest out of the 100)
```
for i in `seq 1 100`; do curl -s "http://${INGRESS_HOST}/productpage" |grep -i -e "</html>" -e color=\"; done | sort | uniq -c
```

Traffic split:
```
kubectl apply -f docs/multi-cluster-mesh/active-active/demoapp-vs-dr.yaml
```
