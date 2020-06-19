
<p align="center"><img src="docs/img/istio_operator_logo.png" width="260"></p>

<p align="center">

  <a href="https://hub.docker.com/r/banzaicloud/istio-operator/">
    <img src="https://img.shields.io/docker/cloud/automated/banzaicloud/istio-operator.svg" alt="Docker Automated build">
  </a>

  <a href="https://circleci.com/gh/banzaicloud/istio-operator/tree/master">
    <img src="https://circleci.com/gh/banzaicloud/istio-operator/tree/master.svg?style=shield" alt="CircleCI">
  </a>

  <a href="https://goreportcard.com/report/github.com/banzaicloud/istio-operator">
    <img src="https://goreportcard.com/badge/github.com/banzaicloud/istio-operator" alt="Go Report Card">
  </a>

  <a href="https://github.com/banzaicloud/istio-operator/">
    <img src="https://img.shields.io/badge/license-Apache%20v2-orange.svg" alt="license">
  </a>

</p>

# Istio-operator

Istio-operator is a Kubernetes operator to deploy and manage [Istio](https://istio.io/) resources for a Kubernetes cluster.

## Overview

[Istio](https://istio.io/) is an open platform to connect, manage, and secure microservices and it is emerging as the `standard` for building service meshes on Kubernetes.
It is built out on multiple components and a rather complex deployment scheme (20+ CRDs).
Installing, upgrading and operating these components requires deep understanding of Istio.

The goal of the **Istio-operator** is to automate and simplify these and enable popular service mesh use cases (multi cluster federation, multiple gateways support, resource reconciliation, etc) by introducing easy higher level abstractions.

![Istio Operator](/docs/img/operator.png)

## Istio operator vs [Backyards](https://banzaicloud.com/products/backyards/)

[Backyards](https://banzaicloud.com/products/backyards/) is Banzai Cloud's **production ready Istio distribution**.
The Banzai Cloud Istio operator is a core part of Backyards, which helps with installing, upgrading and managing an Istio mesh, but [Backyards](https://banzaicloud.com/products/backyards/) provides many other components to conveniently secure, operate and observe Istio as well.

The differences are presented in this table:

|                           |   Istio operator   |      Backyards     |
|:-------------------------:|:------------------:|:------------------:|
|       Install Istio       | :heavy_check_mark: | :heavy_check_mark: |
|        Manage Istio       | :heavy_check_mark: | :heavy_check_mark: |
|       Upgrade Istio       | :heavy_check_mark: | :heavy_check_mark: |
|      Uninstall Istio      | :heavy_check_mark: | :heavy_check_mark: |
|   Multi cluster support   | :heavy_check_mark: | :heavy_check_mark: |
| Multiple gateways support | :heavy_check_mark: | :heavy_check_mark: |
|         Prometheus        |                    | :heavy_check_mark: |
|          Grafana          |                    | :heavy_check_mark: |
|           Jaeger          |                    | :heavy_check_mark: |
|        Cert manager       |                    | :heavy_check_mark: |
|         Dashboard         |                    | :heavy_check_mark: |
|            CLI            |                    | :heavy_check_mark: |
|   Enhanced observability  |                    | :heavy_check_mark: |
|       Topology graph      |                    | :heavy_check_mark: |
|      Live access logs     |                    | :heavy_check_mark: |
|      mTLS management      |                    | :heavy_check_mark: |
|     Gateway management    |                    | :heavy_check_mark: |
|     Sidecar management    |                    | :heavy_check_mark: |
|          Routing          |                    | :heavy_check_mark: |
|      Circuit Breaking     |                    | :heavy_check_mark: |
|      Fault Injection      |                    | :heavy_check_mark: |
|         Mirroring         |                    | :heavy_check_mark: |
|      Canary releases      |                    | :heavy_check_mark: |
|         Validations       |                    | :heavy_check_mark: |

For a complete list of [Backyards](https://banzaicloud.com/products/backyards/) features please check out the [features](https://banzaicloud.com/docs/backyards/overview/) page.

## Istio operator installation

The operator (`release-1.6` branch) installs the 1.6.2 version of Istio, and can run on Minikube v1.1.1+ and Kubernetes 1.15.0+.

As a pre-requisite it needs a Kubernetes cluster (you can create one using [Pipeline](https://github.com/banzaicloud/pipeline)).

1. Set `KUBECONFIG` pointing towards your cluster
2. Run `make deploy` (deploys the operator in the `istio-system` namespace to the cluster)
3. Set your Istio configurations in a Kubernetes custom resource (sample: `config/samples/istio_v1beta1_istio.yaml`) and run this command to deploy the Istio components:

```bash
kubectl create -n istio-system -f config/samples/istio_v1beta1_istio.yaml
```

### Installation with [Backyards](https://banzaicloud.com/products/backyards/)

Go grab and install Istio with the [Backyards CLI](https://github.com/banzaicloud/backyards-cli) tool.

```bash
curl https://getbackyards.sh | sh && backyards istio install
```

### Installation with Helm

Alternatively, if you just can’t let go of Helm completely, you can deploy the operator using a Helm chart, which is available in the Banzai Cloud stable [Helm repo](deploy/charts/istio-operator):

```bash
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com/
helm install --name=istio-operator --namespace=istio-system banzaicloud-stable/istio-operator
```

### Installation with Kustomize

You can also have your own `kustomization.yaml` file with a reference to Istio operator as a base without the need to clone the repo. See more info in the [Kustomize usage doc](config/README.md).

```bash
bases:
  - github.com/banzaicloud/istio-operator/config?ref=release-1.6
  - github.com/banzaicloud/istio-operator/config/overlays/auth-proxy-enabled?ref=release-1.6
```

## Istio upgrade

Check out the [upgrade docs](docs/upgrade.md) to see how to upgrade between minor or major Istio versions.

## Multi-cluster federation

Check out the [multi-cluster federation docs](docs/federation/README.md).

## Development

Check out the [developer docs](docs/developer.md).

## Uninstall

To remove Istio and Istio operator completely from your cluster execute the following steps:

1. Delete the Istio configuration custom resource you have created earlier (Istio operator will take care of deleting all Istio resources from your cluster after the custom resource is deleted)
2. Delete the `istio-system` namespace to delete Istio operator itself

```bash
kubectl delete -n istio-system -f config/samples/istio_v1beta1_istio.yaml
kubectl delete namespace istio-system
```

## Issues, feature requests and roadmap

Please note that the Istio operator is constantly under development and new releases might introduce breaking changes.
We are striving to keep backward compatibility as much as possible while adding new features at a fast pace.
Issues, new features or bugs are tracked on the projects [GitHub page](https://github.com/banzaicloud/istio-operator/issues) - please feel free to add yours!

To track some of the significant features and future items from the roadmap, please visit the [roadmap doc](docs/roadmap.md).

## Contributing

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## Community

If you have any questions about the Istio operator, and would like to talk to us and the other members of the Banzai Cloud community, please join our **#istio-operator** channel on [Slack](https://pages.banzaicloud.com/invite-slack).

We also frequently write blog posts about [Istio](https://banzaicloud.com/tags/istio/) itself and the [Istio operator](https://banzaicloud.com/tags/istio-operator/).

## License

Copyright (c) 2017-2020 [Banzai Cloud, Inc.](https://banzaicloud.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
