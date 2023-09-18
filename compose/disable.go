package compose

import (
	"slices"

	"github.com/compose-spec/compose-go/types"
)

// DisableProfiles disables service which does match selected profiles.
// Unlike (*types.Project).ApplyProfiles, DisableProfiles ignores already disabled services,
// or services which has no profile set.
func DisableProfiles(p *types.Project, profiles []string) {
	var services []types.ServiceConfig
	if slices.Contains(profiles, "*") {
		services = append(services, p.Services...)
	} else {
		for _, s := range p.Services {
			if len(s.Profiles) > 0 && s.HasProfile(profiles) {
				services = append(services, s)
			}
		}
	}

	var filtered []types.ServiceConfig
	for i := 0; i < len(p.Services); i++ {
		if !slices.ContainsFunc(services, func(sc types.ServiceConfig) bool {
			return sc.Name == p.Services[i].Name
		}) {
			p.DisableService(p.Services[i])
			filtered = append(filtered, p.Services[i])
		}
	}
	p.Services = filtered
	p.DisabledServices = services
}
