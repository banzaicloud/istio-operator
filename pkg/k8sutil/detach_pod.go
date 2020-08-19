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

	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DetachPodsFromDeployment(c client.Client, deployment *appsv1.Deployment, log logr.Logger, additionalLabels ...map[string]string) error {
	ls, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return err
	}

	pods := &corev1.PodList{}
	err = c.List(context.Background(), &client.ListOptions{
		LabelSelector: ls,
	}, pods)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if len(pod.OwnerReferences) != 1 {
			log.V(1).Info("evaluting pod for detaching", "action", "skip", "deploymentName", deployment.Name, "name", pod.Name, "reason", "notExactlyOneOwnerReference")
			continue
		}
		ownerRef := pod.OwnerReferences[0]
		if ownerRef.Kind != "ReplicaSet" {
			log.V(1).Info("evaluting pod for detaching", "action", "skip", "deploymentName", deployment.Name, "name", pod.Name, "reason", "ownerIsNotReplicaSet")
			continue
		}
		rs := &appsv1.ReplicaSet{}
		err = c.Get(context.Background(), client.ObjectKey{
			Name:      ownerRef.Name,
			Namespace: pod.Namespace,
		}, rs)
		if err != nil {
			return err
		}

		if len(rs.OwnerReferences) != 1 {
			log.V(1).Info("evaluting pod for detaching", "action", "skip", "deploymentName", deployment.Name, "name", pod.Name, "reason", "replicaSetHasMultipleOwners")
			continue
		}

		if rs.OwnerReferences[0].UID != deployment.UID {
			log.V(1).Info("evaluting pod for detaching", "action", "skip", "deploymentName", deployment.Name, "name", pod.Name, "reason", "replicaSetOwnerMismatch")
			continue
		}

		log.V(1).Info("evaluting pod for detaching", "action", "detach", "deploymentName", deployment.Name, "name", pod.Name)

		p := pod.DeepCopy()
		p.OwnerReferences = nil
		if p.Labels == nil {
			p.Labels = make(map[string]string)
		} else {
			delete(p.Labels, "pod-template-hash")
		}
		p.Labels = util.MergeStringMaps(p.Labels, util.MergeMultipleStringMaps(additionalLabels...))

		err = c.Update(context.Background(), p)
		if err != nil {
			return err
		}
	}

	return nil
}
