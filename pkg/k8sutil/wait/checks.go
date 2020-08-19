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

package wait

import (
	"github.com/banzaicloud/istio-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceConditionCheck func(runtime.Object, error) bool
type CustomResourceConditionCheck func() (bool, error)

func ExistsConditionCheck(obj runtime.Object, k8serror error) bool {
	return k8serror == nil
}

func NonExistsConditionCheck(obj runtime.Object, k8serror error) bool {
	return k8serrors.IsNotFound(k8serror)
}

func CRDEstablishedConditionCheck(obj runtime.Object, k8serror error) bool {
	var resource *apiextensionsv1beta1.CustomResourceDefinition
	var ok bool
	if resource, ok = obj.(*apiextensionsv1beta1.CustomResourceDefinition); !ok {
		return true
	}

	for _, condition := range resource.Status.Conditions {
		if condition.Type == apiextensionsv1beta1.Established {
			if condition.Status == apiextensionsv1beta1.ConditionTrue {
				return true
			}
		}
	}

	return false
}

func ReadyReplicasConditionCheck(obj runtime.Object, k8serror error) bool {
	var deployment *appsv1.Deployment
	var ok bool

	if deployment, ok = obj.(*appsv1.Deployment); ok {
		return util.PointerToInt32(deployment.Spec.Replicas) == deployment.Status.ReadyReplicas && deployment.Status.ReadyReplicas == deployment.Status.Replicas
	}

	var statefulset *appsv1.StatefulSet
	if statefulset, ok = obj.(*appsv1.StatefulSet); ok {
		return util.PointerToInt32(statefulset.Spec.Replicas) == statefulset.Status.ReadyReplicas && statefulset.Status.ReadyReplicas == statefulset.Status.Replicas
	}

	var daemonset *appsv1.DaemonSet
	if daemonset, ok = obj.(*appsv1.DaemonSet); ok {
		return daemonset.Status.DesiredNumberScheduled == daemonset.Status.NumberReady
	}

	// return true for unconvertable objects
	return true
}
