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

package config

import "github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1"

// Configuration defines operator configuration
type Configuration struct {
	WebhookServiceAddress    string                `json:"webhookServiceAddress,omitempty"`
	WebhookConfigurationName string                `json:"webhookConfigurationName,omitempty"`
	SupportedJWTPolicy       v1beta1.JWTPolicyType `json:"supportedJWTPolicy,omitempty"`
}
