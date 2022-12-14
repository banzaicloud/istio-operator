# Istio operator

Istio operator is a Kubernetes operator to deploy and manage [Istio](https://istio.io/) resources for a Kubernetes cluster.

## Overview

[Istio](https://istio.io/) is an open platform to connect, manage, and secure microservices and it is emerging as the `standard` for building service meshes on Kubernetes.

The goal of the **Istio-operator** is to enable popular service mesh use cases (multi cluster topologies, multiple gateways support etc) by introducing easy to use higher level abstractions.

## In this README

- [Getting started](#getting-started)
- [Issues, feature requests](#issues-feature-requests)
- [Contributing](#contributing)
- [Got stuck? Find help!](#got-stuck-find-help)

## Getting started

### Prerequisites
- kubectl installed
- kubernetes cluster (version 1.21+)
- active kubecontext to the kubernetes cluster

###  Build and deploy
Download or check out the latest stable release.

Run `make deploy` to deploy the operator controller-manager on your kubernetes cluster.

Check if the controller is running in the `istio-system` namespace:
```
$ kubectl get pod -n istio-system

NAME                                                READY   STATUS    RESTARTS   AGE
istio-operator-controller-manager-6f764787c-rbnht   2/2     Running   0          5m18s
```

Deploy the [Istio control plane sample](config/samples/servicemesh_v1alpha1_istiocontrolplane.yaml) to the `istio-system` namespace
```
$ kubectl -n istio-system apply -f config/samples/servicemesh_v1alpha1_istiocontrolplane.yaml
istiocontrolplane.servicemesh.cisco.com/icp-v115x-sample created
```

Label the namespace, where you would like to enable sidecar injection for your pods. The label should consist of the name of the deployed IstioControlPlane and the namespace where it is deployed.
```
$ kubectl label namespace demoapp istio.io/rev=icp-v115x-sample.istio-system
namespace/demoapp labeled
```

Deploy the [Istio ingress gateway sample](config/samples/servicemesh_v1alpha1_istiomeshgateway.yaml) to your desired namespace
```
$ kubectl -n demoapp apply -f config/samples/servicemesh_v1alpha1_istiomeshgateway.yaml
istiomeshgateway.servicemesh.cisco.com/imgw-sample created
```

Deploy your application (or the [sample bookinfo app](https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml)).
```
$ kubectl -n demoapp apply -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml
service/details created
serviceaccount/bookinfo-details created
deployment.apps/details-v1 created
service/ratings created
serviceaccount/bookinfo-ratings created
deployment.apps/ratings-v1 created
service/reviews created
serviceaccount/bookinfo-reviews created
deployment.apps/reviews-v1 created
deployment.apps/reviews-v2 created
deployment.apps/reviews-v3 created
service/productpage created
serviceaccount/bookinfo-productpage created
deployment.apps/productpage-v1 created
```

Verify that all applications pods are running and have the sidecar proxy injected. The READY column shows the number of containers for the pod: this should be 1/1 for the gateway, and at least 2/2 for the other pods (the original container of the pods + the sidecar container).
```
$ kubectl get pod -n demoapp
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-79f774bdb9-8xqwj       2/2     Running   0          35s
imgw-sample-66555d5b84-kv62w      1/1     Running   0          7m21s
productpage-v1-6b746f74dc-cx6x6   2/2     Running   0          33s
ratings-v1-b6994bb9-g9vm2         2/2     Running   0          35s
reviews-v1-545db77b95-rdmsp       2/2     Running   0          34s
reviews-v2-7bf8c9648f-rzmvj       2/2     Running   0          34s
reviews-v3-84779c7bbc-t5rfq       2/2     Running   0          33s
```

Deploy the VirtualService and Gateway needed for your application.
**For the [demo bookinfo](https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/networking/bookinfo-gateway.yaml) application, you need to modify the Istio Gateway entry!** The `spec.selector.istio` field should be set from `ingressgateway` to `imgw-sample` so it will be applied to the sample IstioMeshGateway deployed before. The port needs to be set to the targetPort of the deployed IstioMeshGateway.
```
curl https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/networking/bookinfo-gateway.yaml | sed 's/istio: ingressgateway # use istio default controller/istio: imgw-sample/g;s/number: 80/number: 9080/g' | kubectl apply -f -
```
```
$ kubectl -n demoapp apply -f bookinfo-gateway.yaml
gateway.networking.istio.io/bookinfo-gateway created
virtualservice.networking.istio.io/bookinfo created
```

To access your application, use the public IP address of the `imgw-sample` LoadBalancer service.
```
$ IP=$(kubectl -n demoapp get svc imgw-sample -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
$ curl -I $IP/productpage
HTTP/1.1 200 OK
content-type: text/html; charset=utf-8
content-length: 4183
server: istio-envoy
date: Mon, 02 May 2022 14:20:49 GMT
x-envoy-upstream-service-time: 739
```

## Issues, feature requests

Please note that the Istio operator is constantly under development and new releases might introduce breaking changes.
We are striving to keep backward compatibility as much as possible while adding new features at a fast pace.
Issues, new features or bugs are tracked on the projects [GitHub page](https://github.com/banzaicloud/istio-operator/issues) - please feel free to add yours!

## Contributing

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## Got stuck? Find help!

### Community support

If you encounter any problems that is not addressed in our documentation, [open an issue](https://github.com/banzaicloud/istio-operator/issues) or talk to us on the [Banzai Cloud Slack channel #istio-operator.](https://pages.banzaicloud.com/invite-slack).

### Engineering blog

We occasionally write blog posts about [Istio](https://ciscotechblog.com/tags/istio/) itself and the [Istio operator](https://ciscotechblog.com/tags/istio-operator/).

## License

Copyright (c) 2021 Cisco Systems, Inc. and/or its affiliates

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
