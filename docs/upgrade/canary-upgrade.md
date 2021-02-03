# Istio canary upgrade guide

## High level flow

Let us suppose that we have an Istio 1.6 control plane running in our cluster and we would like to upgrade to an Istio 1.7 control plane.
To make sure that we are limiting the potential blast radius of the upgrade, we introduced the Canary upgrade flow. This consists of the following steps:

1. Deploy an Istio 1.7 control plane *next* to the Istio 1.6 control plane
1. Migrate the data plane *gradually* to the new control plane
1. When migration is fully finished, uninstall the Istio 1.6 control plane

![Istio control plane canary upgrade](istio-cp-canary-upgrade.gif)

This canary upgrade model is much safer than the in-place upgrade workflow, mainly because it can be done gradually, at your own pace.
If issues arise they will only affect a small portion of your applications, that you have selected as the initial scope for the upgrade. If that does happen, it will be much easier to rollback only a part of your application for use with the original control plane.

This upgrade flow gives us more flexibility and confidence when making Istio upgrades, dramatically reducing the potential for service disruptions.

### Canary upgrade process with the Istio operator

Let's take a closer look at how the process works in conjunction with our open source Istio operator.
So our starting point is a running Istio operator and Istio 1.6 control plane, with applications using that control plane version.

1. Deploy another Istio operator (alongside the previous one), which has Istio 1.7 support
1. Apply a new Istio CR with Istio version 1.7 and let the operator reconcile an Istio 1.7 control plane

   It is recommended that you turn off gateways for the new control plane and migrate the 1.6 gateway pods to the 1.7 control plane. That way the ingress gateway's typically LoadBalancer-type service will not be recreated, neither will the cloud platform-managed LoadBalancer, and hence the IP address won't change either.

   > This new control plane is often referred to as a *canary*, but, in practice, it is advisable that this be named based on Istio version, as it will remain on the cluster for the long term.

1. Migrate gateways to use the new Istio 1.7 control plane

   This has been intentionally left as a manual step during the new control plane installation with the Istio operator, so that users can make sure that they're ready to use the control plane for their gateways, and only migrate after.

1. Migrate data plane applications

   It is recommended that for safety reasons you perform this migration gradually.
  During the data plane migration, it is important to keep in mind that the two Istio control planes share trust (because they use the same certificates for encrypted communication). Therefore, when an application pod that still uses Istio's 1.6 control plane calls another pod that already uses Istio's 1.7 control plane, the encrypted communication will succeed because of that shared trust.
   That's why this migration can be performed on a namespace by namespace basis and the **communication** of the pods **won't be affected.**

1. Delete the Istio 1.6 control plane

   Once the migration is finished, and you've made sure that your applications are working properly in conjunction with the new control plane, the older Istio 1.6 control plane can be safely deleted.
   It's recommended that you take some time to make sure that everything is working on the new control plane before turning off the old one. The overhead of doing so is minimal as it's only an `istiod` deployment running.

   > In the traditional sense, a canary upgrade flow ends with a rolling update of the old application into the new one.
   > That's not what's happening here.
   > Instead, the *canary* control plane will remain in the cluster for the long term and the original control plane won't be rolled out, instead it is deleted (this is why it's recommended that you name the new control plane based on a version number, rather than naming it *canary*).

## Try it out!

First, we'll deploy an Istio 1.6 control plane with the Istio operator and two demo applications in separate namespaces.
Then, we'll deploy an Istio 1.7 control plane alongside the other control plane and migrate the demo applications to the new control plane gradually.
During the process we'll make sure that the communication works even when the demo apps are on different control planes at the time.
When the data plane migration is finished, we'll delete the Istio 1.6 control plane and operator.

### Setup

1. Create a Kubernetes cluster.

   > If you need a hand with this, you can use our free version of [Banzai Cloud's Pipeline platform](https://try.pipeline.banzai.cloud/) to create a cluster.

### Deploy Istio 1.6 control plane

1. Deploy an Istio operator version, which can install an Istio 1.6 control plane.

   ```bash
   $ helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
   $ helm install istio-operator-v16x --create-namespace --namespace=istio-system --set-string operator.image.tag=0.6.12 --set-string istioVersion=1.6 banzaicloud-stable/istio-operator
   ```

1. Apply an `Istio` Custom Resource and let the operator reconcile the Istio 1.6 control plane.

   ```bash
   $ kubectl apply -n istio-system -f https://raw.githubusercontent.com/banzaicloud/istio-operator/release-1.6/config/samples/istio_v1beta1_istio.yaml
   ```

1. Check that the Istio 1.6 control plane is deployed.

   ```bash
   $ kubectl get po -n=istio-system
   NAME                                    READY   STATUS    RESTARTS   AGE
   istio-ingressgateway-55b89d99d7-4m884   1/1     Running   0          17s
   istio-operator-v16x-0                   2/2     Running   0          57s
   istiod-5865cb6547-zp5zh                 1/1     Running   0          29s
   ```

### Deploy demo app

1. Create two namespaces for the demo apps.

   ```bash
   $ kubectl create ns demo-a
   $ kubectl create ns demo-b
   ```

1. Add those namespaces to the mesh.

   ```bash
   $ kubectl patch istio -n istio-system istio-sample --type=json -p='[{"op": "replace", "path": "/spec/autoInjectionNamespaces", "value": ["demo-a", "demo-b"]}]'
   ```

1. Make sure that the namespaces are labeled for sidecar injection (if not, wait a few seconds, then please re-check the namespaces).

   ```bash
   $ kubectl get ns demo-a demo-b -L istio-injection
   NAME     STATUS   AGE     ISTIO-INJECTION
   demo-a   Active   2m11s   enabled
   demo-b   Active   2m9s    enabled
   ```

1. Deploy two sample applications in those two namespaces.

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: app-a
     labels:
       k8s-app: app-a
     namespace: demo-a
   spec:
     replicas: 1
     selector:
       matchLabels:
         k8s-app: app-a
     template:
       metadata:
         labels:
           k8s-app: app-a
       spec:
         terminationGracePeriodSeconds: 2
         containers:
         - name: echo-service
           image: k8s.gcr.io/echoserver:1.10
           ports:
           - containerPort: 8080
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: app-a
     labels:
       k8s-app: app-a
     namespace: demo-a
   spec:
     ports:
     - name: http
       port: 80
       targetPort: 8080
     selector:
       k8s-app: app-a
   ```

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: app-b
     labels:
       k8s-app: app-b
     namespace: demo-b
   spec:
     replicas: 1
     selector:
       matchLabels:
         k8s-app: app-b
     template:
       metadata:
         labels:
           k8s-app: app-b
       spec:
         terminationGracePeriodSeconds: 2
         containers:
         - name: echo-service
           image: k8s.gcr.io/echoserver:1.10
           ports:
           - containerPort: 8080
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: app-b
     labels:
       k8s-app: app-b
     namespace: demo-b
   spec:
     ports:
     - name: http
       port: 80
       targetPort: 8080
     selector:
       k8s-app: app-b
   ```

### Test communication

1. Save the application pod names for easier access.

   ```bash
   $ APP_A_POD_NAME=$(kubectl get pods -n demo-a -l k8s-app=app-a -o=jsonpath='{.items[0].metadata.name}')
   $ APP_B_POD_NAME=$(kubectl get pods -n demo-b -l k8s-app=app-b -o=jsonpath='{.items[0].metadata.name}')
   ```

1. Test if app-a can access app-b.

   ```bash
   $ kubectl exec -n=demo-a -ti -c echo-service $APP_A_POD_NAME -- curl -Ls -o /dev/null -w "%{http_code}" app-b.demo-b.svc.cluster.local
   200
   ```

1. Test if app-b can access app-a.

   ```bash
   $ kubectl exec -n=demo-b -ti -c echo-service $APP_B_POD_NAME -- curl -Ls -o /dev/null -w "%{http_code}" app-a.demo-a.svc.cluster.local
   200
   ```

### Deploy Istio 1.7 control plane

1. Deploy an Istio operator version, which can install an Istio 1.7 control plane.

   ```bash
   $ helm install istio-operator-v17x --create-namespace --namespace=istio-system --set-string operator.image.tag=0.7.12 banzaicloud-stable/istio-operator
   ```

   *Note: In case you upgrade from an earlier chart version your Istio operator CRD definitions might be outdated in which case you should apply the [new CRDs](../../deploy/charts/istio-operator/crds) manually!*

1. Apply an `Istio` Custom Resource and let the operator reconcile the Istio 1.7 control plane.

   ```bash
   $ kubectl apply -n istio-system -f https://raw.githubusercontent.com/banzaicloud/istio-operator/release-1.7/config/samples/istio_v1beta1_istio.yaml
   ```

1. Check that the Istio 1.7 control plane is also deployed.

   ```bash
   $ kubectl get po -n=istio-system
   NAME                                        READY   STATUS    RESTARTS   AGE
   istio-ingressgateway-55b89d99d7-4m884       1/1     Running   0          6m38s
   istio-operator-v16x-0                       2/2     Running   0          7m18s
   istio-operator-v17x-0                       2/2     Running   0          76s
   istiod-676fc6d449-9jwfj                     1/1     Running   0          10s
   istiod-istio-sample-v17x-7dbdf4f9fc-bfxhl   1/1     Running   0          18s
   ```

### Migrate data plane

#### Migrate ingress

1. Change the ingress gateway so that it utilizes the new Istio 1.7 control plane.

   ```bash
   $ kubectl patch mgw -n istio-system istio-ingressgateway --type=json -p='[{"op": "replace", "path": "/spec/istioControlPlane/name", "value": "istio-sample-v17x"}]'
   ```

#### Migrate first namespace

1. Label the first namespace so that all workloads there utilize the new control plane.

   ```bash
   $ kubectl label ns demo-a istio-injection- istio.io/rev=istio-sample-v17x.istio-system
   ```

   The new `istio.io/rev` label needs to be used for the new revisioned control planes to indicate that it should perform sidecar injection.
   The `istio-injection` label must be removed because it takes precedence over the `istio.io/rev` label for backward compatibility.

1. Make sure that the labeling is correct.

   ```bash
   $ kubectl get ns demo-a -L istio-injection -L istio.io/rev
   NAME     STATUS   AGE   ISTIO-INJECTION   REV
   demo-a   Active   12m                     istio-sample-v17x.istio-system
   ```

1. Restart the pod in the namespace.

   ```bash
   $ kubectl rollout restart deployment -n demo-a
   ```

1. Make sure that the new 1.7 sidecar proxy is used for the new pod.

   ```bash
   $ APP_A_POD_NAME=$(kubectl get pods -n demo-a -l k8s-app=app-a -o=jsonpath='{.items[0].metadata.name}')
   $ kubectl get po -n=demo-a $APP_A_POD_NAME -o yaml | grep istio/proxyv2:
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
   ```

#### Test communication

Let's make sure that encrypted communication still works between pods that use different control planes.
The reason why the data plane migration can be done gradually is that the communication works even between pods on different control planes. In this step it will be verified that the communication works in such a way, making the upgrade flow a safe one.
Remember that the pod(s) in the `demo-a` namespace are already on the Istio 1.7 control plane, but the pod in `demo-b` is still using the Istio 1.6 version.

1. Save the application pod names for easier access.

   ```bash
   $ APP_A_POD_NAME=$(kubectl get pods -n demo-a -l k8s-app=app-a -o=jsonpath='{.items[0].metadata.name}')
   $ APP_B_POD_NAME=$(kubectl get pods -n demo-b -l k8s-app=app-b -o=jsonpath='{.items[0].metadata.name}')
   ```

1. Test if app-a can access app-b.

   ```bash
   $ kubectl exec -n=demo-a -ti -c echo-service $APP_A_POD_NAME -- curl -Ls -o /dev/null -w "%{http_code}" app-b.demo-b.svc.cluster.local
   200
   ```

1. Test if app-b can access app-a.

   ```bash
   $ kubectl exec -n=demo-b -ti -c echo-service $APP_B_POD_NAME -- curl -Ls -o /dev/null -w "%{http_code}" app-a.demo-a.svc.cluster.local
   200
   ```

#### Migrate second namespace

1. Label the second namespace so that all workloads there utilize the new control plane.

   ```bash
   $ kubectl label ns demo-b istio-injection- istio.io/rev=istio-sample-v17x.istio-system
   ```

1. Make sure that the labeling is correct.

   ```bash
   $ kubectl get ns demo-b -L istio-injection -L istio.io/rev
   NAME     STATUS   AGE   ISTIO-INJECTION   REV
   demo-b   Active   19m                     istio-sample-v17x.istio-system
   ```

1. Restart the pod in the namespace.

   ```bash
   $ kubectl rollout restart deployment -n demo-b
   ```

1. Make sure that now the new 1.7 sidecar proxy is used for the new pod.

   ```bash
   $ APP_B_POD_NAME=$(kubectl get pods -n demo-b -l k8s-app=app-b -o=jsonpath='{.items[0].metadata.name}')
   $ kubectl get po -n=demo-b $APP_B_POD_NAME -o yaml | grep istio/proxyv2:
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
       image: docker.io/istio/proxyv2:1.7.7
   ```

### Uninstall the Istio 1.6 control plane

When the data plane is fully migrated to the 1.7 version and you made sure that it works as expected, we can delete the "old" Istio 1.6 control plane.

1. Delete the 1.6 `Istio` Custom Resource to delete the Istio 1.6 control plane.

   ```bash
   $ kubectl delete -n istio-system -f https://raw.githubusercontent.com/banzaicloud/istio-operator/release-1.6/config/samples/istio_v1beta1_istio.yaml
   ```

1. Uninstall the Istio operator for version 1.6.

   > :warning: If you installed the Istio operator chart for Istio 1.6 with Helm2, then first upgrade that chart with Helm version 3.1.0, otherwise your CRDs will be deleted!
   > ```bash
   > $ helm upgrade istio-operator-v16x --namespace=istio-system --set-string operator.image.tag=0.6.13 --set-string istioVersion=1.6 banzaicloud-stable/istio-operator
   > ```
   > For more info on CRD handling issues in Helm, check out these two PRs and the related issues: [https://github.com/helm/helm/pull/7320](https://github.com/helm/helm/pull/7320), [https://github.com/helm/helm/pull/7571](https://github.com/helm/helm/pull/7571).

   ```bash
   $ helm uninstall -n=istio-system istio-operator-v16x
   ```

# Canary upgrade with [Backyards](https://banzaicloud.com/products/backyards/)

If you are looking for an automated and assisted canary upgrade experience based on the Istio operator, check out [Backyards](https://banzaicloud.com/products/backyards/), Banzai Cloud's Istio distribution.

![Istio control plane canary upgrade with Backyards](istio-cp-canary-upgrade-with-backyards.gif)
