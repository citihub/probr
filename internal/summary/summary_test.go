package summary

import (
	"path/filepath"
	"testing"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/utils"
)

func TestSummaryState_LogPodName(t *testing.T) {

	var fakeSummaryState SummaryState
	fakeSummaryState.Probes = make(map[string]*Probe)
	fakeSummaryState.Meta = make(map[string]interface{})
	fakeSummaryState.Meta["names of pods created"] = []string{}

	type args struct {
		podName string
	}
	tests := []struct {
		testName string
		s        *SummaryState
		args     args
	}{
		{
			testName: "MetaShouldContainPodName",
			s:        &fakeSummaryState,
			args:     args{podName: "testPod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tt.s.LogPodName(tt.args.podName)
			tt.s.LogPodName("Anotherpod")
			loggedPods := fakeSummaryState.Meta["names of pods created"].([]string)
			actualPosition, actualFound := utils.FindString(loggedPods, tt.args.podName)
			if !(actualPosition >= 0 && actualFound == true) {
				t.Errorf("State.Meta doesn't contain pod name: %v", loggedPods)
			}
		})
	}
}

func TestSummaryState_initProbe(t *testing.T) {

	type args struct {
		fakeName string
	}
	var fakeName = "testProbe"
	var fakeSummaryState SummaryState
	fakeSummaryState.Probes = make(map[string]*Probe)
	ap := filepath.Join(config.AuditDir(), (fakeName + ".json")) // Needed in both Probe and ProbeAudit
	fakeSummaryState.Probes[fakeName] = &Probe{
		name:          fakeName,
		Meta:          make(map[string]interface{}),
		PodsDestroyed: 0,
		audit: &ProbeAudit{
			Name: fakeName,
			path: ap,
		},
	}
	fakeSummaryState.Probes[fakeName].Meta["audit_path"] = ap
	fakeSummaryState.Probes[fakeName].audit.PodsDestroyed = &fakeSummaryState.Probes[fakeName].PodsDestroyed
	fakeSummaryState.Probes[fakeName].audit.ScenariosAttempted = &fakeSummaryState.Probes[fakeName].ScenariosAttempted
	fakeSummaryState.Probes[fakeName].audit.ScenariosSucceeded = &fakeSummaryState.Probes[fakeName].ScenariosSucceeded
	fakeSummaryState.Probes[fakeName].audit.ScenariosFailed = &fakeSummaryState.Probes[fakeName].ScenariosFailed
	fakeSummaryState.Probes[fakeName].audit.Result = &fakeSummaryState.Probes[fakeName].Result
	fakeSummaryState.Probes[fakeName].countResults()

	tests := []struct {
		testName string
		s        *SummaryState
		args     args
	}{
		{
			testName: "TestProbeInitialized",
			s:        &fakeSummaryState,
			args:     args{fakeName: "testProbe"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tt.s.initProbe(tt.args.fakeName)
			tt.s.initProbe("AnotherProbe")
			createdProbes := fakeSummaryState.Probes
			v, found := createdProbes["testProbe"]
			if !found {
				t.Errorf("Summary State doesn't contain probe name: %v", createdProbes)
				t.Logf("probe name found: %v", v)
			}

		})
	}
}
