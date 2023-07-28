# Installing Istio-operator on OpenShift
Istio-operator supports OpenShift clusters with full functionality. There are some permissions that are needed for certain Istio components to function.
## Enable OpenShift specific permissions
Allow Istio CP components to run as UID 1337

`oc adm policy add-scc-to-group anyuid system:serviceaccounts:istio-system`

Allow Istio CNI components to run as privileged containers. This is needed to set iptables rules on nodes, to allow istio to function.

`oc adm policy add-scc-to-group privileged system:serviceaccounts:istio-system`

Allow Istio sidecar proxies to run as UID 1337 in the demoapp namespace. This step is needed for any namespaces where sidecar injection is enabled.

`oc adm policy add-scc-to-group anyuid system:serviceaccounts:demoapp`

## Deploy Istio Control Plane
`kubectl apply -n istio-system -f docs/openshift/icp-openshift.yaml`

## Deploy Demo app and Istio Gateway
```
kubectl create ns demoapp
kubectl label namespace demoapp istio.io/rev=icp-v117x.istio-system
kubectl apply -n demoapp -f docs/openshift/gw.yaml
kubectl apply -n demoapp -f docs/openshift/nad.yaml
kubectl -n demoapp apply -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml
```


