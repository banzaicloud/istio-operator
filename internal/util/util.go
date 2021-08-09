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
	"encoding/json"
	"io/fs"
	"path"
	"text/template"

	"emperror.dev/errors"
	"github.com/Masterminds/sprig"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

func TransformICPToStriMapWithTemplate(icp *v1alpha1.IstioControlPlane, filesystem fs.FS, templateFileName string) (helm.Strimap, error) {
	tt, err := template.New(path.Base(templateFileName)).Funcs(template.FuncMap{
		"PointerToBool": utils.PointerToBool,
		"toYaml": func(value interface{}) string {
			y, err := yaml.Marshal(value)
			if err != nil {
				return ""
			}

			return string(y)
		},
	}).Funcs(sprig.TxtFuncMap()).ParseFS(filesystem, templateFileName)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "template cannot be parsed", "template", templateFileName)
	}

	var tpl bytes.Buffer
	err = tt.Execute(&tpl, icp)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "template cannot be executed", "template", templateFileName)
	}

	// fmt.Printf("PY: %s\n", string(tpl.Bytes()))

	values := &helm.Strimap{}
	err = yaml.Unmarshal(tpl.Bytes(), values)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "values string cannot be unmarshalled", tpl.String())
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
