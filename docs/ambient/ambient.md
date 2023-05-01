# Ambient topology with the istio-operator
This guide will walk through the process of deploying an ambient topology
It follows very closely to what is described in the upstream documentation
[Istio documentation](https://preliminary.istio.io/latest/docs/ops/ambient/getting-started/)
Most of the steps in this guide just reference the upstream guide.

## Setup:

### Create a multi-node cluster

Using Kind with the same configuration as in the upstream guide
has been tested.

### Build and Install the Istio Operator:
1. Pull a branch that includes the ambient changes
2. Build the image
```
make docker-build
```
3. Deploy the Operator
```
make deploy
```
4. Apply ambient `IstioControlPlane` Custom Resource to the `istio-system` namespace:
```
kubectl -n=istio-system apply -f docs/ambient/icp-ambient.yaml
```
This configuration points to both ztunnel proxy and Istio images built
by the Istio project.
### Install the bookinfo application:
Follow the same steps as the upstream guide.
#### Label the namespace
Follow the same steps as the upstream guide.
#### Test the connections through the ambient proxy
Follow the same steps as the upstream guide.  Intra-node, Inter-node
and external client connections through the ambient proxy(s)
can all be tested.

## Debugging
Ambient replaces the envoy proxy with a purpose built Rust proxy.
Understainding the dataflow and debugging is quite different.
Here are some commands and links that you might find helpful:
```
istioctl pc workload ztunnel-8cvt2.istio-system
```
```
kubectl exec -it ztunnel-xn2tl -n istio-system curl localhost:15000/config_dump
```
```
kubectl exec -it ztunnel-xn2tl -n istio-system curl localhost:15020/metrics
```
Link to Istio blog [Ztunnel debug](https://istio.io/latest/blog/2023/rust-based-ztunnel/)

Note:
With ambient it is not uncommon to think everything is working because
communication is happening entirely within Kubernetes networking with
no traffic being redirected to the ztunnels pods or handled by Istio
at all. Ensure to check the ztunnel pods metrics to ensure traffic is
passing through it as expected.

## Issues
There seem to be some race conditions as traffic is not always blocked
appropriately by the L4 authorization policy.  This commonly
occurs on the initial install.  Deleting the bookinfo application
and re-installing fixes the issue.
