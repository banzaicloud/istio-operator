package util

import "k8s.io/apimachinery/pkg/util/intstr"

func StrPointer(s string) *string {
	return &s
}

func IntPointer(i int32) *int32 {
	return &i
}

func BoolPointer(b bool) *bool {
	return &b
}

func IntstrPointer(i int) *intstr.IntOrString {
	is := intstr.FromInt(i)
	return &is
}

func MergeLabels(l map[string]string, l2 map[string]string) map[string]string {
	for lKey, lValue := range l2 {
		l[lKey] = lValue
	}
	return l
}
