# Istio in-place upgrade guide

The steps are listed in this doc to perform an Istio in-place version upgrade with the operator.

## Istio Control Plane Upgrade

Let us suppose that we have a [Kubernetes](https://kubernetes.io/) cluster with Istio 1.8.2, and we would like to upgrade our Istio components to Istio version 1.9.2. Here are the steps we need to perform to accomplish this with the operator:

1. Deploy a version of the operator which supports Istio 1.9.x
2. Apply a [Custom Resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) using Istio 1.9.2 components

What happens is that once the operator discerns that the Custom Resource it's watching has changed, it reconciles all Istio-related components in order to perform a control plane upgrade.

> The upgrade process described above is between major Istio versions. Upgrading between minor Istio versions is even easier. You only need to perform the second of those two steps, i.e. apply a Custom Resource with the desired Istio version set for Istio's components, and the operator will do the rest for you.

### Try it out

#### Requirements

- Minikube v1.1.1+ or Kubernetes 1.17.0+
- `KUBECONFIG` set to an existing Kubernetes cluster

If you already have Istio 1.8.x installed on your cluster you can skip the next section and can jump right to [Deploy sample BookInfo application](#deploy-sample-bookinfo-application).

#### Install Istio 1.8.2

We install Istio with our operator, so first we need to check out the `release-1.8` branch of our operator (this branch supports Istio versions 1.8.x):

```bash
$ git clone git@github.com:banzaicloud/istio-operator.git
$ git checkout release-1.8
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
$ helm install istio-operator --create-namespace --namespace=istio-system --set-string operator.image.tag=0.8.6 --set-string istioVersion=1.8 banzaicloud-stable/istio-operator
```

*Note: As of now, the `0.8.6` tag is the latest version of our operator to support Istio versions 1.8.x

**Apply the Custom Resource**

Once you've applied the Custom Resource to your cluster, the operator will start reconciling all of Istio's components.

There are some sample Custom Resource configurations in the `config/samples` folder. To deploy Istio 1.8.2 with its default configuration options, use the following command:

```bash
$ kubectl apply -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

After some time, you should see that the Istio pods are running:

```bash
$ kubectl get pods -n istio-system --watch
NAME                                      READY     STATUS    RESTARTS   AGE
istio-ingressgateway-678ff4cc87-gkdzt     1/1       Running   0          1m
istio-operator-controller-manager-0       2/2       Running   0          9m
istiod-fc664fcbd-kgl2k                    2/2       Running   0          1m
```

The `Istio` Custom Resource is showing `Available` in its status field and the Istio components are using `1.8.2` images :

```bash
$ kubectl get -n istio-system istio -o yaml | grep "image:"
    image: docker.io/istio/citadel:1.8.2
    image: docker.io/istio/galley:1.8.2
    image: docker.io/istio-mixer:1.8.2
    image: docker.io/istio-pilot:1.8.2
    image: docker.io/istio/proxyv2:1.8.2
    image: docker.io/istio/sidecar_injector:1.8.2
```

#### Deploy sample BookInfo application

Let's make sure that Istio 1.8.2 is properly installed with Istio's BookInfo application:

```bash
$ kubectl -n default apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/platform/kube/bookinfo.yaml
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

$ kubectl -n default apply -f https://raw.githubusercontent.com/istio/istio/1.8.2/samples/bookinfo/networking/bookinfo-gateway.yaml
gateway.networking.istio.io "bookinfo-gateway" created
virtualservice.networking.istio.io "bookinfo" created
```

Determine the external hostname of the ingress gateway and open productpage in a browser:

```bash
$ INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
$ open http://$INGRESS_HOST/productpage
```

#### Install Istio 1.9.2

To install Istio 1.9.2, first we need to check out the `release-1.9` branch of our operator (this branch supports the Istio 1.9.x versions):

```bash
$ git clone git@github.com:banzaicloud/istio-operator.git
$ git checkout release-1.9
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

> :warning: If you installed the original chart with Helm2, then please make sure to use Helm version 3.1.0 to issue the following upgrade command, otherwise your CRDs will be deleted!
> For more info on CRD handling issues in Helm, check out these two PRs and the related issues: [https://github.com/helm/helm/pull/7320](https://github.com/helm/helm/pull/7320), [https://github.com/helm/helm/pull/7571](https://github.com/helm/helm/pull/7571).

```bash
$ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
$ helm upgrade istio-operator --install --namespace=istio-system --set-string operator.image.tag=0.9.1 --set-string istioVersion=1.9.2 banzaicloud-stable/istio-operator
```

*Note: As of now, the `0.9.1` tag is the latest version of our operator to support Istio versions 1.9.x*

*Note: In case you upgrade from an earlier chart version your Istio operator CRD definitions might be outdated in which case you should apply the [new CRDs](../../deploy/charts/istio-operator/crds) manually!*

**Use the new Custom Resource**

> If you've installed Istio 1.8.2 or earlier with the Istio operator, and if you check the logs of the operator pod at this point, you will see the following error message: `intended Istio version is unsupported by this version of the operator`. We need to update the Istio Custom Resource with Istio 1.9's components for the operator to be reconciled with the Istio control plane.

To deploy Istio 1.9.2 with its default configuration options, use the following command (please modify the sample CR to have the same name as the one on your cluster so that replace command can work):

```bash
$ kubectl replace -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

After some time, you should see that new Istio pods are running:

```bash
$ kubectl get pods -n istio-system --watch
NAME                                      READY     STATUS    RESTARTS   AGE
istio-ingressgateway-5c48b96cb4-lnfsn     1/1       Running   0          7m
istio-operator-controller-manager-0       2/2       Running   0          16m
istiod-84588fff4c-4lhq8                   2/2       Running   0          7m
```

The `Istio` Custom Resource is showing `Available` in its status field, and the Istio components are now using `1.9.2` images:

```bash
$ kubectl describe istio -n istio-system istio | grep Image:
    Image:                         docker.io/istio/citadel:1.9.2
    Image:          docker.io/istio/galley:1.9.2
    Image:    docker.io/istio/node-agent-k8s:1.9.2
        Image:    docker.io/istio/node-agent-k8s:1.9.2
    Image:          coredns/coredns:1.9.2
    Plugin Image:   docker.io/istio/coredns-plugin:0.2-istio-1.1
    Image:          docker.io/istio/mixer:1.9.2
    Image:    docker.io/istio/node-agent-k8s:1.9.2
    Image:          docker.io/istio/pilot:1.9.2
    Image:             docker.io/istio/proxyv2:1.9.2
    Image:  docker.io/istio/proxyv2:1.9.2
    Image:                          docker.io/istio/sidecar_injector:1.9.2
      Image:                 docker.io/istio/install-cni:1.9.2
```

At this point, your Istio control plane is upgraded to Istio 1.9.2 and your BookInfo application should still be available at:
```bash
$ open http://$INGRESS_HOST/productpage
```
