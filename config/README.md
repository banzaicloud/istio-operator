# Kustomize usage

Developers can have their own `kustomization.yaml` file with a reference to Istio operator as a base without the need to clone the repo.

Firstly, you'll need to install the necessary crds and namespace with the following base:

```
bases:
  - github.com/banzaicloud/istio-operator/config?ref=release-1.1
```

 Secondly, you can install the operator with multiple possible configurations with the use of overlays (choose one option).

 - `basic`: installs the clusterrole, clusterrolebinding and statefulset for the operator

```
bases:
  - github.com/banzaicloud/istio-operator/config/overlays/basic?ref=release-1.1
```

 - `auth-proxy-enabled`: besides the basic configs, installs the auth proxy resources as well

```
bases:
  - github.com/banzaicloud/istio-operator/config/overlays/auth-proxy-enabled?ref=release-1.1
```

 - `prometheus-scpraping-enabled`: besides the basic configs, enables Prometheus scraping for the manager pod

```
bases:
  - github.com/banzaicloud/istio-operator/config/overlays/prometheus-scpraping-enabled?ref=release-1.1
```

 - `psp`: besides the basic configs, add basic pod security policy for the operator and the Istio component pods

```
bases:
  - github.com/banzaicloud/istio-operator/config/overlays/psp?ref=release-1.1
```
