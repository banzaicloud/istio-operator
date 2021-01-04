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

package gateways

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istiov1beta1 "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"
	"github.com/banzaicloud/istio-operator/pkg/k8sutil"
	"github.com/banzaicloud/istio-operator/pkg/resources"
	"github.com/banzaicloud/istio-operator/pkg/util"
)

const (
	componentName = "meshgateway"
)

type Reconciler struct {
	resources.Reconciler
	gw      *istiov1beta1.MeshGateway
	dynamic dynamic.Interface
	scheme  *runtime.Scheme
}

func New(client client.Client, dc dynamic.Interface, config *istiov1beta1.Istio, gw *istiov1beta1.MeshGateway, scheme *runtime.Scheme) *Reconciler {
	return &Reconciler{
		Reconciler: resources.Reconciler{
			Client: client,
			Config: config,
		},
		gw:      gw,
		dynamic: dc,
		scheme:  scheme,
	}
}

func (r *Reconciler) Reconcile(log logr.Logger) error {
	log = log.WithValues("component", componentName)

	log.Info("Reconciling")

	err := r.waitForIstiod()
	if err != nil {
		return err
	}

	pdbDesiredState := k8sutil.DesiredStateAbsent
	if util.PointerToBool(r.Config.Spec.DefaultPodDisruptionBudget.Enabled) {
		pdbDesiredState = k8sutil.DesiredStatePresent
	}

	hpaDesiredState := k8sutil.DesiredStateAbsent
	if r.gw.Spec.MinReplicas != nil && r.gw.Spec.MaxReplicas != nil && *r.gw.Spec.MaxReplicas > *r.gw.Spec.MinReplicas {
		hpaDesiredState = k8sutil.DesiredStatePresent
	}

	// add specific desired state to support re-creation
	deploymentDesiredState := k8sutil.NewRecreateAwareDeploymentDesiredState(r.Client, r.scheme, log, r.labels())

	for _, res := range []resources.ResourceWithDesiredState{
		{Resource: r.serviceAccount, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.clusterRole, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.clusterRoleBinding, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.role, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.roleBinding, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.deployment, DesiredState: deploymentDesiredState},
		{Resource: r.service, DesiredState: k8sutil.DesiredStatePresent},
		{Resource: r.horizontalPodAutoscaler, DesiredState: hpaDesiredState},
		{Resource: r.podDisruptionBudget, DesiredState: pdbDesiredState},
	} {
		o := res.Resource()
		err := k8sutil.Reconcile(log, r.Client, o, res.DesiredState)
		if err != nil {
			return emperror.WrapWith(err, "failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
	}

	log.Info("Reconciled")

	return nil
}

func (r *Reconciler) waitForIstiod() error {
	if !util.PointerToBool(r.Config.Spec.Istiod.Enabled) || util.PointerToBool(r.Config.Spec.SidecarInjector.Enabled) {
		return nil
	}

	var pods v1.PodList
	ls, err := labels.Parse("app=istiod")
	if err != nil {
		return err
	}

	err = r.Client.List(context.Background(), &pods, client.InNamespace(r.Config.Namespace), client.MatchingLabelsSelector{
		Selector: ls,
	})
	if err != nil {
		return emperror.Wrap(err, "could not list pods")
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			readyContainers := 0
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Ready {
					readyContainers++
				}
			}
			if readyContainers == len(pod.Status.ContainerStatuses) {
				return nil
			}
		}
	}

	return errors.Errorf("Istiod is not running yet")
}
