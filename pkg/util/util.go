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

package util

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func StrPointer(s string) *string {
	return &s
}

func IntPointer(i int32) *int32 {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}

func BoolPointer(b bool) *bool {
	return &b
}

func PointerToBool(flag *bool) bool {
	if flag == nil {
		return false
	}

	return *flag
}

func PointerToString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func PointerToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}

	return *i
}

func IntstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}

func MergeStringMaps(l map[string]string, l2 map[string]string) map[string]string {
	merged := make(map[string]string)
	if l == nil {
		l = make(map[string]string)
	}
	for lKey, lValue := range l {
		merged[lKey] = lValue
	}
	for lKey, lValue := range l2 {
		merged[lKey] = lValue
	}
	return merged
}

func MergeMultipleStringMaps(stringMaps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, stringMap := range stringMaps {
		merged = MergeStringMaps(merged, stringMap)
	}
	return merged
}

func EmptyTypedStrSlice(s ...string) []interface{} {
	ret := make([]interface{}, len(s))
	for i := 0; i < len(s); i++ {
		ret[i] = s[i]
	}
	return ret
}

func EmptyTypedFloatSlice(f ...float64) []interface{} {
	ret := make([]interface{}, len(f))
	for i := 0; i < len(f); i++ {
		ret[i] = f[i]
	}
	return ret
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func GetPodSecurityContextFromSecurityContext(sc *corev1.SecurityContext) *corev1.PodSecurityContext {
	if sc == nil || *sc == (corev1.SecurityContext{}) {
		return &corev1.PodSecurityContext{}
	}
	return &corev1.PodSecurityContext{
		RunAsGroup:   sc.RunAsGroup,
		RunAsNonRoot: sc.RunAsNonRoot,
		RunAsUser:    sc.RunAsUser,
		FSGroup:      sc.RunAsGroup,
	}
}
