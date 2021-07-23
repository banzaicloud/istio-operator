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
	"fmt"
	"text/template"

	"emperror.dev/errors"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/banzaicloud/istio-operator/v2/api/v1alpha1"
	"github.com/banzaicloud/operator-tools/pkg/helm"
)

func TransformICPSpecToStriMapWithTemplate(spec v1alpha1.IstioControlPlaneSpec, valuesTemplatePath string, valuesTemplateFileName string) (helm.Strimap, error) {
	valuesTemplateFilePath := fmt.Sprintf("%s/%s", valuesTemplatePath, valuesTemplateFileName)
	tmpl, err := template.New(valuesTemplateFileName).ParseFiles(valuesTemplateFilePath)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "values template cannot be parsed", valuesTemplateFilePath)
	}
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, spec)
	if err != nil {
		return nil, errors.WrapWithDetails(err, "values template cannot be applied to ICP spec", valuesTemplateFilePath)
	}

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
