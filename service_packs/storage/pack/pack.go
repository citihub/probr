package kubernetes_pack

import (
	"log"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
	"github.com/citihub/probr/internal/utils"
	"github.com/citihub/probr/service_packs/storage/access_whitelisting"
	"github.com/citihub/probr/service_packs/storage/encryption_at_rest"
	"github.com/citihub/probr/service_packs/storage/encryption_in_flight"
)

func GetProbes() []coreengine.Probe {
	conf := config.Vars.ServicePacks.Storage
	if conf.Provider == "" {
		file, line := utils.CallerFileLine()
		log.Printf("[WARN] %s:%v: Ignoring Storage service pack due to required vars not being present.", file, line)
		return nil
	}
	return []coreengine.Probe{
		access_whitelisting.Probe,
		encryption_at_rest.Probe,
		encryption_in_flight.Probe,
	}
}
