/*
Copyright 2020 Banzai Cloud.

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
	corev1 "k8s.io/api/core/v1"
)

// MergeEnvVars merges env variables by name
func MergeEnvVars(envs []corev1.EnvVar, additionalEnvs []corev1.EnvVar) []corev1.EnvVar {
	if len(additionalEnvs) == 0 {
		return envs
	}

	indexedByName := make(map[string]int)
	variables := make([]corev1.EnvVar, len(envs))

	for i, env := range envs {
		indexedByName[env.Name] = i
		variables[i] = env
	}

	for _, env := range additionalEnvs {
		if idx, ok := indexedByName[env.Name]; ok {
			variables[idx] = env
		} else {
			variables = append(variables, env)
		}
	}

	return variables
}
