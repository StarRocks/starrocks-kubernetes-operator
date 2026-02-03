package predicates

import (
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
		denyList    string
		want        bool
	}{
		{
			name:      "allow object in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:      "deny object in denied namespace",
			namespace: "kube-system",
			denyList:  "kube-system",
			want:      false,
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
			name:      "deny object in one of multiple denied namespaces",
			namespace: "monitoring",
			denyList:  "kube-system,monitoring,logging",
			want:      false,
		},
		{
			name:      "allow object when namespace not in deny list",
			namespace: "production",
			denyList:  "kube-system,monitoring",
			want:      true,
		},
		{
			name:      "handle deny list with spaces",
			namespace: "monitoring",
			denyList:  "kube-system, monitoring , logging",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Test predicate with constructor
			gp := NewGenericPredicates(tt.denyList)
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
		denyList    string
		want        bool
	}{
		{
			name:      "allow update in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:      "deny update in denied namespace",
			namespace: "kube-system",
			denyList:  "kube-system",
			want:      false,
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

			// Test predicate with constructor
			gp := NewGenericPredicates(tt.denyList)
			if got := gp.Update(e); got != tt.want {
				t.Errorf("GenericPredicates.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenericPredicates_Delete(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		denyList  string
		want      bool
	}{
		{
			name:      "allow delete in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:      "deny delete in denied namespace",
			namespace: "kube-system",
			denyList:  "kube-system",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: tt.namespace,
				},
			}

			e := event.DeleteEvent{
				Object: obj,
			}

			gp := NewGenericPredicates(tt.denyList)
			if got := gp.Delete(e); got != tt.want {
				t.Errorf("GenericPredicates.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenericPredicates_Generic(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		denyList  string
		want      bool
	}{
		{
			name:      "allow generic in allowed namespace",
			namespace: "default",
			want:      true,
		},
		{
			name:      "deny generic in denied namespace",
			namespace: "kube-system",
			denyList:  "kube-system",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &srapi.StarRocksCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: tt.namespace,
				},
			}

			e := event.GenericEvent{
				Object: obj,
			}

			gp := NewGenericPredicates(tt.denyList)
			if got := gp.Generic(e); got != tt.want {
				t.Errorf("GenericPredicates.Generic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGenericPredicates(t *testing.T) {
	tests := []struct {
		name     string
		denyList string
		wantLen  int
	}{
		{
			name:     "empty deny list",
			denyList: "",
			wantLen:  0,
		},
		{
			name:     "single namespace",
			denyList: "kube-system",
			wantLen:  1,
		},
		{
			name:     "multiple namespaces",
			denyList: "kube-system,monitoring,logging",
			wantLen:  3,
		},
		{
			name:     "namespaces with spaces",
			denyList: "kube-system, monitoring , logging",
			wantLen:  3,
		},
		{
			name:     "empty elements filtered",
			denyList: "kube-system,,monitoring,",
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gp := NewGenericPredicates(tt.denyList)
			if len(gp.denyList) != tt.wantLen {
				t.Errorf("NewGenericPredicates() denyList len = %d, want %d", len(gp.denyList), tt.wantLen)
			}
		})
	}
}

func TestShouldReconcile_WithNilObject(t *testing.T) {
	// Test that nil objects are handled gracefully
	gp := NewGenericPredicates("")

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
