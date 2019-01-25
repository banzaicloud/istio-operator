package istio

import (
	istiov1alpha1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1alpha1"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/goph/emperror"
	"github.com/go-logr/logr"
	"fmt"
	"strings"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileIstio) ReconcileCrds(log logr.Logger, istio *istiov1alpha1.Istio) error {
	crdClient := r.crdClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range crds {
		// TODO: parallelize
		oldCRD, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			controllerutil.SetControllerReference(istio, crd, r.scheme)
			if _, err := crdClient.Create(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created", "crd", crd.Spec.Names.Kind)
		}
		if err == nil {
			crd.ResourceVersion = oldCRD.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated", "crd", crd.Spec.Names.Kind)
		}
	}
	return nil
}

type configType int

const (
	Networking     configType = iota
	Authentication
	Apim
	Policy
	Rbac
)

type crdConfig struct {
	version    string
	group      string
	categories []string
}

var crdConfigs = map[configType]crdConfig{
	Networking: {
		"v1alpha3",
		"networking",
		[]string{"istio-io", "networking-istio-io"},
	},
	Authentication: {
		"v1alpha1",
		"authentication",
		[]string{"istio-io", "authentication-istio-io"},
	},
	Apim: {
		"v1alpha2",
		"config",
		[]string{"istio-io", "apim-istio-io"},
	},
	Policy: {
		"v1alpha2",
		"config",
		[]string{"istio-io", "policy-istio-io"},
	},
	Rbac: {
		"v1alpha1",
		"rbac",
		[]string{"istio-io", "rbac-istio-io"},
	},
}

var crds = []*extensionsobj.CustomResourceDefinition{
	crdL("VirtualService", "VirtualServices", crdConfigs[Networking], "istio-pilot", "", "", true),
	crdL("DestinationRule", "DestinationRules", crdConfigs[Networking], "istio-pilot", "", "", true),
	crdL("ServiceEntry", "ServiceEntries", crdConfigs[Networking], "istio-pilot", "", "", true),
	crd("Gateway", "Gateways", crdConfigs[Networking], "istio-pilot", "", ""),
	crd("EnvoyFilter", "EnvoyFilters", crdConfigs[Networking], "istio-pilot", "", ""),
	crd("Policy", "Policies", crdConfigs[Authentication], "", "", ""),
	crd("MeshPolicy", "MeshPolicies", crdConfigs[Authentication], "", "", ""),
	crd("HTTPAPISpecBinding", "HTTPAPISpecBindings", crdConfigs[Apim], "", "", ""),
	crd("HTTPAPISpec", "HTTPAPISpecs", crdConfigs[Apim], "", "", ""),
	crd("QuotaSpecBinding", "QuotaSpecBindings", crdConfigs[Apim], "", "", ""),
	crd("QuotaSpec", "QuotaSpecs", crdConfigs[Apim], "", "", ""),
	crd("rule", "rules", crdConfigs[Policy], "mixer", "istio.io.mixer", "core"),
	crd("attributemanifest", "attributemanifests", crdConfigs[Policy], "mixer", "istio.io.mixer", "core"),
	crd("bypass", "bypasses", crdConfigs[Policy], "mixer", "bypass", "mixer-adapter"),
	crd("circonus", "circonuses", crdConfigs[Policy], "mixer", "circonus", "mixer-adapter"),
	crd("denier", "deniers", crdConfigs[Policy], "mixer", "denier", "mixer-adapter"),
	crd("fluentd", "fluentds", crdConfigs[Policy], "mixer", "fluentd", "mixer-adapter"),
	crd("kubernetesenv", "kubernetesenvs", crdConfigs[Policy], "mixer", "kubernetesenv", "mixer-adapter"),
	crd("listchecker", "listcheckers", crdConfigs[Policy], "mixer", "listchecker", "mixer-adapter"),
	crd("memquota", "memquotas", crdConfigs[Policy], "mixer", "memquota", "mixer-adapter"),
	crd("noop", "noops", crdConfigs[Policy], "mixer", "noop", "mixer-adapter"),
	crd("opa", "opas", crdConfigs[Policy], "mixer", "opa", "mixer-adapter"),
	crd("prometheus", "prometheuses", crdConfigs[Policy], "mixer", "prometheus", "mixer-adapter"),
	crd("rbac", "rbacs", crdConfigs[Policy], "mixer", "rbac", "mixer-adapter"),
	crd("redisquota", "redisquotas", crdConfigs[Policy], "mixer", "redisquota", "mixer-adapter"), // helm chart misses app:mixer label
	crd("servicecontrol", "servicecontrols", crdConfigs[Policy], "mixer", "servicecontrol", "mixer-adapter"),
	crd("signalfx", "signalfxs", crdConfigs[Policy], "mixer", "signalfx", "mixer-adapter"),
	crd("solarwinds", "solarwindses", crdConfigs[Policy], "mixer", "solarwinds", "mixer-adapter"),
	crd("stackdriver", "stackdrivers", crdConfigs[Policy], "mixer", "stackdriver", "mixer-adapter"),
	crd("statsd", "statsds", crdConfigs[Policy], "mixer", "statsd", "mixer-adapter"),
	crd("stdio", "stdios", crdConfigs[Policy], "mixer", "stdio", "mixer-adapter"),
	crd("apikey", "apikeys", crdConfigs[Policy], "mixer", "apikey", "mixer-instance"),
	crd("authorization", "authorizations", crdConfigs[Policy], "mixer", "authorization", "mixer-instance"),
	crd("checknothing", "checknothings", crdConfigs[Policy], "mixer", "checknothing", "mixer-instance"),
	crd("kubernetes", "kuberneteses", crdConfigs[Policy], "mixer", "adapter.template.kubernetes", "mixer-instance"),
	crd("listentry", "listentries", crdConfigs[Policy], "mixer", "listentry", "mixer-instance"),
	crd("logentry", "logentries", crdConfigs[Policy], "mixer", "logentry", "mixer-instance"),
	crd("edge", "edges", crdConfigs[Policy], "mixer", "edge", "mixer-instance"),
	crd("metric", "metrics", crdConfigs[Policy], "mixer", "metric", "mixer-instance"),
	crd("quota", "quotas", crdConfigs[Policy], "mixer", "quota", "mixer-instance"),
	crd("reportnothing", "reportnothings", crdConfigs[Policy], "mixer", "reportnothing", "mixer-instance"),
	crd("servicecontrolreport", "servicecontrolreports", crdConfigs[Policy], "mixer", "servicecontrolreport", "mixer-instance"),
	crd("tracespan", "tracespans", crdConfigs[Policy], "mixer", "tracespan", "mixer-instance"),
	crd("RbacConfig", "RbacConfigs", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac"),
	crd("ServiceRole", "ServiceRoles", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac"),
	crd("ServiceRoleBinding", "ServiceRoleBindings", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac"),
	crd("adapter", "adapters", crdConfigs[Policy], "mixer", "istio.io.mixer", "rbac"),
	crd("instance", "instances", crdConfigs[Policy], "mixer", "instance", "mixer-instance"),
	crd("template", "templates", crdConfigs[Policy], "mixer", "template", "mixer-template"),
	crd("handler", "handlers", crdConfigs[Policy], "mixer", "handler", "mixer-handler"),
}

func crd(kind string, plural string, config crdConfig, appLabel string, pckLabel string, istioLabel string) *extensionsobj.CustomResourceDefinition {
	return crdL(kind, plural, config, appLabel, pckLabel, istioLabel, false)
}

func crdL(kind string, plural string, config crdConfig, appLabel string, pckLabel string, istioLabel string, list bool) *extensionsobj.CustomResourceDefinition {
	singularName := strings.ToLower(kind)
	pluralName := strings.ToLower(plural)
	labels := make(map[string]string)
	if len(appLabel) > 0 {
		labels["app"] = appLabel
	}
	if len(pckLabel) > 0 {
		labels["package"] = pckLabel
	}
	if len(istioLabel) > 0 {
		labels["istio"] = istioLabel
	}
	crd := &extensionsobj.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s.%s.istio.io", pluralName, config.group),
			Labels: labels,
		},
		Spec: extensionsobj.CustomResourceDefinitionSpec{
			Group:   fmt.Sprintf("%s.istio.io", config.group),
			Version: config.version,
			Scope:   extensionsobj.NamespaceScoped,
			Names: extensionsobj.CustomResourceDefinitionNames{
				Plural:     pluralName,
				Kind:       kind,
				Singular:   singularName,
				Categories: config.categories,
			},
		},
	}
	if list {
		crd.Spec.Names.ListKind = kind + "List"
	}
	return crd
}
