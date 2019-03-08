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

package objectmatch

import (
	"encoding/json"

	"github.com/goph/emperror"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
)

type mutatingWebhookConfigurationMatcher struct {
	objectMatcher ObjectMatcher
}

func NewMutatingWebhookConfigurationMatcher(objectMatcher ObjectMatcher) *mutatingWebhookConfigurationMatcher {
	return &mutatingWebhookConfigurationMatcher{
		objectMatcher: objectMatcher,
	}
}

// Match compares two admissionv1beta1.MutatingWebhookConfiguration objects
func (m mutatingWebhookConfigurationMatcher) Match(old, new *admissionv1beta1.MutatingWebhookConfiguration) (bool, error) {
	type MutatingWebhookConfiguration struct {
		ObjectMeta
		Webhooks []admissionv1beta1.Webhook `json:"webhooks,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	}

	old.Webhooks = nullCABundleConditionally(old.Webhooks, new.Webhooks)

	oldData, err := json.Marshal(MutatingWebhookConfiguration{
		ObjectMeta: m.objectMatcher.GetObjectMeta(old.ObjectMeta),
		Webhooks:   old.Webhooks,
	})
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal old object", "name", old.Name)
	}

	newObject := MutatingWebhookConfiguration{
		ObjectMeta: m.objectMatcher.GetObjectMeta(new.ObjectMeta),
		Webhooks:   new.Webhooks,
	}
	newData, err := json.Marshal(newObject)
	if err != nil {
		return false, emperror.WrapWith(err, "could not marshal new object", "name", new.Name)
	}

	matched, err := m.objectMatcher.MatchJSON(oldData, newData, newObject)
	if err != nil {
		return false, emperror.WrapWith(err, "could not match objects", "name", new.Name)
	}

	return matched, nil
}

// nullCABundleConditionally nils ClientConfig.CABundle value in the old object if it is nil in the new to avoid conflict
func nullCABundleConditionally(oldWebhooks, newWebhooks []admissionv1beta1.Webhook) []admissionv1beta1.Webhook {
	for i, wh := range oldWebhooks {
		nwh := getWebhookByName(wh.Name, newWebhooks)
		if nwh == nil || nwh.ClientConfig.CABundle != nil {
			continue
		}
		wh.ClientConfig.CABundle = nil
		oldWebhooks[i] = wh
	}

	return oldWebhooks
}

// getWebhookByName gets webhook from webhooks by its name
func getWebhookByName(name string, webhooks []admissionv1beta1.Webhook) *admissionv1beta1.Webhook {
	for _, webhook := range webhooks {
		if webhook.Name == name {
			return &webhook
		}
	}

	return nil
}
