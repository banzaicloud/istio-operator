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
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"strings"
	"text/template"

	"emperror.dev/errors"
	"github.com/Masterminds/sprig"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

func TransformStructToStriMapWithTemplate(data interface{}, filesystem fs.FS, templateFileName string) (helm.Strimap, error) {
	tt, err := template.New(path.Base(templateFileName)).Funcs(template.FuncMap{
		"PointerToBool": utils.PointerToBool,
		"toYaml": func(value interface{}) string {
			y, err := yaml.Marshal(value)
			if err != nil {
				return ""
			}

			return string(y)
		},
		"toYamlIf": func(value interface{}) string {
			sprig.TxtFuncMap()
			body := []string{}
			if dict, ok := value.(map[string]interface{}); ok { // nolint:nestif
				if key, ok := dict["key"]; ok {
					body = append(body, fmt.Sprintf("%s:", key))
				}
				if value, ok := dict["value"]; ok {
					if value == nil || reflect.ValueOf(value).IsNil() {
						return ""
					}
					y, err := yaml.Marshal(value)
					if err != nil {
						return ""
					}

					indent := func(spaces int, v string) string {
						pad := strings.Repeat(" ", spaces)

						return pad + strings.ReplaceAll(v, "\n", "\n"+pad)
					}

					content := string(y)
					if len(body) > 0 {
						content = indent(2, content)
					}

					body = append(body, content)
				}
			}

			return strings.Join(body, "\n")
		},
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

func AddFinalizer(c client.Client, obj client.Object, finalizerID string) error {
	finalizers := obj.GetFinalizers()
	if obj.GetDeletionTimestamp().IsZero() && !ContainsString(finalizers, finalizerID) {
		finalizers = append(finalizers, finalizerID)
		obj.SetFinalizers(finalizers)
		if err := c.Update(context.Background(), obj); err != nil {
			return errors.WrapIf(err, "could not add finalizer to resource")
		}
	}

	return nil
}

func RemoveFinalizer(c client.Client, obj client.Object, finalizerID string) error {
	finalizers := obj.GetFinalizers()

	if !obj.GetDeletionTimestamp().IsZero() && ContainsString(finalizers, finalizerID) {
		finalizers = RemoveString(finalizers, finalizerID)
		obj.SetFinalizers(finalizers)
		if err := c.Update(context.Background(), obj); err != nil {
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
