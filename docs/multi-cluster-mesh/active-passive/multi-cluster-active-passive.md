# Single mesh multi cluster multiple networks with Istio operator

## Objective:
Setup three clusters (one active – two passive) in a single mesh multi network (two in the same network, one in a different network) setup with Istio operator.

## Setup: 

### Steps:

#### Active setup:
1. Install cluster registry controller:
```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-active --set network.name=network1
```
2. Install istio operator:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
3. Apply ACTIVE ICP CR:
```
kubectl -n=istio-system apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/active-icp.yaml
```

#### Passive-1 setup (different network):
- Install cluster registry controller (replace Kubernetes API server endpoint):
```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-passive-1 --set network.name=network2 --set controller.apiServerEndpointAddress=<KUBERNETES_API_SERVER_FOR_ACTIVE>
```
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get secret,cluster demo-active -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get secret,cluster demo-passive-1 -o yaml | pbcopy     pbpaste | kubectl apply -f -
```
- Install istio operator:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set image.repository=lac21/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
- Apply PASSIVE-1 ICP CR:
```
kubectl -n=istio-system apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/passive-icp-1.yaml
```

#### Passive-2 setup (same network):
- Install cluster registry controller:

```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-passive-2 --set network.name=network1
```
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get secret,cluster demo-passive-1 -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get secret,cluster demo-passive-2 -o yaml | pbcopy     pbpaste | kubectl apply -f -
```
- Install istio operator:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set image.repository=lac21/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
- Apply PASSIVE-2 ICP CR:
```
kubectl -n=istio-system apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/passive-icp-2.yaml
```

### Install echo demo app:
#### Active:
```
kubectl label ns default istio.io/rev=icp-v113x.istio-system
```
To all clusters:
```
kubectl apply -n=default -f https://raw.githubusercontent.com/banzaicloud/istio-operator/release-1.10/docs/federation/flat/echo-service.yml
```
Active:
```
kubectl run -n=default curl-test --image=radial/busyboxplus:curl -i --tty --rm
for i in `seq 1 99 `; do curl -s echo | grep Hostname; done | sort | uniq –c
```

### Install distributed bookinfo:
#### Active:
```
kubectl label ns default istio.io/rev=icp-v113x.istio-system
kubectl apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/demoapp-1.yaml 
```
Passive-1:
```
kubectl apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/demoapp-2.yaml
```
Passive-2:
```
kubectl apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/demoapp-3.yaml
```

Active:
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
kubectl apply -f https://raw.githubusercontent.com/banzaicloud/istio-operator/9948582d6a0dd35af59739ca286a3e72c8ba810c/docs/multi-cluster-mesh/active-passive/demoapp-vs-dr.yaml 
```