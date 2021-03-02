package errors

import "k8s.io/apimachinery/pkg/api/errors"

// IsStatusCode409 checks that err corresponds to status code 409
func IsStatusCode409(err error) bool {
	if se, ok := err.(*errors.StatusError); ok {
		//409 is "already exists"
		return se.ErrStatus.Code == 409
	}
	return false
}

// IsStatusCode403 checks that err corresponds to status code 409
func IsStatusCode403(err error) bool {
	if se, ok := err.(*errors.StatusError); ok {
		//403 is "forbidden"
		return se.ErrStatus.Code == 403
	}
	return false
}

func IsStatusCode(expected int32, err error) bool {
	if se, ok := err.(*errors.StatusError); ok {
		//403 is "forbidden"
		return se.ErrStatus.Code == expected
	}
	return false
}

// PodCreationErrors gives a list of known pod creation errors
func PodCreationErrors() []string {
	return []string{"podcreation-error: undefined",
		"podcreation-error: psp-container-no-privilege",
		"podcreation-error: psp-container-no-privilege-escalation",
		"podcreation-error: psp-allowed-users-groups",
		"podcreation-error: psp-container-allowed-images",
		"podcreation-error: psp-host-namespace",
		"podcreation-error: psp-host-network",
		"podcreation-error: psp-allowed-capabilities",
		"podcreation-error: psp-allowed-portrange",
		"podcreation-error: psp-allowed-volume-types-profile",
		"podcreation-error: psp-allowed-seccomp-profile",
		"podcreation-error: image-pull-error",
		"podcreation-error: blocked"}
}
