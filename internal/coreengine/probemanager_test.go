package coreengine

import (
	"testing"

	"github.com/citihub/probr/internal/config"
)

func createProbeObj(name string) *GodogProbe {
	return &GodogProbe{
		ProbeDescriptor: &ProbeDescriptor{
			Name:  name,
			Group: Kubernetes,
		},
	}
}

func TestNewProbeStore(t *testing.T) {
	ts := NewProbeStore()
	if ts == nil {
		t.Logf("Probe store was not initialized")
		t.Fail()
	} else if ts.Probes == nil {
		t.Logf("Probe store was not ready to add probes")
		t.Fail()
	}
}

func TestTagIsExcluded(t *testing.T) {
	config.Vars.TagExclusions = []string{"tag_name"}
	if tagIsExcluded("not_tag_name") {
		t.Logf("Non-excluded tag was excluded")
		t.Fail()
	}
	if !tagIsExcluded("tag_name") {
		t.Logf("Excluded tag was not excluded")
		t.Fail()
	}
}

func TestIsExcluded(t *testing.T) {
	config.Vars.TagExclusions = []string{"excluded_probe"}
	pd := ProbeDescriptor{Group: Kubernetes, Name: "good_probe"}
	pd_excluded := ProbeDescriptor{Group: Kubernetes, Name: "excluded_probe"}

	if pd.isExcluded() {
		t.Logf("Non-excluded probe was excluded")
		t.Fail()
	}
	if !pd_excluded.isExcluded() {
		t.Logf("Excluded probe was not excluded")
		t.Fail()
	}
}

func TestAddProbe(t *testing.T) {
	probe := "test probe"
	excluded_probe := "different test probe"
	config.Vars.TagExclusions = []string{excluded_probe}
	ps := NewProbeStore()
	probe_obj := createProbeObj(probe)
	excluded_probe_obj := createProbeObj(excluded_probe)
	ps.AddProbe(probe_obj)
	ps.AddProbe(excluded_probe_obj)

	// Verify correct conditions succeed
	if ps.Probes[probe] == nil {
		t.Logf("Probe not added to probe store")
		t.Fail()
	} else if ps.Probes[probe].ProbeDescriptor.Name != probe {
		t.Logf("Probe name not set properly in test store")
		t.Fail()
	}

	// Verify probe1 and probe2 are different
	if ps.Probes[probe] == ps.Probes[excluded_probe] {
		t.Logf("Probes that should not match are equal to each other")
		t.Fail()
	}

	// Verify status is properly set
	if *ps.Probes[excluded_probe].Status != Excluded {
		t.Logf("Excluded probe was not excluded from probe store")
		t.Fail()
	}
	if *ps.Probes[probe].Status == Excluded {
		t.Logf("Excluded probe was not excluded from probe store")
		t.Fail()
	}
	// Note: this is not currently testing whether the summary or audit
	// are properly set for this because we may change how that is handled
	// without effecting probr functionality
}
