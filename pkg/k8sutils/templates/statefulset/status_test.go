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

	appsv1 "k8s.io/api/apps/v1"
)

func TestStatus_OnDelete(t *testing.T) {
	replicas := int32(3)

	tests := []struct {
		name     string
		sts      *appsv1.StatefulSet
		wantDone bool
		wantErr  bool
	}{
		{
			name: "OnDelete: revisions differ but all replicas ready - should not error, done=true",
			sts: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas:       &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				},
				Status: appsv1.StatefulSetStatus{
					ObservedGeneration: 1,
					ReadyReplicas:      3,
					Replicas:           3,
					UpdatedReplicas:    1,
					CurrentRevision:    "v1",
					UpdateRevision:     "v2",
				},
			},
			wantDone: true,
			wantErr:  false,
		},
		{
			name: "OnDelete: not all replicas ready yet",
			sts: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas:       &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				},
				Status: appsv1.StatefulSetStatus{
					ObservedGeneration: 1,
					ReadyReplicas:      2,
				},
			},
			wantDone: false,
			wantErr:  false,
		},
		{
			name: "OnDelete: spec update not yet observed",
			sts: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas:       &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				},
				Status: appsv1.StatefulSetStatus{
					ObservedGeneration: 0,
				},
			},
			wantDone: false,
			wantErr:  false,
		},
		{
			name: "OnDelete: fully rolled out, revisions match",
			sts: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas:       &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				},
				Status: appsv1.StatefulSetStatus{
					ObservedGeneration: 1,
					ReadyReplicas:      3,
					CurrentReplicas:    3,
					UpdatedReplicas:    3,
					CurrentRevision:    "v2",
					UpdateRevision:     "v2",
				},
			},
			wantDone: true,
			wantErr:  false,
		},
		{
			name: "unsupported strategy type still errors",
			sts: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas:       &replicas,
					UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: "Unknown"},
				},
			},
			wantDone: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, done, err := Status(tt.sts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Status() error = %v, wantErr %v", err, tt.wantErr)
			}
			if done != tt.wantDone {
				t.Fatalf("Status() done = %v, want %v", done, tt.wantDone)
			}
		})
	}
}
