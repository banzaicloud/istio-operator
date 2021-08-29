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

package v1alpha1

import (
	"strconv"

	"github.com/gogo/protobuf/jsonpb"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// define new type from k8s quantity to marshal/unmarshal jsonpb
type Quantity struct {
	resource.Quantity
}

// MarshalJSONPB implements the jsonpb.JSONPBMarshaler interface.
func (q *Quantity) MarshalJSONPB(_ *jsonpb.Marshaler) ([]byte, error) {
	return q.Quantity.MarshalJSON()
}

// UnmarshalJSONPB implements the jsonpb.JSONPBUnmarshaler interface.
func (q *Quantity) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, value []byte) error {
	// If its a string that isnt wrapped in quotes add them to appease kubernetes unmarshal
	if _, err := strconv.Atoi(string(value)); err != nil && len(value) > 0 && value[0] != '"' {
		value = append([]byte{'"'}, value...)
		value = append(value, '"')
	}

	return q.Quantity.UnmarshalJSON(value)
}

// define new type from k8s intstr to marshal/unmarshal jsonpb
type IntOrString struct {
	intstr.IntOrString
}

// MarshalJSONPB implements the jsonpb.JSONPBMarshaler interface.
func (intstrpb *IntOrString) MarshalJSONPB(_ *jsonpb.Marshaler) ([]byte, error) {
	return intstrpb.IntOrString.MarshalJSON()
}

// UnmarshalJSONPB implements the jsonpb.JSONPBUnmarshaler interface.
func (intstrpb *IntOrString) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, value []byte) error {
	// If its a string that isnt wrapped in quotes add them to appease kubernetes unmarshal
	if _, err := strconv.Atoi(string(value)); err != nil && len(value) > 0 && value[0] != '"' {
		value = append([]byte{'"'}, value...)
		value = append(value, '"')
	}
	return intstrpb.IntOrString.UnmarshalJSON(value)
}

// FromInt creates an IntOrStringForPB object with an int32 value.
func FromInt(val int) IntOrString {
	return IntOrString{intstr.FromInt(val)}
}

// FromString creates an IntOrStringForPB object with a string value.
func FromString(val string) IntOrString {
	return IntOrString{intstr.FromString(val)}
}

func (m *PodDisruptionBudget) GetMinAvailable() *IntOrString {
	if m != nil {
		return m.MinAvailable
	}
	return nil
}

func (m *PodDisruptionBudget) GetMaxUnavailable() *IntOrString {
	if m != nil {
		return m.MaxUnavailable
	}
	return nil
}
