# Istio Multi Cluster Federation Example

## Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/docs/quickstarts)

## Install & Deploy

### Create two GKE clusters with IP Alias feature to support flat networking

```bash
gcloud container clusters create k8s-central --enable-ip-alias --zone europe-west1-b --machine-type n1-standard-2 --num-nodes=1 --preemptible --async
gcloud container clusters create k8s-remote-1 --enable-ip-alias --zone us-central1-a --machine-type n1-standard-2 --num-nodes=1 --preemptible --async
```

Wait for the clusters getting into `RUNNING` state and get the credentials for them

```bash
gcloud container clusters get-credentials k8s-central --zone europe-west1-b
gcloud container clusters get-credentials k8s-remote-1 --zone us-central1-a
CONTEXT_CENTRAL=$(kubectl config get-contexts -o name | grep k8s-central)
CONTEXT_REMOTE=$(kubectl config get-contexts -o name | grep k8s-remote-1)
```

### Create firewall rule to allow direct communication between the two clusters

```bash
function join_by { local IFS="$1"; shift; echo "$*"; }
ALL_CLUSTER_CIDRS=$(gcloud container clusters list --filter 'name=(k8s-central,k8s-remote-1)' --format='value(clusterIpv4Cidr)' | sort | uniq)
ALL_CLUSTER_CIDRS=$(join_by , $(echo "${ALL_CLUSTER_CIDRS}"))
ALL_CLUSTER_NETTAGS=$(gcloud compute instances list --filter 'name ~ k8s-central|k8s-remote-1' --format='value(tags.items.[0])' | sort | uniq)
ALL_CLUSTER_NETTAGS=$(join_by , $(echo "${ALL_CLUSTER_NETTAGS}"))
gcloud compute firewall-rules create istio-multicluster-remote-test \
  --allow=tcp,udp,icmp,esp,ah,sctp \
  --direction=INGRESS \
  --priority=900 \
  --source-ranges="${ALL_CLUSTER_CIDRS}" \
  --target-tags="${ALL_CLUSTER_NETTAGS}" --quiet
```

### Install the operator onto the central cluster and enable auto sidecar injection in the default namespace (the operator will do this for you if it's set in the `spec`)

```bash
kubectl config use-context ${CONTEXT_CENTRAL}
make deploy
kubectl create -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

#### Additional RBAC permissions maybe needed if the deployment of the operator fails

```text
kubectl create clusterrolebinding user-cluster-admin --clusterrole=cluster-admin --user=<gke account email>
```

### Change the context to the remote cluster and create the `istio-system` namespace

```bash
kubectl config use-context ${CONTEXT_REMOTE}
kubectl create namespace istio-system
```

### Create a service account and generate kubeconfig for the operator to be able to deploy resources to the remote cluster

```bash
cat <<EOF | kubectl create -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-operator
  namespace: istio-system
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: istio-operator
  namespace: istio-system
EOF

REMOTE_KUBECONFIG_FILE=$(docs/federation/example/generate-kubeconfig.sh)
```

### The kubeconfig for the remote cluster must be added to the central cluster as a secret

```bash
kubectl config use-context ${CONTEXT_CENTRAL}
kubectl create secret generic remoteistio-sample --from-file ${REMOTE_KUBECONFIG_FILE} -n istio-system
rm -f ${REMOTE_KUBECONFIG_FILE}
```

### The added secret must be labeled for Istio

```bash
kubectl label secret remoteistio-sample istio/multiCluster=true -n istio-system
```

### Create the Istio remote config on the central cluster and label the default namespace for auto sidecar injection on the remote cluster as well

```bash
kubectl create -n istio-system -f config/samples/istio_v1beta1_remoteistio.yaml
```

### Add a simple test echo-service onto both clusters

```bash
kubectl config use-context ${CONTEXT_CENTRAL}
kubectl apply -f docs/federation/example/echo-service.yml

kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
echo-59d4b7c4cb-v29zb   2/2     Running   0          1m

kubectl config use-context ${CONTEXT_REMOTE}
kubectl apply -f docs/federation/example/echo-service.yml

kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
echo-59d4b7c4cb-2mptg   2/2     Running   0          1m
```

### Check the setup by doing some request to the echo service from the central cluster

```bash
kubectl config use-context ${CONTEXT_CENTRAL}
kubectl -n default exec $(kubectl get pods -n default -l k8s-app=echo -o jsonpath={.items..metadata.name}) -c echo-service -ti -- sh -c 'for i in `seq 1 50`; do curl -s echo | grep -i hostname | cut -d " " -f 2; done | sort | uniq -c'
```

`The output should be something like this`

     25 echo-59d4b7c4cb-2mptg
     25 echo-59d4b7c4cb-v29zb

### Do a similar test from the remote side

```bash
kubectl config use-context ${CONTEXT_REMOTE}
kubectl -n default exec $(kubectl get pods -n default -l k8s-app=echo -o jsonpath={.items..metadata.name}) -c echo-service -ti -- sh -c 'for i in `seq 1 50`; do curl -s echo | grep -i hostname | cut -d " " -f 2; done | sort | uniq -c'
```

`The output should be something like this`

     25 echo-59d4b7c4cb-2mptg
     25 echo-59d4b7c4cb-v29zb

## Cleanup

```bash
kubectl config delete-context ${CONTEXT_CENTRAL}
kubectl config delete-cluster ${CONTEXT_CENTRAL}

kubectl config delete-context ${CONTEXT_REMOTE}
kubectl config delete-cluster ${CONTEXT_REMOTE}

gcloud container clusters delete k8s-central --zone europe-west1-b --async
gcloud container clusters delete k8s-remote-1 --zone us-central1-a --async
gcloud compute firewall-rules delete istio-multicluster-remote-test
```
