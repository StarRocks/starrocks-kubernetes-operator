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

package fake

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

func (fc failingClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return fc.err
}

func (fc failingClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return fc.err
}

func (fc failingClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return fc.err
}

func (fc failingClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return fc.err
}

func (fc failingClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return fc.err
}

func (fc failingClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return fc.err
}

func (fc failingClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
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

func (fsw failingStatusWriter) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return fsw.err
}

func (fsw failingStatusWriter) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return fsw.err
}
