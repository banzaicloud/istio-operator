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

package k8sutil

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil/objectmatch"
)

func Reconcile(log logr.Logger, client runtimeClient.Client, desired runtime.Object) error {
	desiredType := reflect.TypeOf(desired)
	var current = desired.DeepCopyObject()
	key, err := runtimeClient.ObjectKeyFromObject(current)
	if err != nil {
		return emperror.With(err, "kind", desiredType)
	}
	log = log.WithValues("kind", desiredType, "name", key.Name)

	err = client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "kind", desiredType, "name", key.Name)
	}
	if apierrors.IsNotFound(err) {
		if err := client.Create(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "kind", desiredType, "name", key.Name)
		}
		log.Info("resource created")
	}
	if err == nil {
		objectsEquals, err := objectmatch.New(log).Match(current, desired)
		if err != nil {
			log.Error(err, "could not match objects", "kind", desiredType, "name", key.Name)
		} else if objectsEquals {
			log.V(1).Info("resource is in sync")
			return nil
		}

		switch desired.(type) {
		default:
			return emperror.With(errors.New("unexpected resource type"), "kind", desiredType, "name", key.Name)
		case *corev1.Namespace:
			ns := desired.(*corev1.Namespace)
			ns.ResourceVersion = current.(*corev1.Namespace).ResourceVersion
			desired = ns
		case *corev1.ServiceAccount:
			sa := desired.(*corev1.ServiceAccount)
			sa.ResourceVersion = current.(*corev1.ServiceAccount).ResourceVersion
			desired = sa
		case *rbacv1.ClusterRole:
			cr := desired.(*rbacv1.ClusterRole)
			cr.ResourceVersion = current.(*rbacv1.ClusterRole).ResourceVersion
			desired = cr
		case *rbacv1.ClusterRoleBinding:
			crb := desired.(*rbacv1.ClusterRoleBinding)
			crb.ResourceVersion = current.(*rbacv1.ClusterRoleBinding).ResourceVersion
			desired = crb
		case *corev1.ConfigMap:
			cm := desired.(*corev1.ConfigMap)
			cm.ResourceVersion = current.(*corev1.ConfigMap).ResourceVersion
			desired = cm
		case *corev1.Service:
			svc := desired.(*corev1.Service)
			svc.ResourceVersion = current.(*corev1.Service).ResourceVersion
			svc.Spec.ClusterIP = current.(*corev1.Service).Spec.ClusterIP
			desired = svc
		case *appsv1.Deployment:
			deploy := desired.(*appsv1.Deployment)
			deploy.ResourceVersion = current.(*appsv1.Deployment).ResourceVersion
			desired = deploy
		case *extensionsv1beta1.Deployment:
			deploy := desired.(*extensionsv1beta1.Deployment)
			deploy.ResourceVersion = current.(*extensionsv1beta1.Deployment).ResourceVersion
			desired = deploy
		case *autoscalingv2beta1.HorizontalPodAutoscaler:
			hpa := desired.(*autoscalingv2beta1.HorizontalPodAutoscaler)
			hpa.ResourceVersion = current.(*autoscalingv2beta1.HorizontalPodAutoscaler).ResourceVersion
			desired = hpa
		case *admissionregistrationv1beta1.MutatingWebhookConfiguration:
			mwc := desired.(*admissionregistrationv1beta1.MutatingWebhookConfiguration)
			mwc.ResourceVersion = current.(*admissionregistrationv1beta1.MutatingWebhookConfiguration).ResourceVersion
			desired = mwc
		case *policyv1beta1.PodDisruptionBudget:
			pdb := desired.(*policyv1beta1.PodDisruptionBudget)
			pdb.ResourceVersion = current.(*policyv1beta1.PodDisruptionBudget).ResourceVersion
			desired = pdb
		case *appsv1.DaemonSet:
			ds := desired.(*appsv1.DaemonSet)
			ds.ResourceVersion = current.(*appsv1.DaemonSet).ResourceVersion
			desired = ds
		case *rbacv1.Role:
			ds := desired.(*rbacv1.Role)
			ds.ResourceVersion = current.(*rbacv1.Role).ResourceVersion
			desired = ds
		case *rbacv1.RoleBinding:
			ds := desired.(*rbacv1.RoleBinding)
			ds.ResourceVersion = current.(*rbacv1.RoleBinding).ResourceVersion
			desired = ds
		}
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "kind", desiredType, "name", key.Name)
		}
		log.Info("resource updated")
	}
	return nil
}

// ReconcileNamespaceLabelsIgnoreNotFound patches namespaces by adding/removing labels, returns without error if namespace is not found
func ReconcileNamespaceLabelsIgnoreNotFound(log logr.Logger, client runtimeClient.Client, namespace string, labels map[string]string, labelsToRemove []string) error {
	var ns = &corev1.Namespace{}
	err := client.Get(context.TODO(), runtimeClient.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("namespace not found, ignoring", "namespace", namespace)
			return nil
		}

		return emperror.WrapWith(err, "getting namespace failed", "namespace", namespace)
	}

	updateNeeded := false
	for dlk, dlv := range labels {
		if ns.Labels == nil {
			ns.Labels = make(map[string]string)
		}
		if clv, ok := ns.Labels[dlk]; !ok || clv != dlv {
			ns.Labels[dlk] = dlv
			updateNeeded = true
		}
	}
	for _, labelKey := range labelsToRemove {
		if _, ok := ns.Labels[labelKey]; ok {
			delete(ns.Labels, labelKey)
			updateNeeded = true
		}
	}
	if updateNeeded {
		if err := client.Update(context.TODO(), ns); err != nil {
			return emperror.WrapWith(err, "updating namespace failed", "namespace", namespace)
		}
		log.Info("namespace labels reconciled", "namespace", namespace, "labels", labels)
	}

	return nil
}
