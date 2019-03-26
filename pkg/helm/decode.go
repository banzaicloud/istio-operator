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

package helm

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/releaseutil"
)

var regexpIstioCR = regexp.MustCompile("apiVersion: \"?([a-z]+)\\.istio\\.io/([a-z0-9]+)\"?")

func DecodeObjects(log logr.Logger, manifests []manifest.Manifest) ([]metav1.Object, error) {
	var allErrors []error
	origLogger := log
	defer func() { log = origLogger }()
	var resources []metav1.Object
	for _, manifest := range manifests {
		log = origLogger.WithValues("manifest", manifest.Name)
		if !strings.HasSuffix(manifest.Name, ".yaml") {
			log.V(2).Info("Skipping rendering of manifest")
			continue
		}
		log.V(2).Info("Processing resources from manifest")
		// split the manifest into individual objects
		objects := releaseutil.SplitManifests(manifest.Content)
		for _, raw := range objects {
			var obj runtime.Object
			var err error

			decode := scheme.Codecs.UniversalDeserializer().Decode

			// handle Istio CRs differently by decoding them to unstructured objects
			if regexpIstioCR.MatchString(raw) {
				obj, _, err = decode([]byte(raw), nil, &unstructured.Unstructured{})
				if err != nil {
					allErrors = append(allErrors, err)
					continue
				}
			} else {
				obj, _, err = decode([]byte(raw), nil, nil)
				if err != nil {
					allErrors = append(allErrors, err)
					continue
				}
			}

			log.Info("decoded object", "type", reflect.TypeOf(obj))

			switch obj.(type) {
			case *corev1.Namespace:
				ns := obj.(*corev1.Namespace)
				resources = append(resources, ns)
			case *corev1.ServiceAccount:
				sa := obj.(*corev1.ServiceAccount)
				resources = append(resources, sa)
			case *rbacv1.ClusterRole:
				cr := obj.(*rbacv1.ClusterRole)
				resources = append(resources, cr)
			case *rbacv1.ClusterRoleBinding:
				crb := obj.(*rbacv1.ClusterRoleBinding)
				resources = append(resources, crb)
			case *corev1.ConfigMap:
				cm := obj.(*corev1.ConfigMap)
				resources = append(resources, cm)
			case *corev1.Service:
				svc := obj.(*corev1.Service)
				resources = append(resources, svc)
			case *appsv1beta1.Deployment:
				deployment := obj.(*appsv1beta1.Deployment)
				resources = append(resources, deployment)
			case *extensionsv1beta1.Deployment:
				deployment := obj.(*extensionsv1beta1.Deployment)
				resources = append(resources, deployment)
			case *autoscalingv2beta1.HorizontalPodAutoscaler:
				hpa := obj.(*autoscalingv2beta1.HorizontalPodAutoscaler)
				resources = append(resources, hpa)
			case *admissionregistrationv1beta1.MutatingWebhookConfiguration:
				mwc := obj.(*admissionregistrationv1beta1.MutatingWebhookConfiguration)
				resources = append(resources, mwc)
			case *policyv1beta1.PodDisruptionBudget:
				pdb := obj.(*policyv1beta1.PodDisruptionBudget)
				resources = append(resources, pdb)
			case *appsv1.DaemonSet:
				ds := obj.(*appsv1.DaemonSet)
				resources = append(resources, ds)
			case *rbacv1.Role:
				role := obj.(*rbacv1.Role)
				resources = append(resources, role)
			case *rbacv1.RoleBinding:
				rb := obj.(*rbacv1.RoleBinding)
				resources = append(resources, rb)
			case *unstructured.Unstructured:
				us := obj.(*unstructured.Unstructured)
				resources = append(resources, us)
			}
		}
	}

	return resources, utilerrors.NewAggregate(allErrors)
}
