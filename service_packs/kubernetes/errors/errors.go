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
