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

package pod

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCount(t *testing.T) {
	type args struct {
		podList v1.PodList
	}
	tests := []struct {
		name         string
		args         args
		wantCreating []string
		wantReady    []string
		wantFailed   []string
	}{
		{
			name: "test",
			args: args{
				podList: v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test1",
							},
							Status: v1.PodStatus{
								Phase:  v1.PodRunning,
								Reason: "test1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test2",
							},
							Status: v1.PodStatus{
								ContainerStatuses: []v1.ContainerStatus{
									{
										Ready: true,
									},
								},
								Reason: "test2",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test3",
							},
							Status: v1.PodStatus{
								Phase:  v1.PodFailed,
								Reason: "test3",
							},
						},
					},
				},
			},
			wantCreating: []string{"test1"},
			wantReady:    []string{"test2"},
			wantFailed:   []string{"test3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreating, gotReady, gotFailed := Count(tt.args.podList)
			if !reflect.DeepEqual(gotCreating, tt.wantCreating) {
				t.Errorf("Count() gotCreating = %v, want %v", gotCreating, tt.wantCreating)
			}
			if !reflect.DeepEqual(gotReady, tt.wantReady) {
				t.Errorf("Count() gotReady = %v, want %v", gotReady, tt.wantReady)
			}
			if !reflect.DeepEqual(gotFailed, tt.wantFailed) {
				t.Errorf("Count() gotFailed = %v, want %v", gotFailed, tt.wantFailed)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	type args struct {
		podList v1.PodList
	}
	tests := []struct {
		name string
		args args
		want map[string]PodStatus
	}{
		{
			name: "test",
			args: args{
				podList: v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test1",
							},
							Status: v1.PodStatus{
								Phase:  v1.PodRunning,
								Reason: "test1",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test2",
							},
							Status: v1.PodStatus{
								Phase:  v1.PodSucceeded,
								Reason: "test2",
							},
						},
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test3",
							},
							Status: v1.PodStatus{
								Phase:  v1.PodFailed,
								Reason: "test3",
							},
						},
					},
				},
			},
			want: map[string]PodStatus{
				"test1": {
					Phase:  v1.PodRunning,
					Reason: "test1",
				},
				"test2": {
					Phase:  v1.PodSucceeded,
					Reason: "test2",
				},
				"test3": {
					Phase:  v1.PodFailed,
					Reason: "test3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Status(tt.args.podList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}
