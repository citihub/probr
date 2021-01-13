package service_packs

import (
	"github.com/citihub/probr/internal/coreengine"
	kubernetes_pack "github.com/citihub/probr/service_packs/kubernetes/pack"
	opa_pack "github.com/citihub/probr/service_packs/opa/pack"
	storage_pack "github.com/citihub/probr/service_packs/storage/pack"
)

func packs() (packs map[string][]coreengine.Probe) {
	packs = make(map[string][]coreengine.Probe)

	packs["kubernetes"] = kubernetes_pack.GetProbes()
	packs["storage"] = storage_pack.GetProbes()
	packs["opa"] = opa_pack.GetProbes()

	return
}

func makeGodogProbe(pack string, p coreengine.Probe) *coreengine.GodogProbe {
	descriptor := coreengine.ProbeDescriptor{Group: coreengine.Kubernetes, Name: p.Name()}
	return &coreengine.GodogProbe{
		ProbeDescriptor:     &descriptor,
		ProbeInitializer:    p.ProbeInitialize,
		ScenarioInitializer: p.ScenarioInitialize,
		FeaturePath:         p.Path(),
	}
}

func GetAllProbes() []*coreengine.GodogProbe {
	var allProbes []*coreengine.GodogProbe

	for packName, pack := range packs() {
		for _, probe := range pack {
			allProbes = append(allProbes, makeGodogProbe(packName, probe))
		}
	}
	return allProbes
}
