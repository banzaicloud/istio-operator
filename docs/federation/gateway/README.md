## Istio Split Horizon EDS Multi Cluster Federation Example

Two new features which were introduced in Istio v1.1 come in handy to get rid of the necessity of a flat network or VPN between the clusters, Split Horizon EDS and SNI based routing.

> - EDS is short for Endpoint Discovery Service (EDS), a part of Envoy’s API which is used to fetch cluster members called “endpoint” in Envoy terminology
> - SNI based routing leverages the “Server Name Indication” TLS extension to make routing decisions
> - Split-horizon EDS enables Istio to route requests to different endpoints, depending on the location of the request source. Istio Gateways intercept and parse the TLS handshake and use the SNI data to decide on the destination service endpoints.

A single mesh multi-cluster is formed by enabling any number of Kubernetes control planes running a remote Istio configuration to connect to a single Istio control plane. Once one or more Kubernetes clusters is connected to the Istio control plane in that way, Envoy communicates with the Istio control plane in order to form a mesh network across those clusters.

For demo purposes, create 3 clusters, a single node [Banzai Cloud PKE](https://banzaicloud.com/blog/pke-cncf-certified-k8s/) cluster on EC2, a GKE cluster with 2 nodes and an EKS cluster also with 2 nodes.

### Get the latest version of the [Istio operator](https://github.com/banzaicloud/istio-operator)

```bash
❯ git clone https://github.com/banzaicloud/istio-operator.git
❯ cd istio-operator
❯ git checkout release-1.1
```

[Pipeline platform](https://beta.banzaicloud.io/) is the easiest way to setup that environment using our [CLI tool](https://banzaicloud.com/blog/cli-ux/) ([install](https://github.com/banzaicloud/banzai-cli#installation)) for [Pipeline](https:/github.com/banzaicloud/pipeline), simply called `banzai`.

```bash
AWS_SECRET_ID="[[secretID from Pipeline]]"
GKE_SECRET_ID="[[secretID from Pipeline]]"
GKE_PROJECT_ID="[[GKE Project ID]]"

❯ cat docs/federation/gateway/samples/istio-pke-cluster.json | sed "s/{{secretID}}/${AWS_SECRET_ID}/" | banzai cluster create
INFO[0004] cluster is being created
INFO[0004] you can check its status with the command `banzai cluster get "istio-pke"`
Id   Name
541  istio-pke

❯ cat docs/federation/gateway/samples/istio-gke-cluster.json | sed -e "s/{{secretID}}/${GKE_SECRET_ID}/" -e "s/{{projectID}}/${GKE_PROJECT_ID}/" | banzai cluster create
INFO[0004] cluster is being created
INFO[0004] you can check its status with the command `banzai cluster get "istio-gke"`
Id   Name
542  istio-gke

❯ cat docs/federation/gateway/samples/istio-eks-cluster.json | sed "s/{{secretID}}/${AWS_SECRET_ID}/" | banzai cluster create
INFO[0004] cluster is being created
INFO[0004] you can check its status with the command `banzai cluster get "istio-eks"`
Id   Name
543  istio-eks
```

#### Wait for the clusters to be up and running

```bash
❯ banzai cluster list
Id   Name       Distribution  Status    CreatorName  CreatedAt
543  istio-eks  eks           RUNNING   waynz0r      2019-04-14T16:55:46Z
542  istio-gke  gke           RUNNING   waynz0r      2019-04-14T16:54:15Z
541  istio-pke  pke           RUNNING   waynz0r      2019-04-14T16:52:52Z
```

Download the kubeconfigs from [Pipeline UI](https://beta.banzaicloud.io) and set them as k8s contexts.

```bash
❯ export KUBECONFIG=~/Downloads/istio-pke.yaml:~/Downloads/istio-gke.yaml:~/Downloads/istio-eks.yaml

❯ kubectl config get-contexts -o name
istio-eks
istio-gke
kubernetes-admin@istio-pke

❯ export CTX_EKS=istio-eks
❯ export CTX_GKE=istio-gke
❯ export CTX_PKE=kubernetes-admin@istio-pke
```

### Install the operator onto the PKE cluster

```bash
❯ kubectl config use-context ${CTX_PKE}
❯ make deploy
❯ kubectl --context=${CTX_PKE} -n istio-system create -f docs/federation/gateway/samples/istio-multicluster-cr.yaml
```

This command will install a custom resource definition in the cluster, and will deploy the operator in the `istio-system` namespace.
Following a pattern typical of operators, this will allow you to specify your Istio configurations to a Kubernetes custom resource.
Once you apply that to your cluster, the operator will start reconciling all of Istio's components.

Wait for the `gateway-multicluster` Istio resource status become `Available` and also the pods in the `istio-system` become ready as well.

```bash
❯ kubectl --context=${CTX_PKE} -n istio-system get istios
NAME                   STATUS      ERROR   AGE
gateway-multicluster   Available           1m

❯ kubectl --context=${CTX_PKE} -n istio-system get pods
NAME                                      READY   STATUS    RESTARTS   AGE
istio-citadel-67f99b7f5f-lg859            1/1     Running   0          1m30s
istio-galley-665cf4d49d-qrm5s             1/1     Running   0          1m30s
istio-ingressgateway-64f9d4b75b-jh6jr     1/1     Running   0          1m30s
istio-operator-controller-manager-0       2/2     Running   0          2m27s
istio-pilot-df5d467c7-jmj77               2/2     Running   0          1m30s
istio-policy-57dd995b-fq4ss               2/2     Running   2          1m29s
istio-sidecar-injector-746f5cccd9-8mwdb   1/1     Running   0          1m18s
istio-telemetry-6b6b987c94-nmqpg          2/2     Running   2          1m29s
```

Add GKE cluster to the service mesh

```bash
❯ kubectl --context=${CTX_GKE} apply -f docs/federation/gateway/cluster-add
❯ GKE_KUBECONFIG_FILE=$(docs/federation/gateway/cluster-add/generate-kubeconfig.sh ${CTX_GKE})
❯ kubectl --context $CTX_PKE -n istio-system create secret generic istio-gke --from-file=istio-gke=${GKE_KUBECONFIG_FILE}
❯ rm -f ${GKE_KUBECONFIG_FILE}
❯ kubectl --context=${CTX_PKE} create -n istio-system -f docs/federation/gateway/samples/remoteistio-gke-cr.yaml
```

Add EKS cluster to the service mesh

```bash
❯ kubectl --context=${CTX_EKS} apply -f docs/federation/gateway/cluster-add
❯ EKS_KUBECONFIG_FILE=$(docs/federation/gateway/cluster-add/generate-kubeconfig.sh ${CTX_EKS})
❯ kubectl --context $CTX_PKE -n istio-system create secret generic istio-eks --from-file=istio-eks=${EKS_KUBECONFIG_FILE}
❯ rm -f ${EKS_KUBECONFIG_FILE}
❯ kubectl --context=${CTX_PKE} create -n istio-system -f docs/federation/gateway/samples/remoteistio-eks-cr.yaml
```

Wait for the `istio-eks` and `istio-gke` RemoteIstio resource status to become `Available` and also the pods in the `istio-system` on those clusters to become ready as well.

> It could take some time to these resrouces to become `Available` and also reconiliation failures will occur since the reconciliation process must determine the ingress gateway addresses of the clusters.

```text
❯ kubectl --context=${CTX_PKE} -n istio-system get remoteistios
NAME        STATUS      ERROR   GATEWAYS                    AGE
istio-eks   Available           [35.177.214.60 3.8.50.24]   3m
istio-gke   Available           [35.204.1.52]               5m

❯ kubectl --context=${CTX_GKE} -n istio-system get pods
NAME                                      READY   STATUS    RESTARTS   AGE
istio-citadel-75648bdf6b-k8rfl            1/1     Running   0          6m9s
istio-ingressgateway-7f494bcf8-qb692      1/1     Running   0          6m8s
istio-sidecar-injector-746f5cccd9-7b55s   1/1     Running   0          6m8s

❯ kubectl --context=${CTX_EKS} -n istio-system get pods
NAME                                      READY   STATUS    RESTARTS   AGE
istio-citadel-78478cfb44-7h42v            1/1     Running   0          4m
istio-ingressgateway-7f75c479b8-w4qr9     1/1     Running   0          4m
istio-sidecar-injector-56dbb9587f-928h9   1/1     Running   0          4m
```

Deploy the bookinfo sample application in a distributed way

```bash
❯ kubectl --context=${CTX_GKE} apply -n default -f docs/federation/gateway/bookinfo/deployments/productpage-v1.yaml
❯ kubectl --context=${CTX_GKE} apply -n default -f docs/federation/gateway/bookinfo/deployments/reviews-v2.yaml
❯ kubectl --context=${CTX_GKE} apply -n default -f docs/federation/gateway/bookinfo/deployments/reviews-v3.yaml
deployment.extensions/productpage-v1 created
deployment.extensions/reviews-v2 created
deployment.extensions/reviews-v3 created

❯ kubectl --context=${CTX_EKS} apply -n default -f docs/federation/gateway/bookinfo/deployments/details-v1.yaml
❯ kubectl --context=${CTX_EKS} apply -n default -f docs/federation/gateway/bookinfo/deployments/ratings-v1.yaml
❯ kubectl --context=${CTX_EKS} apply -n default -f docs/federation/gateway/bookinfo/deployments/reviews-v1.yaml
deployment.extensions/details-v1 created
deployment.extensions/ratings-v1 created
deployment.extensions/reviews-v1 created

❯ kubectl --context=${CTX_PKE} apply -n default -f docs/federation/gateway/bookinfo/services/
service/details created
service/productpage created
service/ratings created
service/reviews created

❯ kubectl --context=${CTX_EKS} apply -n default -f docs/federation/gateway/bookinfo/services/
service/details created
service/productpage created
service/ratings created
service/reviews created

❯ kubectl --context=${CTX_GKE} apply -n default -f docs/federation/gateway/bookinfo/services/
service/details created
service/productpage created
service/ratings created
service/reviews created
```

Add Istio resources to configure destination rules and virtual services.

```bash
❯ kubectl --context=${CTX_PKE} apply -n default -f docs/federation/gateway/bookinfo/istio/
gateway.networking.istio.io/bookinfo-gateway created
destinationrule.networking.istio.io/productpage created
destinationrule.networking.istio.io/reviews created
destinationrule.networking.istio.io/ratings created
destinationrule.networking.istio.io/details created
virtualservice.networking.istio.io/bookinfo created
virtualservice.networking.istio.io/reviews created
```

### Service mesh in action

The components of the bookinfo app is distributed in the mesh. 3 different versions are deployed from the `reviews` service with the configuration that 1/3 of traffic goes to each.

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: reviews
  namespace: default
spec:
  hosts:
  - reviews
  http:
  - match:
    - headers:
        end-user:
          exact: banzai
    route:
    - destination:
        host: reviews
        subset: v2
  - route:
    - destination:
        host: reviews
        subset: v1
      weight: 33
    - destination:
        host: reviews
        subset: v2
      weight: 33
    - destination:
        host: reviews
        subset: v3
      weight: 34
```

The bookinfo app's product page is reachable through every clusters ingress gateway since they are part of one single mesh. Let's determine the ingress gateway address of the clusters:

```bash
export PKE_INGRESS=$(kubectl --context=${CTX_PKE} -n istio-system get svc/istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
export GKE_INGRESS=$(kubectl --context=${CTX_GKE} -n istio-system get svc/istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export EKS_INGRESS=$(kubectl --context=${CTX_EKS} -n istio-system get svc/istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
```

Hit the PKE cluster's ingress with some traffic to see the spreading of the `reviews` service:

```bash
for i in `seq 1 100`; do curl -s "http://${PKE_INGRESS}/productpage" |grep -i -e "</html>" -e color=\"; done | sort | uniq -c
  62                 <font color="black">
  72                 <font color="red">
 100 </html>
```

It's looks a bit cryptic at first, but that output means that 100 request were successful, 31 hit `reviews-v2`, 36 hit `reviews-v3`, the remaining 33 hit `reviews-v1`. The results could vary, but should be around 1/3 for each service.

Test it through to other clusters ingresses:

```bash
# GKE
for i in `seq 1 100`; do curl -s "http://${GKE_INGRESS}/productpage" |grep -i -e "</html>" -e color=\"; done | sort | uniq -c
  66                 <font color="black">
  72                 <font color="red">
 100 </html>

# EKS
for i in `seq 1 100`; do curl -s "http://${EKS_INGRESS}/productpage" |grep -i -e "</html>" -e color=\"; done | sort | uniq -c
  74                 <font color="black">
  56                 <font color="red">
 100 </html>
```

#### Cleanup

Execute the following commands to clean up the clusters:

```bash
❯ kubectl --context=${CTX_PKE} delete namespace istio-system
❯ kubectl --context=${CTX_GKE} delete namespace istio-system
❯ kubectl --context=${CTX_EKS} delete namespace istio-system

❯ banzai cluster delete istio-pke --no-interactive
❯ banzai cluster delete istio-gke --no-interactive
❯ banzai cluster delete istio-eks --no-interactive
```
