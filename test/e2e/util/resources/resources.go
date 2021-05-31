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
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"emperror.dev/errors"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MustGetResource(t *testing.T, c client.Client, key client.ObjectKey, obj runtime.Object) {
	require.NoError(t, c.Get(context.TODO(), key, obj.(client.Object)))
}

func MustDeleteResources(t *testing.T, c client.Client, filename string) {
	require.NoError(t, DeleteResources(c, filename))
}

func DeleteResources(client client.Client, filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	objects, err := ParseK8sObjectsFromYAMLManifest(string(bytes))
	if err != nil {
		return err
	}

	objects.Sort(deleteObjectOrder())
	for _, obj := range objects {
		err = client.Delete(context.Background(), obj.UnstructuredObject())
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func ApplyResources(client client.Client, filename string) error {
	objects, err := ParseObjects(filename)
	if err != nil {
		return err
	}

	objects.Sort(applyObjectOrder())
	for _, obj := range objects {
		err = client.Create(context.Background(), obj.UnstructuredObject())
		if k8serrors.IsAlreadyExists(err) {
			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func ParseObjects(filename string) (K8sObjects, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseK8sObjectsFromYAMLManifest(string(bytes))
}

func applyObjectOrder() func(o *K8sObject) int {
	var Order = []string{
		"CustomResourceDefinition",
		"Namespace",
		"ResourceQuota",
		"LimitRange",
		"PodSecurityPolicy",
		"PodDisruptionBudget",
		"Secret",
		"ConfigMap",
		"StorageClass",
		"PersistentVolume",
		"PersistentVolumeClaim",
		"ServiceAccount",
		"ClusterRole",
		"ClusterRoleList",
		"ClusterRoleBinding",
		"ClusterRoleBindingList",
		"Role",
		"RoleList",
		"RoleBinding",
		"RoleBindingList",
		"Service",
		"DaemonSet",
		"Pod",
		"ReplicationController",
		"ReplicaSet",
		"Deployment",
		"HorizontalPodAutoscaler",
		"StatefulSet",
		"Job",
		"CronJob",
		"Ingress",
		"APIService",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}

func deleteObjectOrder() func(o *K8sObject) int {
	var Order = []string{
		"APIService",
		"Ingress",
		"Service",
		"CronJob",
		"Job",
		"StatefulSet",
		"HorizontalPodAutoscaler",
		"Deployment",
		"ReplicaSet",
		"ReplicationController",
		"Pod",
		"DaemonSet",
		"RoleBindingList",
		"RoleBinding",
		"RoleList",
		"Role",
		"ClusterRoleBindingList",
		"ClusterRoleBinding",
		"ClusterRoleList",
		"ClusterRole",
		"ServiceAccount",
		"PersistentVolumeClaim",
		"PersistentVolume",
		"StorageClass",
		"ConfigMap",
		"Secret",
		"PodDisruptionBudget",
		"PodSecurityPolicy",
		"LimitRange",
		"ResourceQuota",
		"Policy",
		"Gateway",
		"VirtualService",
		"DestinationRule",
		"Handler",
		"Instance",
		"Rule",
		"Namespace",
		"CustomResourceDefinition",
	}

	order := make(map[string]int, len(Order))
	for i, kind := range Order {
		order[kind] = i
	}

	return func(o *K8sObject) int {
		if nr, ok := order[o.Kind]; ok {
			return nr
		}
		return 1000
	}
}

func MustCreateResources(t *testing.T, c client.Client, file string) {
	require.NoError(t, CreateResources(c, file))
}

func CreateResources(c client.Client, file string) error {
	err := processResources(c, file, createObject)
	if err != nil {
		return errors.WrapIf(err, "unable to create object")
	}
	return nil
}

func createObject(ctx context.Context, c client.Client, obj runtime.Object) error {
	err := c.Create(ctx, obj.(client.Object))
	return errors.WrapIfWithDetails(err, "can't create object", "objectKey", ObjectKey(obj))
}

func MustPatchResources(t *testing.T, c client.Client, file string) {
	require.NoError(t, PatchResources(c, file))
}

func PatchResources(c client.Client, file string) error {
	err := processResources(c, file, patchObject)
	if err != nil {
		return errors.WrapIf(err, "unable to update object")
	}
	return nil
}

func patchObject(ctx context.Context, c client.Client, obj runtime.Object) error {
	return ensureObject(ctx, c, obj, true)
}

func MustCreateOrPatchResources(t *testing.T, c client.Client, file string) {
	require.NoError(t, CreateOrPatchResources(c, file))
}

func CreateOrPatchResources(c client.Client, file string) error {
	err := processResources(c, file, createOrPatchObject)
	if err != nil {
		return errors.WrapIf(err, "unable to create or update object")
	}
	return nil
}

func createOrPatchObject(ctx context.Context, c client.Client, obj runtime.Object) error {
	return ensureObject(ctx, c, obj, false)
}

func processResources(c client.Client, file string, f func(context.Context, client.Client, runtime.Object) error) (err error) {
	objects, err := ParseObjects(file)
	if err != nil {
		return errors.WrapIf(err, "unable to read objects")
	}

	for _, obj := range objects {
		obj := obj.UnstructuredObject()

		err = f(context.TODO(), c, obj)
		if err != nil {
			return errors.WrapIfWithDetails(err, "unable to process object", "objectKey", ObjectKey(obj))
		}
	}

	return nil
}

func ensureObject(ctx context.Context, c client.Client, obj runtime.Object, mustExist bool) error {
	existing := obj.DeepCopyObject().(client.Object)

	objectKey := ObjectKey(existing)
	err := c.Get(ctx, objectKey, existing)
	if err != nil {
		if !mustExist && k8serrors.IsNotFound(err) {
			err := c.Create(ctx, obj.(client.Object))
			return errors.WrapIfWithDetails(err, "can't create object", "objectKey", objectKey)
		}
		return err
	}

	var patchData []byte
	patchData, err = json.Marshal(obj)
	if err != nil {
		return err
	}

	return c.Patch(ctx, existing, client.RawPatch(types.MergePatchType, patchData))
}

func ObjectKey(obj runtime.Object) client.ObjectKey {
	m, err := meta.Accessor(obj)
	if err != nil {
		panic(err)
	}
	return client.ObjectKey{
		Name:      m.GetName(),
		Namespace: m.GetNamespace(),
	}
}
