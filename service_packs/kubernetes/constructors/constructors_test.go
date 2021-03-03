// Package constructors provides functions to prepare new objects (as described by the name of the function)

// This implements factory pattern.

package constructors

import (
	"reflect"
	"strings"
	"testing"

	apiv1 "k8s.io/api/core/v1"
)

func Test_uniquePodName(t *testing.T) {
	tests := []struct {
		testName string
		arg      string
	}{
		{
			testName: "Unique Pod Name Contains Base Name",
			arg:      "basename1",
		},
		{
			testName: "Unique Pod Name Contains Base Name",
			arg:      "base-name-2",
		},
		{
			testName: "Unique Pod Name Contains Base Name",
			arg:      "base_name_3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := uniquePodName(tt.arg)
			if !strings.Contains(got, tt.arg) || len(got) <= len(tt.arg) {
				t.Errorf("uniquePodName() = %v, want %v", got, tt.arg)
			}
		})
	}
}

func TestPodSpec(t *testing.T) {
	type args struct {
		baseName                 string
		namespace                string
		containerSecurityContext *apiv1.SecurityContext
	}
	tests := []struct {
		name string
		args args
		want func(gotPod *apiv1.Pod, args args, t *testing.T)
	}{
		{
			name: "Pod's security context is always the default security context",
			args: args{
				baseName:                 "pod1",
				namespace:                "pod1",
				containerSecurityContext: nil,
			},
			want: func(gotPod *apiv1.Pod, args args, t *testing.T) {
				if !reflect.DeepEqual(gotPod.Spec.SecurityContext, DefaultPodSecurityContext()) {
					t.Errorf("PodSpec() should set the pod's security context using DefaultPodSecurityContext()")
				}
			},
		},
		{
			name: "Container security context uses provided value",
			args: args{
				baseName:                 "pod2",
				namespace:                "pod2",
				containerSecurityContext: DefaultContainerSecurityContext(),
			},
			want: func(gotPod *apiv1.Pod, args args, t *testing.T) {
				gotContainerSC := gotPod.Spec.Containers[0].SecurityContext
				if !reflect.DeepEqual(gotContainerSC, args.containerSecurityContext) {
					t.Errorf("PodSpec() set container security context to %v, wanted %v", gotContainerSC, args.containerSecurityContext)
				}
			},
		},
		{
			name: "Container uses a unique pod name",
			args: args{
				baseName:                 "pod3",
				namespace:                "pod3",
				containerSecurityContext: nil,
			},
			want: func(gotPod *apiv1.Pod, args args, t *testing.T) {
				gotContainerName := gotPod.Spec.Containers[0].Name
				if len(gotContainerName) <= len(args.baseName) || !strings.Contains(gotContainerName, args.baseName) {
					t.Errorf("PodSpec() should use uniquePodName() to create name, but instead got: '%s' from arg '%s'", gotContainerName, args.baseName)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PodSpec(tt.args.baseName, tt.args.namespace, tt.args.containerSecurityContext)
			tt.want(got, tt.args, t)
		})
	}
}
