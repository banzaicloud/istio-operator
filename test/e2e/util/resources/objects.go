/*
Copyright 2021 Banzai Cloud.

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

package resources

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// K8sObject is an in-memory representation of a k8s object, used for moving between different representations
// (Unstructured, JSON, YAML) with cached rendering.
type K8sObject struct {
	object *unstructured.Unstructured

	Group     string
	Kind      string
	Name      string
	Namespace string

	json []byte
	yaml []byte
}

// NewK8sObject creates a new K8sObject and returns a ptr to it.
func NewK8sObject(u *unstructured.Unstructured, json, yaml []byte) *K8sObject {
	o := &K8sObject{
		object: u,
		json:   json,
		yaml:   yaml,
	}

	gvk := u.GetObjectKind().GroupVersionKind()
	o.Group = gvk.Group
	o.Kind = gvk.Kind
	o.Name = u.GetName()
	o.Namespace = u.GetNamespace()

	return o
}

// ParseYAMLToK8sObject parses YAML to an Object.
func ParseYAMLToK8sObject(yaml []byte) (*K8sObject, error) {
	r := bytes.NewReader(yaml)
	decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)

	out := &unstructured.Unstructured{}
	err := decoder.Decode(out)
	if err != nil {
		return nil, fmt.Errorf("error decoding object: %v", err)
	}
	return NewK8sObject(out, nil, yaml), nil
}

// UnstructuredObject exposes the raw object, primarily for testing
func (o *K8sObject) UnstructuredObject() *unstructured.Unstructured {
	return o.object
}

// GroupKind returns the GroupKind for the K8sObject
func (o *K8sObject) GroupKind() schema.GroupKind {
	return o.object.GroupVersionKind().GroupKind()
}

// GroupVersionKind returns the GroupVersionKind for the K8sObject
func (o *K8sObject) GroupVersionKind() schema.GroupVersionKind {
	return o.object.GroupVersionKind()
}

// K8sObjects holds a collection of k8s objects, so that we can filter / sequence them
type K8sObjects []*K8sObject

// ParseK8sObjectsFromYAMLManifest returns a K8sObjects representation of manifest.
func ParseK8sObjectsFromYAMLManifest(manifest string) (K8sObjects, error) {
	var b bytes.Buffer

	var yamls []string
	scanner := bufio.NewScanner(strings.NewReader(manifest))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			// yaml separator
			yamls = append(yamls, b.String())
			b.Reset()
		} else {
			if _, err := b.WriteString(line); err != nil {
				return nil, err
			}
			if _, err := b.WriteString("\n"); err != nil {
				return nil, err
			}
		}
	}
	yamls = append(yamls, b.String())

	objects := make(K8sObjects, 0, len(yamls))

	for _, yaml := range yamls {
		yaml = removeNonYAMLLines(yaml)
		if yaml == "" {
			continue
		}
		o, err := ParseYAMLToK8sObject([]byte(yaml))
		if err != nil {
			return nil, errors.WrapIf(err, "Failed to parse YAML to a k8s object")
		}

		objects = append(objects, o)
	}

	return objects, nil
}

func removeNonYAMLLines(yms string) string {
	out := ""
	for _, s := range strings.Split(yms, "\n") {
		if strings.HasPrefix(s, "#") {
			continue
		}
		out += s + "\n"
	}

	// helm charts sometimes emits blank objects with just a "disabled" comment.
	return strings.TrimSpace(out)
}

// Sort will order the items in K8sObjects in order of score, group, kind, name.  The intent is to
// have a deterministic ordering in which K8sObjects are applied.
func (os K8sObjects) Sort(score func(o *K8sObject) int) {
	sort.Slice(os, func(i, j int) bool {
		iScore := score(os[i])
		jScore := score(os[j])
		return iScore < jScore ||
			(iScore == jScore &&
				os[i].Group < os[j].Group) ||
			(iScore == jScore &&
				os[i].Group == os[j].Group &&
				os[i].Kind < os[j].Kind) ||
			(iScore == jScore &&
				os[i].Group == os[j].Group &&
				os[i].Kind == os[j].Kind &&
				os[i].Name < os[j].Name)
	})
}
