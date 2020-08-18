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
