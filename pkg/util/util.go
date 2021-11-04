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

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/imdario/mergo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"istio.io/api/mesh/v1alpha1"
	k8s_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"
)

func CreateLogger(debug bool, development bool) logr.Logger {
	// create encoder config
	var config zapcore.EncoderConfig
	if development {
		config = zap.NewDevelopmentEncoderConfig()
	} else {
		config = zap.NewProductionEncoderConfig()
	}
	// set human readable timestamp format regardless whether development mode is on
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// create the encoder
	var encoder zapcore.Encoder
	if development {
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		encoder = zapcore.NewJSONEncoder(config)
	}

	// set the log level
	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	return k8s_zap.New(k8s_zap.UseDevMode(development), k8s_zap.Encoder(encoder), k8s_zap.Level(level))
}

type MergoOption = func(*mergo.Config)

func MergeMeshConfigs(mergoOptions []MergoOption, meshConfigs ...v1alpha1.MeshConfig) (v1alpha1.MeshConfig, error) {
	m := jsonpb.Marshaler{}
	dstMeshConfig := v1alpha1.MeshConfig{}
	var dstMap map[string]interface{}

	if mergoOptions == nil {
		mergoOptions = make([]MergoOption, 0)
	}

	if len(mergoOptions) == 0 {
		mergoOptions = append(mergoOptions, mergo.WithOverride)
	}

	for _, mc := range meshConfigs {
		mc := mc
		y, err := m.MarshalToString(&mc)
		if err != nil {
			return dstMeshConfig, err
		}

		var sourceMap map[string]interface{}
		err = json.Unmarshal([]byte(y), &sourceMap)
		if err != nil {
			return dstMeshConfig, err
		}
		err = mergo.Merge(&dstMap, sourceMap, mergoOptions...)
		if err != nil {
			return dstMeshConfig, err
		}
	}

	jsonBytes, err := json.Marshal(&dstMap)
	if err != nil {
		return dstMeshConfig, err
	}

	err = jsonpb.UnmarshalString(string(jsonBytes), &dstMeshConfig)
	if err != nil {
		return dstMeshConfig, err
	}

	return dstMeshConfig, nil
}

func MergeYAMLs(mergoOptions []MergoOption, yamls ...string) ([]byte, error) {
	var l map[string]interface{}

	if mergoOptions == nil {
		mergoOptions = make([]func(*mergo.Config), 0)
	}

	if len(mergoOptions) == 0 {
		mergoOptions = append(mergoOptions, mergo.WithOverride)
	}

	for _, y := range yamls {
		var r map[string]interface{}
		err := yaml.Unmarshal([]byte(y), &r)
		if err != nil {
			return nil, err
		}

		err = mergo.Merge(&l, r, mergoOptions...)
		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(l)
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
