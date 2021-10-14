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

package util

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"path"
	"text/template"

	"emperror.dev/errors"
	"github.com/Masterminds/sprig"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/resources"
)

func TransformStructToStriMapWithTemplate(data interface{}, filesystem fs.FS, templateFileName string) (helm.Strimap, error) {
	t := template.New(path.Base(templateFileName))
	tt, err := t.Funcs(template.FuncMap{
		"include":      includeTemplateFunc(t),
		"toYaml":       toYamlTemplateFunc,
		"fromYaml":     fromYamlTemplateFunc,
		"fromJson":     fromJSONTemplateFunc,
		"valueIf":      valueIfTemplateFunc,
		"reformatYaml": reformatYamlTemplateFunc,
		"toYamlIf":     toYamlIfTemplateFunc,
		"toJsonPB":     toJSONPBTemplateFunc,
	}).Funcs(sprig.TxtFuncMap()).ParseFS(filesystem, templateFileName)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "template cannot be parsed", "template", templateFileName)
	}

	var tpl bytes.Buffer
	err = tt.Execute(&tpl, data)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "template cannot be executed", "template", templateFileName)
	}

	values := &helm.Strimap{}
	err = yaml.Unmarshal(tpl.Bytes(), values)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "values string cannot be unmarshalled", "template", tpl.String())
	}

	return *values, nil
}

func ProtoFieldToStriMap(protoField proto.Message, striMap *helm.Strimap) error {
	marshaller := jsonpb.Marshaler{}
	stringField, err := marshaller.MarshalToString(protoField)
	if err != nil {
		return errors.Errorf("proto field cannot be converted into string: %+v", protoField)
	}

	err = json.Unmarshal([]byte(stringField), striMap)
	if err != nil {
		return errors.Errorf("proto field cannot be converted into map[string]interface{}: %+v", protoField)
	}

	return nil
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

func AddFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizerID string) error {
	finalizers := obj.GetFinalizers()
	if obj.GetDeletionTimestamp().IsZero() && !ContainsString(finalizers, finalizerID) {
		finalizers = append(finalizers, finalizerID)
		obj.SetFinalizers(finalizers)
		if err := c.Update(ctx, obj); err != nil {
			return errors.WrapIf(err, "could not add finalizer to resource")
		}
	}

	return nil
}

func RemoveFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizerID string, onDeleteOnly bool) error {
	finalizers := obj.GetFinalizers()

	if onDeleteOnly && obj.GetDeletionTimestamp().IsZero() {
		return nil
	}

	if ContainsString(finalizers, finalizerID) {
		finalizers = RemoveString(finalizers, finalizerID)
		obj.SetFinalizers(finalizers)
		if err := c.Update(ctx, obj); err != nil {
			return errors.WrapIf(err, "could not remove finalizer from resource")
		}
	}

	return nil
}

func CompareYAMLs(left, right []byte) (dyff.Report, error) {
	y1, err := ytbx.LoadDocuments(left)
	if err != nil {
		return dyff.Report{}, err
	}
	y2, err := ytbx.LoadDocuments(right)
	if err != nil {
		return dyff.Report{}, err
	}

	return dyff.CompareInputFiles(ytbx.InputFile{
		Location:  "left",
		Documents: y1,
	}, ytbx.InputFile{
		Location:  "right",
		Documents: y2,
	},
		dyff.IgnoreOrderChanges(false),
		dyff.KubernetesEntityDetection(true),
	)
}

func ConvertK8sOverlays(overlays []*v1alpha1.K8SResourceOverlayPatch) ([]resources.K8SResourceOverlay, error) {
	var o []resources.K8SResourceOverlay

	j, err := json.Marshal(overlays)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &o)
	if err != nil {
		return nil, err
	}

	return o, nil
}
