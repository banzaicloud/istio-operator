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
kubectl -n=istio-system apply -f https://gist.githubusercontent.com/Laci21/182237cda49e1cbb7df4cee051804673/raw/4cef6d5d8f07e4de4a1f621a57d576af4c418cf7/gistfile1.txt
```

#### Passive-1 setup (different network):
- Install cluster registry controller (replace Kubernetes API server endpoint):
```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-passive-1 --set network.name=network2 --set apiServerEndpointAddress=<KUBERNETES_API_SERVER_FOR_ACTIVE>
```
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get secret demo-active -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get secret demo-passive-1 -o yaml | pbcopy     pbpaste | kubectl apply -f -
kubectl get cluster demo-active -o yaml | pbcopy      pbpaste | kubectl apply -f -
kubectl get cluster demo-passive-1 -o yaml | pbcopy    pbpaste | kubectl apply -f -
```
- Install istio operator:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set image.repository=lac21/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
- Apply PASSIVE-1 ICP CR:
```
kubectl -n=istio-system apply -f https://gist.githubusercontent.com/Laci21/1d0a1dc3aefaa81260e1b4f87876d0d5/raw/cfcf0a92cc0499f2fbfffe19194e45525186b876/gistfile1.txt
```

#### Passive-2 setup (same network):
- Install cluster registry controller:

```
helm install --namespace=cluster-registry --create-namespace cluster-registry-controller deploy/charts/cluster-registry  --set image.tag=v0.1.9 --set localCluster.name=demo-passive-2 --set network.name=network1
```
Copy/paste secret and cluster resources from active->passive and passive->active as well:
```
kubectl get secret demo-passive-1 -o yaml | pbcopy       pbpaste | kubectl apply -f -
kubectl get secret demo-passive-2 -o yaml | pbcopy     pbpaste | kubectl apply -f -
kubectl get cluster demo-passive-1 -o yaml | pbcopy      pbpaste | kubectl apply -f -
kubectl get cluster demo-passive-2 -o yaml | pbcopy    pbpaste | kubectl apply -f -
```
- Install istio operator:
```
helm install --namespace=istio-system --create-namespace istio-operator-v113x deploy/charts/istio-operator --set image.repository=lac21/istio-operator --set clusterRegistry.clusterAPI.enabled=true --set clusterRegistry.resourceSyncRules.enabled=true --set image.tag=v2.13.2
```
- Apply PASSIVE-2 ICP CR:
```
kubectl -n=istio-system apply -f https://gist.githubusercontent.com/Laci21/6d3c03e372c8b578c652222e755e9f36/raw/877bd766a7f75e0e59a0e36faf4a6b1c00f95c84/gistfile1.txt
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
kubectl apply -f https://gist.githubusercontent.com/Laci21/cd867a7fd393608934d794806e6f4174/raw/f53c7cbd164a7803ebd5468547b78ec26d4dd543/single-mesh-multi-cluster-bookinfo-demo-active.yaml 
```
Passive-1:
```
kubectl apply -f https://gist.githubusercontent.com/Laci21/1ec3cdb518151ce8869e6dfd03afe1d5/raw/dac76d4b70b6fbc1e07d479ed21e75de6308904d/single-mesh-multi-cluster-bookinfo-demo-passive-1.yaml
```
Passive-2:
```
kubectl apply -f https://gist.githubusercontent.com/Laci21/a0e59d737dc1ba2d340bbad93839617b/raw/f799f2f5fe53314b0ef70e80cb6f5b9813866e79/single-mesh-multi-cluster-bookinfo-demo-passive-2.yaml
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
kubectl apply -f https://gist.githubusercontent.com/Laci21/84c5331238f1bd993ad24e1e014791ea/raw/ceb207c2d5bf50d648677ff9f20f2c619d77a363/single-mesh-multi-cluster-bookinfo-demo-routing.yaml 
```