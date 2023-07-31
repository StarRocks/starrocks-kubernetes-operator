/*
Copyright 2021-present, StarRocks Inc.

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

package k8sutils

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewFakeClient creates a new fake Kubernetes client.
func NewFakeClient(scheme *runtime.Scheme, initObjs ...runtime.Object) Client {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).WithScheme(scheme).Build()
}

type Client client.Client

var (
	_ Client              = failingClient{}
	_ client.StatusWriter = failingStatusWriter{}
)

type failingClient struct {
	err error
}

// NewFailingClient returns a client that always returns the provided error when called.
func NewFailingClient(err error) Client {
	return failingClient{err: err}
}

func (fc failingClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return fc.err
}

func (fc failingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return fc.err
}

func (fc failingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return fc.err
}

func (fc failingClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return fc.err
}

func (fc failingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return fc.err
}

func (fc failingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return fc.err
}

func (fc failingClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return fc.err
}

func (fc failingClient) Status() client.StatusWriter {
	return failingStatusWriter{err: fc.err} //nolint:gosimple
}

func (fc failingClient) Scheme() *runtime.Scheme {
	return runtime.NewScheme()
}

func (fc failingClient) RESTMapper() meta.RESTMapper {
	return nil
}

type failingStatusWriter struct {
	err error
}

func (fsw failingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return fsw.err
}

func (fsw failingStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return fsw.err
}
