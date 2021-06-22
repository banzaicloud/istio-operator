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

	"emperror.dev/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/banzaicloud/operator-tools/pkg/helm"
)

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
