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
	"crypto/rand"

	"github.com/pkg/errors"
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

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read random bytes")
	}
	return b, nil
}

const lowercaseAlphanumeric = "0123456789abcdefghijklmnopqrstuvwxyz"

func GenerateRandomLowercaseString(n int) (string, error) {
	return generateRandomString(n, lowercaseAlphanumeric)
}

func generateRandomString(n int, letters string) (string, error) {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
