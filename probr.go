package probr

import (
	"github.com/citihub/probr/internal/clouddriver/kubernetes"
	"github.com/citihub/probr/internal/coreengine"
	_ "github.com/citihub/probr/probes/clouddriver"
	k8s_probes "github.com/citihub/probr/probes/kubernetes"
)

//TODO: revise when interface this bit up ...
var kube = kubernetes.GetKubeInstance()

func RunAllProbes() (int, *coreengine.ProbeStore, error) {
	ts := coreengine.NewProbeStore() // get the test mgr

	for _, probe := range k8s_probes.Probes {
		ts.AddProbe(probe.GetGodogProbe())
	}

	s, err := ts.ExecAllProbes() // Executes all added (queued) tests
	return s, ts, err
}

//GetAllProbeResults ...
func GetAllProbeResults(ts *coreengine.ProbeStore) (map[string]string, error) {
	out := make(map[string]string)
	for name := range ts.Tests {
		r, n, err := ReadProbeResults(ts, name)
		if err != nil {
			return nil, err
		}
		if r != "" {
			out[n] = r
		}
	}
	return out, nil
}

//ReadProbeResults ...
func ReadProbeResults(ts *coreengine.ProbeStore, name string) (string, string, error) {
	t, err := ts.GetProbe(name)
	test := t
	if err != nil {
		return "", "", err
	}
	r := test.Results
	n := test.ProbeDescriptor.Name
	if r != nil {
		b := r.Bytes()
		return string(b), n, nil
	}
	return "", "", nil
}
