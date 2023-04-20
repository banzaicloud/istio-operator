# Ambient topology with the istio-operator
This guide will walk through the process of configuring an ambient topology
with this repo's istio-operator similar to what is described in the upstream 
documentation [Istio documentation](https://preliminary.istio.io/latest/docs/ops/ambient/getting-started/)
Most of the steps in this guide just reference the upstream guide.

## Setup:

### Create a multinode cluster 
### Build and Install the Istio Operator:
1. Pull the branch with teh ambient changes
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
### Install the bookinfo application:
Follow the same steps as the upstream guide.
#### Label the namespace
Follow the same steps as the upstream guide.
#### Test connection using the ingress gateway on the ACTIVE-1 cluster:
Follow the same steps as the upstream guide.

## Debugging 
Ambient replace the envoy proxy with a purpose built Rust proxy so debugging is quite different.  Here are some commands and links:
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
With ambient it is not uncommon to think everything is working because communication is happening entirely within Kubernetes networking with no traffic being 
redirected to the ztunnels pods or handled by Istio at all.   

## Issues
There seem to be some race conditions as traffic is not always blocked
appropriately by the L4 authorization policy.  This commonly
occurs on the initial install.  Deleting the bookinfo application
and re-installing fixes the issue. 
