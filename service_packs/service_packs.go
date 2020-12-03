package service_packs

import (
	"path/filepath"

	"github.com/cucumber/godog"

	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/utils"
	"github.com/citihub/probr/service_packs/kubernetes/container_registry_access"
	"github.com/citihub/probr/service_packs/kubernetes/general"
	"github.com/citihub/probr/service_packs/kubernetes/iam"
	"github.com/citihub/probr/service_packs/kubernetes/internet_access"
	"github.com/citihub/probr/service_packs/kubernetes/pod_security_policy"
	"github.com/citihub/probr/service_packs/storage/encryption_in_flight"
)

type probe interface {
	ProbeInitialize(*godog.TestSuiteContext)
	ScenarioInitialize(*godog.ScenarioContext)
	Name() string
}

//var packs map[string][]probe
var packs map[coreengine.ServicePack][]probe

func init() {
	packs = make(map[coreengine.ServicePack][]probe)
	packs[coreengine.Kubernetes] = []probe{
		container_registry_access.Probe,
		general.Probe,
		pod_security_policy.Probe,
		internet_access.Probe,
		iam.Probe,
	}
	packs[coreengine.Storage] = []probe{
		encryption_in_flight.Probe,
	}
}

func makeGodogProbe(pack coreengine.ServicePack, p probe) *coreengine.GodogProbe {
	box := utils.BoxStaticFile(pack.String()+p.Name(), "service_packs", pack.String(), p.Name()) // Establish static files for binary build
	descriptor := coreengine.ProbeDescriptor{ServicePack: pack, Name: p.Name()}
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

	for servicePack, pack := range packs {
		for _, probe := range pack {
			allProbes = append(allProbes, makeGodogProbe(servicePack, probe))
		}
	}
	return allProbes
}
