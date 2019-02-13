package templates

import (
	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/operator/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ObjectMeta(name string, labels map[string]string, owner *istiov1beta1.Config) metav1.ObjectMeta {
	o := metav1.ObjectMeta{
		Name:      name,
		Namespace: owner.Namespace,
		Labels:    labels,
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: owner.APIVersion,
				Kind:       owner.Kind,
				Name:       owner.Name,
				UID:        owner.UID,
			},
		},
	}
	return o
}
