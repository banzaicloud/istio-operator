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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(log logr.Logger, client runtimeClient.Client, desired runtime.Object) error {
	log = log.WithValues("type", reflect.TypeOf(desired))
	var current = desired.DeepCopyObject()
	key, err := runtimeClient.ObjectKeyFromObject(current)
	if err != nil {
		return emperror.With(err)
	}
	err = client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return emperror.WrapWith(err, "getting resource failed", "resource", desired.GetObjectKind().GroupVersionKind(), "type", reflect.TypeOf(desired))
	}
	if apierrors.IsNotFound(err) {
		if err := client.Create(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "creating resource failed", "resource", desired.GetObjectKind().GroupVersionKind(), "type", reflect.TypeOf(desired))
		}
		log.Info("resource created", "resource", desired.GetObjectKind().GroupVersionKind())
	}
	if err == nil {
		switch desired.(type) {
		default:
			return emperror.With(errors.New("unexpected resource type"), "type", reflect.TypeOf(desired))
		case *apiv1.Namespace:
			ns := desired.(*apiv1.Namespace)
			ns.ResourceVersion = current.(*apiv1.Namespace).ResourceVersion
			desired = ns
		case *apiv1.ServiceAccount:
			sa := desired.(*apiv1.ServiceAccount)
			sa.ResourceVersion = current.(*apiv1.ServiceAccount).ResourceVersion
			desired = sa
		case *rbacv1.ClusterRole:
			cr := desired.(*rbacv1.ClusterRole)
			cr.ResourceVersion = current.(*rbacv1.ClusterRole).ResourceVersion
			desired = cr
		case *rbacv1.ClusterRoleBinding:
			crb := desired.(*rbacv1.ClusterRoleBinding)
			crb.ResourceVersion = current.(*rbacv1.ClusterRoleBinding).ResourceVersion
			desired = crb
		case *apiv1.ConfigMap:
			cm := desired.(*apiv1.ConfigMap)
			cm.ResourceVersion = current.(*apiv1.ConfigMap).ResourceVersion
			desired = cm
		case *apiv1.Service:
			svc := desired.(*apiv1.Service)
			svc.ResourceVersion = current.(*apiv1.Service).ResourceVersion
			svc.Spec.ClusterIP = current.(*apiv1.Service).Spec.ClusterIP
			desired = svc
		case *appsv1.Deployment:
			deploy := desired.(*appsv1.Deployment)
			deploy.ResourceVersion = current.(*appsv1.Deployment).ResourceVersion
			desired = deploy
		case *v2beta1.HorizontalPodAutoscaler:
			hpa := desired.(*v2beta1.HorizontalPodAutoscaler)
			hpa.ResourceVersion = current.(*v2beta1.HorizontalPodAutoscaler).ResourceVersion
			desired = hpa
		}
		if err := client.Update(context.TODO(), desired); err != nil {
			return emperror.WrapWith(err, "updating resource failed", "resource", desired.GetObjectKind().GroupVersionKind(), "type", reflect.TypeOf(desired))
		}
		log.Info("resource updated", "resource", desired.GetObjectKind().GroupVersionKind())
	}
	return nil
}
