# Developer Guide

## How to run Istio-operator in your cluster with your changes

The Istio operator is built on the [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) project.

To build the operator and run tests:

1. Run `make vendor`
2. Run `make`

If you make changes and would like to try your own version, create your own image:

1. `make docker-build IMG={YOUR_USERNAME}/istio-operator:v0.0.1`
2. `make docker-push IMG={YOUR_USERNAME}/istio-operator:v0.0.1`
3. `make deploy IMG={YOUR_USERNAME}/istio-operator:v0.0.1`

Watch the operator's logs with:

`kubectl logs -f -n istio-system istio-operator-controller-manager-0 manager`

Create CR and let the operator set up Istio in your cluster (you can change the `spec` of the `Config` for your needs in the yaml file):

`kubectl create -n istio-system -f config/samples/operator_v1beta1_config.yaml`

You should be able to setup Istio's [Bookinfo Application](https://istio.io/docs/examples/bookinfo/) at this point or start using it as you wish.
