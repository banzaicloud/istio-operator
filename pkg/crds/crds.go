/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crds

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	extensionsobj "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil/objectmatch"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "crds"
)

type CrdOperator struct {
	crds   []*extensionsobj.CustomResourceDefinition
	config *rest.Config
}

type configType int

const (
	Networking configType = iota
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
	Networking:     {"v1alpha3", "networking", []string{"istio-io", "networking-istio-io"}},
	Authentication: {"v1alpha1", "authentication", []string{"istio-io", "authentication-istio-io"}},
	Apim:           {"v1alpha2", "config", []string{"istio-io", "apim-istio-io"}},
	Policy:         {"v1alpha2", "config", []string{"istio-io", "policy-istio-io"}},
	Rbac:           {"v1alpha1", "rbac", []string{"istio-io", "rbac-istio-io"}},
}

func New(cfg *rest.Config, crds []*extensionsobj.CustomResourceDefinition) (*CrdOperator, error) {
	return &CrdOperator{
		crds:   crds,
		config: cfg,
	}, nil
}

func InitCrds() []*extensionsobj.CustomResourceDefinition {
	return []*extensionsobj.CustomResourceDefinition{
		crdL("VirtualService", "VirtualServices", crdConfigs[Networking], "istio-pilot", "", "", extensionsobj.NamespaceScoped, true),
		crdL("DestinationRule", "DestinationRules", crdConfigs[Networking], "istio-pilot", "", "", extensionsobj.NamespaceScoped, true),
		crdL("ServiceEntry", "ServiceEntries", crdConfigs[Networking], "istio-pilot", "", "", extensionsobj.NamespaceScoped, true),
		crd("Gateway", "Gateways", crdConfigs[Networking], "istio-pilot", "", "", extensionsobj.NamespaceScoped),
		crd("EnvoyFilter", "EnvoyFilters", crdConfigs[Networking], "istio-pilot", "", "", extensionsobj.NamespaceScoped),
		crd("Policy", "Policies", crdConfigs[Authentication], "", "", "", extensionsobj.NamespaceScoped),
		crd("MeshPolicy", "MeshPolicies", crdConfigs[Authentication], "", "", "", extensionsobj.ClusterScoped),
		crd("HTTPAPISpecBinding", "HTTPAPISpecBindings", crdConfigs[Apim], "", "", "", extensionsobj.NamespaceScoped),
		crd("HTTPAPISpec", "HTTPAPISpecs", crdConfigs[Apim], "", "", "", extensionsobj.NamespaceScoped),
		crd("QuotaSpecBinding", "QuotaSpecBindings", crdConfigs[Apim], "", "", "", extensionsobj.NamespaceScoped),
		crd("QuotaSpec", "QuotaSpecs", crdConfigs[Apim], "", "", "", extensionsobj.NamespaceScoped),
		crd("rule", "rules", crdConfigs[Policy], "mixer", "istio.io.mixer", "core", extensionsobj.NamespaceScoped),
		crd("attributemanifest", "attributemanifests", crdConfigs[Policy], "mixer", "istio.io.mixer", "core", extensionsobj.NamespaceScoped),
		crd("bypass", "bypasses", crdConfigs[Policy], "mixer", "bypass", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("circonus", "circonuses", crdConfigs[Policy], "mixer", "circonus", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("denier", "deniers", crdConfigs[Policy], "mixer", "denier", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("fluentd", "fluentds", crdConfigs[Policy], "mixer", "fluentd", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("kubernetesenv", "kubernetesenvs", crdConfigs[Policy], "mixer", "kubernetesenv", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("listchecker", "listcheckers", crdConfigs[Policy], "mixer", "listchecker", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("memquota", "memquotas", crdConfigs[Policy], "mixer", "memquota", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("noop", "noops", crdConfigs[Policy], "mixer", "noop", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("opa", "opas", crdConfigs[Policy], "mixer", "opa", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("prometheus", "prometheuses", crdConfigs[Policy], "mixer", "prometheus", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("rbac", "rbacs", crdConfigs[Policy], "mixer", "rbac", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("redisquota", "redisquotas", crdConfigs[Policy], "mixer", "redisquota", "mixer-adapter", extensionsobj.NamespaceScoped), // helm chart misses app:mixer label
		crd("servicecontrol", "servicecontrols", crdConfigs[Policy], "mixer", "servicecontrol", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("signalfx", "signalfxs", crdConfigs[Policy], "mixer", "signalfx", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("solarwinds", "solarwindses", crdConfigs[Policy], "mixer", "solarwinds", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("stackdriver", "stackdrivers", crdConfigs[Policy], "mixer", "stackdriver", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("statsd", "statsds", crdConfigs[Policy], "mixer", "statsd", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("stdio", "stdios", crdConfigs[Policy], "mixer", "stdio", "mixer-adapter", extensionsobj.NamespaceScoped),
		crd("apikey", "apikeys", crdConfigs[Policy], "mixer", "apikey", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("authorization", "authorizations", crdConfigs[Policy], "mixer", "authorization", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("checknothing", "checknothings", crdConfigs[Policy], "mixer", "checknothing", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("kubernetes", "kuberneteses", crdConfigs[Policy], "mixer", "adapter.template.kubernetes", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("listentry", "listentries", crdConfigs[Policy], "mixer", "listentry", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("logentry", "logentries", crdConfigs[Policy], "mixer", "logentry", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("edge", "edges", crdConfigs[Policy], "mixer", "edge", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("metric", "metrics", crdConfigs[Policy], "mixer", "metric", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("quota", "quotas", crdConfigs[Policy], "mixer", "quota", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("reportnothing", "reportnothings", crdConfigs[Policy], "mixer", "reportnothing", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("servicecontrolreport", "servicecontrolreports", crdConfigs[Policy], "mixer", "servicecontrolreport", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("tracespan", "tracespans", crdConfigs[Policy], "mixer", "tracespan", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("RbacConfig", "RbacConfigs", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac", extensionsobj.NamespaceScoped),
		crd("ServiceRole", "ServiceRoles", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac", extensionsobj.NamespaceScoped),
		crd("ServiceRoleBinding", "ServiceRoleBindings", crdConfigs[Rbac], "mixer", "istio.io.mixer", "rbac", extensionsobj.NamespaceScoped),
		crd("adapter", "adapters", crdConfigs[Policy], "mixer", "istio.io.mixer", "rbac", extensionsobj.NamespaceScoped),
		crd("instance", "instances", crdConfigs[Policy], "mixer", "instance", "mixer-instance", extensionsobj.NamespaceScoped),
		crd("template", "templates", crdConfigs[Policy], "mixer", "template", "mixer-template", extensionsobj.NamespaceScoped),
		crd("handler", "handlers", crdConfigs[Policy], "mixer", "handler", "mixer-handler", extensionsobj.NamespaceScoped),
	}
}

func crd(kind string, plural string, config crdConfig, appLabel string, pckLabel string, istioLabel string, scope extensionsobj.ResourceScope) *extensionsobj.CustomResourceDefinition {
	return crdL(kind, plural, config, appLabel, pckLabel, istioLabel, scope, false)
}

func crdL(kind string, plural string, config crdConfig, appLabel string, pckLabel string, istioLabel string, scope extensionsobj.ResourceScope, list bool) *extensionsobj.CustomResourceDefinition {
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
			Versions: []extensionsobj.CustomResourceDefinitionVersion{
				{
					Name:    config.version,
					Served:  true,
					Storage: true,
				},
			},
			Scope: scope,
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

func (r *CrdOperator) Reconcile(config *istiov1beta1.Istio, log logr.Logger) error {
	log = log.WithValues("component", componentName)
	apiExtensions, err := apiextensionsclient.NewForConfig(r.config)
	if err != nil {
		return emperror.Wrap(err, "instantiating apiextensions client failed")
	}
	crdClient := apiExtensions.ApiextensionsV1beta1().CustomResourceDefinitions()
	for _, crd := range r.crds {
		log := log.WithValues("kind", crd.Spec.Names.Kind)
		current, err := crdClient.Get(crd.Name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return emperror.WrapWith(err, "getting CRD failed", "kind", crd.Spec.Names.Kind)
		}
		if apierrors.IsNotFound(err) {
			if config.Name != "" {
				crd.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
					{
						Kind:               config.Kind,
						APIVersion:         config.APIVersion,
						Name:               config.Name,
						UID:                config.GetUID(),
						Controller:         util.BoolPointer(true),
						BlockOwnerDeletion: util.BoolPointer(true),
					},
				}
			}
			if _, err := crdClient.Create(crd); err != nil {
				return emperror.WrapWith(err, "creating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD created")
		}
		if err == nil {
			objectsEquals, err := objectmatch.Match(current, crd)
			if err != nil {
				log.Error(err, "could not match objects", "kind", crd.Spec.Names.Kind)
			} else if objectsEquals {
				log.V(1).Info("CRD is in sync")
				continue
			}
			crd.ResourceVersion = current.ResourceVersion
			if _, err := crdClient.Update(crd); err != nil {
				return emperror.WrapWith(err, "updating CRD failed", "kind", crd.Spec.Names.Kind)
			}
			log.Info("CRD updated")
		}
	}

	log.Info("Reconciled")

	return nil
}
