# Istio upgrade guide

The steps are listed in this doc to perform an Istio version upgrade with the operator.

## Istio Control Plane Upgrade

Let us suppose that we have a [Kubernetes](https://kubernetes.io/) cluster with Istio 1.3.4, and we would like to upgrade our Istio components to Istio version 1.4.0. Here are the steps we need to perform to accomplish this with the operator:

1. Deploy a version of the operator which supports Istio 1.4.x
2. Apply a [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) using Istio 1.4.0 components

What happens is that once the operator discerns that the Custom Resource it's watching has changed, it reconciles all Istio-related components in order to perform a control plane upgrade.

> The upgrade process described above is between major Istio versions. Upgrading between minor Istio versions is even easier. You only need to perform the second of those two steps, i.e. apply a Custom Resource with the desired Istio version set for Istio's components, and the operator will do the rest for you.

### Try it out

#### Requirements

- Minikube v1.1.1+ or Kubernetes 1.13.0+
- `KUBECONFIG` set to an existing Kubernetes cluster

If you already have Istio 1.3.x installed on your cluster you can skip the next section and can jump right to [Deploy sample BookInfo application](#deploy-sample-bookinfo-application).

#### Install Istio 1.3.4

We install Istio with our operator, so first we need to check out the `release-1.3` branch of our operator (this branch supports Istio versions 1.3.x):

```bash
$ git clone git@github.com:banzaicloud/istio-operator.git
$ git checkout release-1.3
```

**Install Istio Operator with make**

Simply run the following `make` goal from the project root to install the operator:

```bash
$ make deploy
```

This command will install a Custom Resource Definition in the cluster, and will deploy the operator to the `istio-system` namespace.
As is typical of operators, this will allow you to specify your Istio configurations to a Kubernetes Custom Resource.

**Install Istio Operator with Helm**

Alternatively, if you just can't let go of Helm completely, you can deploy the operator using a [Helm chart](https://github.com/banzaicloud/banzai-charts/tree/master/istio-operator), which is available in the Banzai Cloud stable Helm repo:

```bash
$ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
$ helm install --name=istio-operator --namespace=istio-system --set-string operator.image.tag=0.3.5 banzaicloud-stable/istio-operator
```

*Note: As of now, the `0.3.5` tag is the latest version of our operator to support Istio versions 1.3.x

**Apply the Custom Resource**

Once you've applied the Custom Resource to your cluster, the operator will start reconciling all of Istio's components.

There are some sample Custom Resource configurations in the `config/samples` folder. To deploy Istio 1.3.4 with its default configuration options, use the following command:

```bash
$ kubectl apply -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

After some time, you should see that the Istio pods are running:

```bash
$ kubectl get pods -n istio-system --watch
NAME                                      READY     STATUS    RESTARTS   AGE
istio-citadel-5b69bd4749-h24xk            1/1       Running   0          1m
istio-egressgateway-bb8b48cf8-w65hm       1/1       Running   0          1m
istio-galley-5ddd798686-jpdlm             1/1       Running   0          1m
istio-ingressgateway-678ff4cc87-gkdzt     1/1       Running   0          1m
istio-operator-controller-manager-0       2/2       Running   0          9m
istio-pilot-fc664fcbd-kgl2k               2/2       Running   0          1m
istio-policy-5cf55ff648-xfhf7             2/2       Running   0          1m
istio-sidecar-injector-596f8dddbb-gvzk9   1/1       Running   0          1m
istio-telemetry-7cbf75c5cf-wk4v8          2/2       Running   0          1m
```

The `Istio` Custom Resource is showing `Available` in its status field and the Istio components are using `1.3.4` images :

```bash
$ kubectl describe istio -n istio-system istio -o yaml | grep "image:"
    image: docker.io/istio/citadel:1.3.4
    image: docker.io/istio/galley:1.3.4
    image: docker.io/istio-mixer:1.3.4
    image: docker.io/istio-pilot:1.3.4
    image: docker.io/istio/proxyv2:1.3.4
    image: docker.io/istio/sidecar_injector:1.3.4
```

#### Deploy sample BookInfo application

Let's make sure that Istio 1.3.4 is properly installed with Istio's BookInfo application:

```bash
$ kubectl -n default apply -f https://raw.githubusercontent.com/istio/istio/release-1.3/samples/bookinfo/platform/kube/bookinfo.yaml
service "details" created
deployment.extensions "details-v1" created
service "ratings" created
deployment.extensions "ratings-v1" created
service "reviews" created
deployment.extensions "reviews-v1" created
deployment.extensions "reviews-v2" created
deployment.extensions "reviews-v3" created
service "productpage" created
deployment.extensions "productpage-v1" created

$ kubectl -n default apply -f https://raw.githubusercontent.com/istio/istio/release-1.3/samples/bookinfo/networking/bookinfo-gateway.yaml
gateway.networking.istio.io "bookinfo-gateway" created
virtualservice.networking.istio.io "bookinfo" created
```

Determine the external hostname of the ingress gateway and open productpage in a browser:

```bash
$ INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
$ open http://$INGRESS_HOST/productpage
```

#### Install Istio 1.4.0

To install Istio 1.4.0, first we need to check out the `release-1.4` branch of our operator (this branch supports the Istio 1.4.x versions):

```bash
$ git clone git@github.com:banzaicloud/istio-operator.git
$ git checkout release-1.4
```

> If you installed Istio operator with `make` in the previous section go to to `Install Istio Operator with make`, if you installed it with `helm` go to `Install Istio Operator with helm`. If you haven't installed Istio operator so far you can choose whichever install option you like.

**Install Istio Operator with make**

Simply run the following `make` goal from the project root to install the operator:

```bash
$ make deploy
```

This command will install a Custom Resource Definition in the cluster, and will deploy the operator to the `istio-system` namespace.

**Install Istio Operator with Helm**

Alternatively, you can deploy the operator using a [Helm chart](https://github.com/banzaicloud/banzai-charts/tree/master/istio-operator), which is available in the Banzai Cloud stable Helm repo:

```bash
$ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
$ helm upgrade istio-operator --install --namespace=istio-system --set-string operator.image.tag=0.4.1 banzaicloud-stable/istio-operator
```

*Note: As of now, the `0.4.1` tag is the latest version of our operator to support Istio versions 1.4.x*

**Apply the new Custom Resource**

> If you've installed Istio 1.3.4 or earlier with the Istio operator, and if you check the logs of the operator pod at this point, you will see the following error message: `intended Istio version is unsupported by this version of the operator`. We need to update the Istio Custom Resource with Istio 1.4's components for the operator to be reconciled with the Istio control plane.

To deploy Istio 1.4.0 with its default configuration options, use the following command:

```bash
$ kubectl apply -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

After some time, you should see that new Istio pods are running:

```bash
$ kubectl get pods -n istio-system --watch
NAME                                      READY     STATUS    RESTARTS   AGE
istio-citadel-7664c58768-l8zgb            1/1       Running   0          7m
istio-egressgateway-8588c7c8d-wkpgk       1/1       Running   0          7m
istio-galley-78b8467b4d-b5dqs             1/1       Running   0          7m
istio-ingressgateway-5c48b96cb4-lnfsn     1/1       Running   0          7m
istio-operator-controller-manager-0       2/2       Running   0          16m
istio-pilot-84588fff4c-4lhq8              2/2       Running   0          7m
istio-policy-75f84689f5-78dxr             2/2       Running   0          7m
istio-sidecar-injector-66cd99d8c8-bp4j7   1/1       Running   0          7m
istio-telemetry-7b667c5fbb-2lfdc          2/2       Running   0          7m
```

The `Istio` Custom Resource is showing `Available` in its status field, and the Istio components are now using `1.4.0` images:

```bash
$ kubectl describe istio -n istio-system istio | grep Image:
    Image:                         docker.io/istio/citadel:1.4.0
    Image:          docker.io/istio/galley:1.4.0
    Image:    docker.io/istio/node-agent-k8s:1.4.0
        Image:    docker.io/istio/node-agent-k8s:1.4.0
    Image:          coredns/coredns:1.6.2
    Plugin Image:   docker.io/istio/coredns-plugin:0.2-istio-1.1
    Image:          docker.io/istio/mixer:1.4.0
    Image:    docker.io/istio/node-agent-k8s:1.4.0
    Image:          docker.io/istio/pilot:1.4.0
    Image:             docker.io/istio/proxyv2:1.4.0
    Image:  docker.io/istio/proxyv2:1.4.0
    Image:                          docker.io/istio/sidecar_injector:1.4.0
      Image:                 docker.io/istio/install-cni:1.4.0

```

At this point, your Istio control plane is upgraded to Istio 1.4.0 and your BookInfo application should still be available at:
```bash
$ open http://$INGRESS_HOST/productpage
```

## Istio Data Plane Upgrade

**1. Sidecar upgrades**

In order to change sidecars running older versions of the Istio proxy we need to perform a few manual steps (see [sidecar-upgrade](https://istio.io/docs/setup/kubernetes/upgrade/steps/#sidecar-upgrade)).
We're already working on providing a seamless upgrade path through the operator for sidecars.

**2. In-place upgrades**

An in-place upgrade is impossible right now due to traffic disruptions.
It is our plan to support this use-case as well, in the future.
