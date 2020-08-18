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
	"time"

	"github.com/banzaicloud/istio-operator/pkg/k8sutil/wait"
	"github.com/banzaicloud/istio-operator/pkg/util"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredState interface {
	AfterRecreate(current, desired runtime.Object) error
	BeforeRecreate(current, desired runtime.Object) error
}

type StaticDesiredState string

func (s StaticDesiredState) AfterRecreate(current, desired runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeRecreate(current, desired runtime.Object) error {
	return nil
}

const (
	DesiredStatePresent StaticDesiredState = "present"
	DesiredStateAbsent  StaticDesiredState = "absent"
	DesiredStateExists  StaticDesiredState = "exists"
)

type RecreateAwareDesiredState struct {
	afterCreateFunc  func(current, desired runtime.Object) error
	beforeCreateFunc func(current, desired runtime.Object) error
}

func NewRecreateAwareDesiredState(afterCreateFunc, beforeCreateFunc func(current, desired runtime.Object) error) RecreateAwareDesiredState {
	return RecreateAwareDesiredState{
		afterCreateFunc:  afterCreateFunc,
		beforeCreateFunc: beforeCreateFunc,
	}
}

func (s RecreateAwareDesiredState) AfterRecreate(current, desired runtime.Object) error {
	return s.afterCreateFunc(current, desired)
}

func (s RecreateAwareDesiredState) BeforeRecreate(current, desired runtime.Object) error {
	return s.beforeCreateFunc(current, desired)
}

func DeploymentDesiredStateWithReCreateHandling(c client.Client, scheme *runtime.Scheme, log logr.Logger, podLabels map[string]string) RecreateAwareDesiredState {
	podLabels = util.MergeMultipleStringMaps(map[string]string{
		"detached": "true",
	}, podLabels)

	return NewRecreateAwareDesiredState(
		// after re-create
		func(current, desired runtime.Object) error {
			var deployment *appsv1.Deployment
			var ok bool
			if deployment, ok = desired.(*appsv1.Deployment); !ok {
				return nil
			}

			rcc := wait.NewResourceConditionChecks(c, wait.Backoff{
				Duration: time.Second * 5,
				Factor:   1,
				Jitter:   0,
				Steps:    12,
			}, log.WithName("wait"), scheme)

			err := rcc.WaitForResources("readiness", []runtime.Object{deployment}, wait.ExistsConditionCheck, wait.ReadyReplicasConditionCheck)
			if err != nil {
				return err
			}

			pods := &corev1.PodList{}
			err = c.List(context.Background(), &client.ListOptions{
				Namespace:     deployment.GetNamespace(),
				LabelSelector: labels.Set(podLabels).AsSelector(),
			}, pods)
			if err != nil {
				return err
			}
			for _, pod := range pods.Items {
				log.Info("removing detached pods")
				err = c.Delete(context.Background(), &pod)
				if err != nil {
					return err
				}
			}

			return nil
		},
		// before re-create
		func(current, desired runtime.Object) error {
			var deployment *appsv1.Deployment
			var ok bool
			if deployment, ok = current.(*appsv1.Deployment); !ok {
				return nil
			}

			err := DetachPodsFromDeployment(c, deployment, log, podLabels)
			if err != nil {
				return err
			}

			return nil
		})
}
