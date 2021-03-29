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

package client

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options are creation options for a Client
type Options struct {
	// Scheme, if provided, will be used to map go structs to GroupVersionKinds
	Scheme *runtime.Scheme

	// Mapper, if provided, will be used to map GroupVersionKinds to Resources
	Mapper meta.RESTMapper
}

// NewClient returns a new Client using the provided config and Options.
// The returned client reads *and* writes directly from the server
// (it doesn't use object caches).  It understands how to work with
// normal types (both custom resources and aggregated/built-in resources),
// as well as unstructured types.
//
// In the case of normal types, the scheme will be used to look up the
// corresponding group, version, and kind for the given type.  In the
// case of unstructured types, the group, version, and kind will be extracted
// from the corresponding fields on the object.
func NewClient(config *rest.Config, options Options) (client.Client, error) {
	if options.Scheme == nil {
		options.Scheme = GetScheme()
	}
	c, err := client.New(config, client.Options{
		Scheme: options.Scheme,
		Mapper: options.Mapper,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}

func NewClientFromKubeConfigAndContext(kubeConfigPath, kubeContext string) (client.Client, error) {
	config, err := GetConfigWithContext(kubeConfigPath, kubeContext)
	if err != nil {
		return nil, err
	}

	c, err := NewClient(config, Options{})
	if err != nil {
		return nil, err
	}

	return c, nil
}
