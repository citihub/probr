package coreengine

import (
	"path/filepath"
	"testing"
)

func TestGetProbesPath(t *testing.T) {
	var failed bool
	r, _ := getRootDir()
	desired_path := filepath.Join(r, "probes", "clouddriver", "probe_definitions", "accountmanager")

	// Test with feature path provided
	p := filepath.Join("probes", "clouddriver", "probe_definitions", "accountmanager")
	test := &GodogProbe{FeaturePath: &p}
	path, err := getProbesPath(test)
	if err != nil || desired_path != path {
		t.Logf("Custom feature path not handled properly")
		failed = true
	}

	// Test building path from properties
	test = &GodogProbe{ProbeDescriptor: &ProbeDescriptor{Group: CloudDriver, Name: "account_manager"}}
	path, err = getProbesPath(test)
	if err != nil || desired_path != path {
		t.Logf("Failed to build probe path from GodogProbe properties")
		failed = true
	}

	// Allow both failures to log before ending, if applicable
	if failed {
		t.Fail()
	}
}
