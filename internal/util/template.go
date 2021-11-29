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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/golang/protobuf/proto"
	"github.com/waynz0r/protobuf/jsonpb"
	"sigs.k8s.io/yaml"
)

func includeTemplateFunc(t *template.Template) interface{} {
	return func(name string, data interface{}) (string, error) {
		var buf strings.Builder
		err := t.ExecuteTemplate(&buf, name, data)

		return buf.String(), err
	}
}

func toYamlTemplateFunc(value interface{}) (string, error) {
	y, err := yaml.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(y), nil
}

func toJSONPBTemplateFunc(value interface{}) (string, error) {
	if value == nil || reflect.ValueOf(value).IsZero() {
		return "", nil
	}

	if v, ok := value.(proto.Message); ok {
		m := jsonpb.Marshaler{}
		y, err := m.MarshalToString(v)
		if err != nil {
			return "", err
		}

		return y, nil
	}

	return "", nil
}

func fromJSONTemplateFunc(value string) (map[string]interface{}, error) {
	var out map[string]interface{}
	err := json.Unmarshal([]byte(value), &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func fromYamlTemplateFunc(value string) (map[string]interface{}, error) {
	var out map[string]interface{}
	err := yaml.Unmarshal([]byte(value), &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func valueIfTemplateFunc(value interface{}) (string, error) {
	if dict, ok := value.(map[string]interface{}); ok { // nolint:nestif
		var value interface{}
		var key string

		if value, ok = dict["value"]; !ok {
			return "", nil
		}

		if key, ok = dict["key"].(string); !ok {
			return "", nil
		}

		if value == nil || reflect.ValueOf(value).IsZero() {
			return "", nil
		}

		if key == "" {
			return "", nil
		}

		y, err := yaml.Marshal(value)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s: %s", key, y), nil
	}

	return "", nil
}

func reformatYamlTemplateFunc(value interface{}) (string, error) {
	var m map[string]interface{}
	if v, ok := value.(string); ok {
		err := yaml.Unmarshal([]byte(v), &m)
		if err != nil {
			return "", err
		}

		if m == nil {
			return "", nil
		}

		c, err := yaml.Marshal(m)
		if err != nil {
			return "", err
		}

		return string(c), nil
	}

	return "", nil
}

func toYamlIfTemplateFunc(value interface{}) (string, error) {
	sprig.TxtFuncMap()
	body := []string{}
	if dict, ok := value.(map[string]interface{}); ok { // nolint:nestif
		if key, ok := dict["key"]; ok {
			body = append(body, fmt.Sprintf("%s:", key))
		}
		if value, ok := dict["value"]; ok {
			if value == nil || reflect.ValueOf(value).IsZero() {
				return "", nil
			}
			y, err := yaml.Marshal(value)
			if err != nil {
				return "", err
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

	return strings.Join(body, "\n"), nil
}
