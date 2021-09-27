/*
Copyright 2021 Cisco Systems, Inc. and/or its affiliates.

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

package models

// ClusterRegistryConfiguration contains the settings to cooperate with the cluster registry APIs
type ClusterRegistryConfiguration struct {
	ClusterAPI        ClusterAPIConfiguration        `json:"clusterApi,omitempty"`
	ResourceSyncRules ResourceSyncRulesConfiguration `json:"resourceSyncRules,omitempty"`
}

type ClusterAPIConfiguration struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ResourceSyncRulesConfiguration struct {
	Enabled bool `json:"enabled,omitempty"`
}
