# Istio-operator

Istio-operator is a Kubernetes operator to deploy and manage [Istio](https://istio.io/) resources for a Kubernetes cluster.

## Installation

1. Set `KUBECONFIG` for your cluster
2. Run `make deploy`

To build and run tests:
1. Run `make vendor`
2. Run `make`

If you make changes and would like to try your own version create your own image before `make deploy`:
1. `make docker-build IMG={YOUR_USERNAME}/istio-operator:v0.0.1`
2. `make docker-push IMG={YOUR_USERNAME}/istio-operator:v0.0.1`

## Circle

## Contributing

If you find this project useful here's how you can help:

- Send a pull request with your new features and bug fixes
- Help new users with issues they may encounter
- Support the development of this project and star this repo!

## License

Copyright (c) 2017-2019 [Banzai Cloud, Inc.](https://banzaicloud.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
