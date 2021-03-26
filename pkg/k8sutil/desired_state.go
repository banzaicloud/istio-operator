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
	"github.com/goph/emperror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	detachedPodLabel = "istio.banzaicloud.io/detached-pod"
)

type DesiredState interface {
	AfterRecreate(current, desired runtime.Object) error
	BeforeRecreate(current, desired runtime.Object) error
	ShouldRecreate(current, desired runtime.Object) (bool, error)
	AfterCreate(desired runtime.Object) error
	BeforeCreate(desired runtime.Object) error
	ShouldCreate(desired runtime.Object) (bool, error)
	AfterUpdate(current, desired runtime.Object, inSync bool) error
	BeforeUpdate(current, desired runtime.Object) error
	ShouldUpdate(current, desired runtime.Object) (bool, error)
	AfterDelete(current runtime.Object) error
	BeforeDelete(current runtime.Object) error
	ShouldDelete(current runtime.Object) (bool, error)
}

type StaticDesiredState string

func (s StaticDesiredState) AfterRecreate(_, _ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeRecreate(_, _ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) ShouldRecreate(_, _ runtime.Object) (bool, error) {
	return true, nil
}

func (s StaticDesiredState) AfterCreate(_ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeCreate(_ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) ShouldCreate(_ runtime.Object) (bool, error) {
	return true, nil
}

func (s StaticDesiredState) AfterUpdate(_, _ runtime.Object, _ bool) error {
	return nil
}

func (s StaticDesiredState) BeforeUpdate(_, _ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) ShouldUpdate(_, _ runtime.Object) (bool, error) {
	return true, nil
}

func (s StaticDesiredState) AfterDelete(_ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeDelete(_ runtime.Object) error {
	return nil
}

func (s StaticDesiredState) ShouldDelete(_ runtime.Object) (bool, error) {
	return true, nil
}

const (
	DesiredStatePresent StaticDesiredState = "present"
	DesiredStateAbsent  StaticDesiredState = "absent"
	DesiredStateExists  StaticDesiredState = "exists"
)

type MeshGatewayCreateOnlyDesiredState struct {
	StaticDesiredState
}

func (MeshGatewayCreateOnlyDesiredState) ShouldUpdate(current, desired runtime.Object) (bool, error) {
	currentMeta, err := meta.Accessor(current)
	if err != nil {
		return false, emperror.WrapWith(err, "could not get desired object metadata")
	}
	if len(currentMeta.GetOwnerReferences()) > 0 {
		return true, nil
	}

	return false, nil
}

type RecreateAwareDeploymentDesiredState struct {
	StaticDesiredState

	client    client.Client
	scheme    *runtime.Scheme
	log       logr.Logger
	podLabels map[string]string
}

func NewRecreateAwareDeploymentDesiredState(c client.Client, scheme *runtime.Scheme, log logr.Logger, podLabels map[string]string) RecreateAwareDeploymentDesiredState {
	podLabels = util.MergeMultipleStringMaps(map[string]string{
		detachedPodLabel: "true",
	}, podLabels)

	return RecreateAwareDeploymentDesiredState{
		client:    c,
		scheme:    scheme,
		log:       log,
		podLabels: podLabels,
	}
}

func (r RecreateAwareDeploymentDesiredState) AfterRecreate(current, desired runtime.Object) error {
	var deployment *appsv1.Deployment
	var ok bool
	if deployment, ok = desired.(*appsv1.Deployment); !ok {
		return nil
	}

	return r.waitForDeploymentAndRemoveDetached(deployment)
}

func (r RecreateAwareDeploymentDesiredState) AfterUpdate(current, desired runtime.Object, inSync bool) error {
	var deployment *appsv1.Deployment
	var ok bool
	if deployment, ok = desired.(*appsv1.Deployment); !ok {
		return nil
	}

	return r.waitForDeploymentAndRemoveDetached(deployment)
}

func (r RecreateAwareDeploymentDesiredState) BeforeRecreate(current, desired runtime.Object) error {
	var deployment *appsv1.Deployment
	var ok bool
	if deployment, ok = current.(*appsv1.Deployment); !ok {
		return nil
	}

	err := DetachPodsFromDeployment(r.client, deployment, r.log, r.podLabels)
	if err != nil {
		return err
	}

	return nil
}

func (r RecreateAwareDeploymentDesiredState) waitForDeploymentAndRemoveDetached(deployment *appsv1.Deployment) error {
	rcc := wait.NewResourceConditionChecks(r.client, wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1,
		Jitter:   0,
		Steps:    12,
	}, r.log.WithName("wait"), r.scheme)

	err := rcc.WaitForResources("readiness", []runtime.Object{deployment}, wait.ExistsConditionCheck, wait.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	pods := &corev1.PodList{}
	err = r.client.List(context.Background(), pods, client.InNamespace(deployment.GetNamespace()), client.MatchingLabelsSelector{
		Selector: labels.Set(r.podLabels).AsSelector(),
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		r.log.Info("removing detached pods")
		err = r.client.Delete(context.Background(), &pod)
		if err != nil {
			return err
		}
	}

	return nil
}
