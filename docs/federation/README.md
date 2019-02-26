# Istio Multi Cluster Federation

Multi-cluster federation functions by enabling Kubernetes control planes running a remote configuration to connect to one Istio control plane. Once one or more remote Kubernetes clusters are connected to the Istio control plane, Envoy can then communicate with the single Istio control plane and form a mesh network across multiple Kubernetes clusters.

The main requirements for multi-cluster federation to work are that all pod CIDRs must be unique and routable to each other in every cluster and also the API servers must be routable to each other.

## tl;dr

The operator takes care of deploying Istio components to the remote clusters and also provides a constant sync mechanism to provide reachability of Istio's central components from remote clusters.

## How the operator manages federated Istios

- the first step is to create a `kubeconfig` for the remote cluster and add that as a secret to the central cluster where the operator is running.

- create a `RemoteConfig` custom resource which contains the configuration for the operator to be able to deploy the Istio components to the remote k8s cluster and add that cluster into Istio federation.

The major caveat of managing a remote Istio is that it needs a constant connection to some of the components of the central Istio control plane. As we have direct pod reachability between the clusters that sounds an easy thing to do, but keeping the pod IP addresses up-to-date is something that must be solved. The operator triggers an update of those IP address for every remote cluster upon any failure or pod restart of those central components.

You can find a [detailed example](example/README.md) about how to setup a 2 member federated Istio multi-cluster environment on GKE.

> PS. - [Pipeline](http://beta.banzaicloud.io) can setup and automate the whole shebang for you!
