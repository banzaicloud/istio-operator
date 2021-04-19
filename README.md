# Istio operator v2

Istio operator v2 is a Kubernetes operator to deploy and manage [Istio](https://istio.io/) resources for a Kubernetes cluster.

This is the second/revamped version of the original Banzai Cloud Istio operator.
Istio has evolved a lot over the last releases and while we kept up with that with the original Istio operator, we felt like a complete rewrite was needed to more naturally support the new Istio architecture and all of its new features.

> :warning: Istio operator v2 contains **breaking changes** compared to the earlier Istio operator version.

## Overview

[Istio](https://istio.io/) is an open platform to connect, manage, and secure microservices and it is emerging as the `standard` for building service meshes on Kubernetes.

The goal of the **Istio-operator** is to enable popular service mesh use cases (multi cluster topologies, multiple gateways support etc) by introducing easy to use higher level abstractions.

## Issues, feature requests

Please note that the Istio operator is constantly under development and new releases might introduce breaking changes.
We are striving to keep backward compatibility as much as possible while adding new features at a fast pace.
Issues, new features or bugs are tracked on the projects [GitHub page](https://github.com/banzaicloud/istio-operator/issues) - please feel free to add yours!

## Contributing

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## Istio operator support

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
