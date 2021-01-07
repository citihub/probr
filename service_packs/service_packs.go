package service_packs

import (
	"log"
	"path/filepath"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/utils"
	kubernetes_pack "github.com/citihub/probr/service_packs/kubernetes/pack"
	storage_pack "github.com/citihub/probr/service_packs/storage/pack"
)

func packs() (packs map[string][]coreengine.Probe) {
	packs = make(map[string][]coreengine.Probe)

	packs["kubernetes"] = kubernetes_pack.GetProbes()
	packs["storage"] = storage_pack.GetProbes()

	if config.Vars.Meta.RunOnly != "" {
		log.Printf("[INFO] Running only the %s service pack", config.Vars.Meta.RunOnly)
		pack := packs[config.Vars.Meta.RunOnly]          // store desired pack
		packs = make(map[string][]coreengine.Probe) // clear all unspecified packs
		packs[config.Vars.Meta.RunOnly] = pack           // queue only the specified pack
	}
	return
}

func makeGodogProbe(pack string, p coreengine.Probe) *coreengine.GodogProbe {
	box := utils.BoxStaticFile(pack+p.Name(), "service_packs", pack, p.Name()) // Establish static files for binary build
	descriptor := coreengine.ProbeDescriptor{Group: coreengine.Kubernetes, Name: p.Name()}
	path := filepath.Join(box.ResolutionDir, p.Name()+".feature")
	return &coreengine.GodogProbe{
		ProbeDescriptor:     &descriptor,
		ProbeInitializer:    p.ProbeInitialize,
		ScenarioInitializer: p.ScenarioInitialize,
		FeaturePath:         path,
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
