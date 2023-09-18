package compose

import (
	"slices"
	"sort"

	"github.com/compose-spec/compose-go/types"
)

// EnableAllService sets p.Profile every possible profile in p.Services and p.DisabledServices,
// and call p.EnableServices with all service names.
func EnableAllService(p *types.Project) {
	var serviceNames []string
	profileSet := map[string]struct{}{}
	for _, s := range p.AllServices() {
		serviceNames = append(serviceNames, s.Name)
		for _, p := range s.Profiles {
			profileSet[p] = struct{}{}
		}
	}

	var profiles []string
	for p := range profileSet {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)

	for _, prof := range profiles {
		if !slices.Contains(p.Profiles, prof) {
			p.Profiles = append(p.Profiles, prof)
		}
	}

	if err := p.EnableServices(serviceNames...); err != nil {
		panic(err)
	}
}
