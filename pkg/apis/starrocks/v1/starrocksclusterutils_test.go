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
 * limitations under the License.
 */

package v1

import "testing"

func TestGetFeExternalServiceName(t *testing.T) {
	type args struct {
		clusterName string
		spec        SpecInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// NOTE: we must not input a nil value for spec, otherwise the following error will occur:
		// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
		// {
		//	name: "test1",
		//	args: args{
		//		clusterName: "test",
		//		spec:        (*StarRocksFeSpec)(nil),
		//	},
		//	want: "test-fe-service",
		// },
		{
			name: "test2",
			args: args{
				clusterName: "test",
				spec:        &StarRocksFeSpec{},
			},
			want: "test-fe-service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExternalServiceName(tt.args.clusterName, tt.args.spec); got != tt.want {
				t.Errorf("GetExternalServiceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
