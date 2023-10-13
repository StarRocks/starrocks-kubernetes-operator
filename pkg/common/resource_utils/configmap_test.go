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

package resource_utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestResolveConfigMap(t *testing.T) {
	configMap := corev1.ConfigMap{
		Data: map[string]string{
			"fe.conf": "http_port = 8030",
		},
	}
	res, err := ResolveConfigMap(&configMap, "fe.conf")
	require.NoError(t, err)

	_, ok := res["http_port"]
	require.Equal(t, true, ok)
}
