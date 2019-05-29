# Multi cluster scenarios

Istio supports the following multi-cluster patterns:

- **single mesh** – which combines multiple clusters into one unit managed by one Istio control plane
- **multi mesh** – in which they act as individual management domains and the service exposure between those domains is handled separately, controlled by one Istio control plane for each domain

## Single mesh multi-cluster

The single mesh scenario is most suited to those use cases wherein clusters are configured together, sharing resources and typically treated as a single infrastructural component within an organization. A single mesh multi-cluster is formed by enabling any number of Kubernetes control planes running a remote Istio configuration to connect to a single Istio control plane. Once one or more Kubernetes clusters are connected to the Istio control plane in that way, Envoy communicates with the Istio control plane in order to form a mesh network across those clusters.

A multi cluster - single mesh setup has the advantage of all its services looking the same to clients, regardless of where the workloads are actually running; a service named `foo` in namespace `baz` of `cluster1` is the same service as the `foo` in `baz` of `cluster2`. It’s transparent to the application whether it’s been deployed in a single or multi-cluster mesh.

### Single mesh multi-cluster with flat network or VPN

The [Istio operator](https://github.com/banzaicloud/istio-operator) supports setting up single mesh, multi-cluster meshes. This setup has a few network constraints, since all pod CIDRs, as well as API server communications, need to be unique and routable to each other in every cluster.

You can read more about this scenario [here](flat/README.md).

>It's fairly straightforward to set up such an environment on-premise or Google Cloud (which allows the creation of flat networks)

### Single mesh multi-cluster **without** flat network or VPN

The [Istio operator](https://github.com/banzaicloud/istio-operator) supports such a setup as well, using some of the features originally introduced in Istio v1.1: **Split Horizon EDS** and **SNI-based routing**. By using these features, the network constraints for this setup are not untenably steep, since communication passes through the clusters' ingress gateways.

You can read more about this [here](gateway/README.md).

## Multi mesh multi-cluster

In a multi-mesh multi-cluster multiple service meshes are treated as independent fault domains, but with inter-mesh communication.

You can read more about this [here](multimesh/README.md).
