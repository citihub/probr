package probr

import (
	"log"
	"os"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/service_packs"
)

func RunAllProbes() (int, *coreengine.ProbeStore, error) {
	ts := coreengine.NewProbeStore()

	for _, probe := range service_packs.GetAllProbes() {
		ts.AddProbe(probe)
	}

	s, err := ts.ExecAllProbes() // Executes all added (queued) tests
	return s, ts, err
}

//GetAllProbeResults maps ProbeStore results to strings
func GetAllProbeResults(ps *coreengine.ProbeStore) map[string]string {

	defer cleanup()

	out := make(map[string]string)
	for name := range ps.Probes {
		results, name, err := readProbeResults(ps, name)
		if err != nil {
			out[name] = err.Error()
		} else {
			out[name] = results
		}
	}
	return out
}

func readProbeResults(ps *coreengine.ProbeStore, name string) (probeResults, probeName string, err error) {
	p, err := ps.GetProbe(name)
	if err != nil {
		return
	}
	probeResults = p.Results.String()
	probeName = p.ProbeDescriptor.Name
	return
}

// cleanup is used to dispose of any temp resources used during probe execution
func cleanup() {
	// Remove tmp folder and its content
	err := os.RemoveAll(config.Vars.TmpDir())
	if err != nil {
		log.Printf("[ERROR] Error removing tmp folder %v", err)
	}
}
