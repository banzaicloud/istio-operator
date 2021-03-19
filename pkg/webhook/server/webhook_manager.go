/*
Copyright 2021 Banzai Cloud.

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

package server

import (
	"context"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// ManagerWithCustomClient is able to inject custom client into a runnable
type ManagerWithCustomClient struct {
	manager.Manager
	client client.Client
}

func (m *ManagerWithCustomClient) Add(r manager.Runnable) error {
	err := m.Manager.Add(r)
	if err != nil {
		return err
	}

	// override client in the runnable to handle webhook update
	if _, err := inject.ClientInto(m.client, r); err != nil {
		return err
	}

	return nil
}

// WebhookUpdateClient is a specific runtime client that modifies validating webhook sideeffect property
type WebhookUpdateClient struct {
	client.Client
}

func (w *WebhookUpdateClient) Update(ctx context.Context, obj runtime.Object) error {
	se := admissionregistrationv1beta1.SideEffectClassNone
	if o, ok := obj.(*admissionregistrationv1beta1.ValidatingWebhookConfiguration); ok {
		for i := range o.Webhooks {
			o.Webhooks[i].SideEffects = &se
		}
	}

	return w.Client.Update(ctx, obj)
}
