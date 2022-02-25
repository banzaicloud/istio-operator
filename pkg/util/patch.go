/*
Copyright 2022 Cisco Systems, Inc. and/or its affiliates.

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
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/iancoleman/strcase"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	forkedjson "k8s.io/apimachinery/third_party/forked/golang/json"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

func NewProtoCompatiblePatchMaker(preconditionFuncs ...mergepatch.PreconditionFunc) patch.Maker {
	return patch.NewPatchMaker(patch.DefaultAnnotator, NewProtoCompatibleK8sStrategicMergePatcher(preconditionFuncs...), &patch.BaseJSONMergePatcher{})
}

type ProtoCompatibleK8sStrategicMergePatcher struct {
	patch.K8sStrategicMergePatcher
}

func NewProtoCompatibleK8sStrategicMergePatcher(preconditionFuncs ...mergepatch.PreconditionFunc) patch.StrategicMergePatcher {
	return &ProtoCompatibleK8sStrategicMergePatcher{
		K8sStrategicMergePatcher: patch.K8sStrategicMergePatcher{
			PreconditionFuncs: preconditionFuncs,
		},
	}
}

func (p *ProtoCompatibleK8sStrategicMergePatcher) StrategicMergePatch(original, patch []byte, dataStruct interface{}) ([]byte, error) {
	schema, err := NewPatchMetaFromStruct(dataStruct)
	if err != nil {
		return nil, err
	}

	return strategicpatch.StrategicMergePatchUsingLookupPatchMeta(original, patch, schema)
}

func (p *ProtoCompatibleK8sStrategicMergePatcher) CreateTwoWayMergePatch(original, modified []byte, dataStruct interface{}) ([]byte, error) {
	schema, err := NewPatchMetaFromStruct(dataStruct)
	if err != nil {
		return nil, err
	}

	return strategicpatch.CreateTwoWayMergePatchUsingLookupPatchMeta(original, modified, schema, p.PreconditionFuncs...)
}

func (p *ProtoCompatibleK8sStrategicMergePatcher) CreateThreeWayMergePatch(original, modified, current []byte, dataStruct interface{}) ([]byte, error) {
	lookupPatchMeta, err := NewPatchMetaFromStruct(dataStruct)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "failed to lookup patch meta", "current object", dataStruct)
	}

	return strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true, p.PreconditionFuncs...)
}

type PatchMetaFromStruct struct {
	strategicpatch.PatchMetaFromStruct
}

func NewPatchMetaFromStruct(dataStruct interface{}) (PatchMetaFromStruct, error) {
	pm, err := strategicpatch.NewPatchMetaFromStruct(dataStruct)
	if err != nil {
		return PatchMetaFromStruct{}, err
	}

	return PatchMetaFromStruct{
		PatchMetaFromStruct: pm,
	}, nil
}

func (s PatchMetaFromStruct) LookupPatchMetadataForStruct(key string) (strategicpatch.LookupPatchMeta, strategicpatch.PatchMeta, error) {
	fieldType, fieldPatchStrategies, fieldPatchMergeKey, err := forkedjson.LookupPatchMetadataForStruct(s.PatchMetaFromStruct.T, key)
	if err != nil && strings.Contains(err.Error(), "unable to find api field in struct") {
		// trying with camel or snake case again depending on the original case
		if strings.Contains(key, "_") {
			key = strcase.ToLowerCamel(key)
		} else {
			key = strcase.ToSnake(key)
		}
		fieldType, fieldPatchStrategies, fieldPatchMergeKey, err = forkedjson.LookupPatchMetadataForStruct(s.PatchMetaFromStruct.T, key)
	}
	if err != nil {
		return nil, strategicpatch.PatchMeta{}, err
	}

	pm := strategicpatch.PatchMeta{}
	pm.SetPatchMergeKey(fieldPatchMergeKey)
	pm.SetPatchStrategies(fieldPatchStrategies)

	return PatchMetaFromStruct{PatchMetaFromStruct: strategicpatch.PatchMetaFromStruct{
		T: fieldType,
	}}, pm, nil
}

func (s PatchMetaFromStruct) LookupPatchMetadataForSlice(key string) (strategicpatch.LookupPatchMeta, strategicpatch.PatchMeta, error) {
	subschema, patchMeta, err := s.LookupPatchMetadataForStruct(key)
	if err != nil {
		return nil, strategicpatch.PatchMeta{}, err
	}
	var t reflect.Type
	if elemPatchMetaFromStruct, ok := subschema.(PatchMetaFromStruct); ok {
		t = elemPatchMetaFromStruct.T
	} else {
		return nil, strategicpatch.PatchMeta{}, errors.New("invalid sub schema type")
	}

	var elemType reflect.Type
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		elemType = t.Elem()
		if elemType.Kind() == reflect.Array || elemType.Kind() == reflect.Slice {
			return nil, strategicpatch.PatchMeta{}, errors.New("unexpected slice of slice")
		}
	case reflect.Ptr:
		t = t.Elem()
		if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		elemType = t
	default:
		return nil, strategicpatch.PatchMeta{}, errors.NewWithDetails("expected slice or array type", "type", s.T.Kind().String())
	}

	return PatchMetaFromStruct{
		PatchMetaFromStruct: strategicpatch.PatchMetaFromStruct{
			T: elemType,
		},
	}, patchMeta, nil
}
