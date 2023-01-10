Steps:

oc adm policy add-scc-to-group anyuid system:serviceaccounts:istio-system

oc adm policy add-scc-to-group anyuid system:serviceaccounts:demoapp

oc adm policy add-scc-to-group privileged system:serviceaccounts:istio-system

k apply -n istio-system -f docs/openshift/icp-openshift.yaml

k create ns demoapp

k label namespace demoapp istio.io/rev=icp-v115x.istio-system

k apply -n demoapp -f docs/openshift/gw.yaml

k apply -n demoapp -f docs/openshift/nad.yaml

k -n demoapp apply -f https://raw.githubusercontent.com/istio/istio/master/samples/bookinfo/platform/kube/bookinfo.yaml


