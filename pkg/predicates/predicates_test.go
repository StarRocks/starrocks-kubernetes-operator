package predicates

import (
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	srapi "github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1"
)

func TestGenericPredicates_Create(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		annotations map[string]string
		envDenyList string
		want        bool
	}{
		{
			name:      "allow object in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:        "deny object in denied namespace",
			namespace:   "kube-system",
			envDenyList: "kube-system",
			want:        false,
		},
		{
			name:        "deny object with ignored annotation",
			namespace:   "default",
			annotations: map[string]string{ignoredAnnotation: "true"},
			want:        false,
		},
		{
			name:        "allow object with ignored annotation set to false",
			namespace:   "default",
			annotations: map[string]string{ignoredAnnotation: "false"},
			want:        true,
		},
		{
			name:        "deny object in one of multiple denied namespaces",
			namespace:   "monitoring",
			envDenyList: "kube-system,monitoring,logging",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envDenyList != "" {
				os.Setenv("DENY_LIST", tt.envDenyList)
				defer os.Unsetenv("DENY_LIST")
			}

			// Create test object
			obj := &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-cluster",
					Namespace:   tt.namespace,
					Annotations: tt.annotations,
				},
			}

			// Create event
			e := event.CreateEvent{
				Object: obj,
			}

			// Test predicate
			gp := GenericPredicates{}
			if got := gp.Create(e); got != tt.want {
				t.Errorf("GenericPredicates.Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenericPredicates_Update(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		annotations map[string]string
		envDenyList string
		want        bool
	}{
		{
			name:      "allow update in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:        "deny update in denied namespace",
			namespace:   "kube-system",
			envDenyList: "kube-system",
			want:        false,
		},
		{
			name:        "deny update with ignored annotation",
			namespace:   "default",
			annotations: map[string]string{ignoredAnnotation: "true"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envDenyList != "" {
				os.Setenv("DENY_LIST", tt.envDenyList)
				defer os.Unsetenv("DENY_LIST")
			}

			// Create test objects
			oldObj := &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: tt.namespace,
				},
			}
			newObj := &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-cluster",
					Namespace:   tt.namespace,
					Annotations: tt.annotations,
				},
			}

			// Create event
			e := event.UpdateEvent{
				ObjectOld: oldObj,
				ObjectNew: newObj,
			}

			// Test predicate
			gp := GenericPredicates{}
			if got := gp.Update(e); got != tt.want {
				t.Errorf("GenericPredicates.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		defaultVal []string
		separator  string
		want       []string
	}{
		{
			name:      "single value",
			envValue:  "namespace1",
			separator: ",",
			want:      []string{"namespace1"},
		},
		{
			name:      "multiple values",
			envValue:  "namespace1,namespace2,namespace3",
			separator: ",",
			want:      []string{"namespace1", "namespace2", "namespace3"},
		},
		{
			name:      "values with spaces",
			envValue:  "namespace1, namespace2 , namespace3",
			separator: ",",
			want:      []string{"namespace1", "namespace2", "namespace3"},
		},
		{
			name:       "empty value returns default",
			envValue:   "",
			defaultVal: []string{"default1", "default2"},
			separator:  ",",
			want:       []string{"default1", "default2"},
		},
		{
			name:      "empty elements are filtered",
			envValue:  "namespace1,,namespace2,",
			separator: ",",
			want:      []string{"namespace1", "namespace2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("TEST_ENV", tt.envValue)
			defer os.Unsetenv("TEST_ENV")

			got := getEnvAsSlice("TEST_ENV", tt.defaultVal, tt.separator)
			if len(got) != len(tt.want) {
				t.Errorf("getEnvAsSlice() returned %d values, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getEnvAsSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestShouldReconcile_WithNilObject(t *testing.T) {
	// Test that nil objects are handled gracefully
	gp := GenericPredicates{}

	// Test Create with nil object
	e1 := event.CreateEvent{Object: nil}
	if got := gp.Create(e1); got != false {
		t.Errorf("GenericPredicates.Create() with nil object = %v, want false", got)
	}

	// Test Delete with nil object
	e2 := event.DeleteEvent{Object: nil}
	if got := gp.Delete(e2); got != false {
		t.Errorf("GenericPredicates.Delete() with nil object = %v, want false", got)
	}

	// Test Generic with nil object
	e3 := event.GenericEvent{Object: nil}
	if got := gp.Generic(e3); got != false {
		t.Errorf("GenericPredicates.Generic() with nil object = %v, want false", got)
	}
}
