/*
 * Copyright 2021-present, StarRocks Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package statefulset

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

func TestStatus_OnDelete(t *testing.T) {
	replicas := int32(3)
	sts := func(observed int64, ready, updated int32) *appsv1.StatefulSet {
		s := &appsv1.StatefulSet{}
		s.Generation = 2
		s.Spec.Replicas = &replicas
		s.Spec.UpdateStrategy = appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}
		s.Status.ObservedGeneration = observed
		s.Status.ReadyReplicas = ready
		s.Status.UpdatedReplicas = updated
		s.Status.UpdateRevision = "rev-new"
		s.Status.CurrentRevision = "rev-old" // never advanced by the StatefulSet controller under OnDelete
		return s
	}

	_, done, err := Status(sts(1, 3, 0))
	require.NoError(t, err)
	require.False(t, done, "spec not observed yet")

	_, done, err = Status(sts(2, 2, 1))
	require.NoError(t, err)
	require.False(t, done, "a pod is not ready")

	msg, done, err := Status(sts(2, 3, 1))
	require.NoError(t, err)
	require.False(t, done, "pods still on the old revision")
	require.Contains(t, msg, "1 out of 3")

	_, done, err = Status(sts(2, 3, 3))
	require.NoError(t, err)
	require.True(t, done, "all pods on the update revision")
}

func TestStatus_UnknownStrategy(t *testing.T) {
	s := &appsv1.StatefulSet{}
	s.Spec.UpdateStrategy.Type = "Bogus"
	_, _, err := Status(s)
	require.Error(t, err)
}
