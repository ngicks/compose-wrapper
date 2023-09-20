package compose

import (
	"slices"

	"github.com/compose-spec/compose-go/types"
)

// Reverse changes dst so that its enabled services are disabled in src.
func Reverse(src, dst *types.Project) error {
	serviceNames := src.ServiceNames()

	var disabledServices []string
	for _, disabled := range src.DisabledServices {
		if !slices.Contains(serviceNames, disabled.Name) {
			// In case caller is not correctly set up src.
			disabledServices = append(disabledServices, disabled.Name)
		}
	}

	if err := dst.EnableServices(disabledServices...); err != nil {
		return err
	}
	if err := dst.ForServices(disabledServices, types.IgnoreDependencies); err != nil {
		return err
	}
	return nil
}
