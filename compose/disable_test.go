package compose

import (
	"context"
	"testing"

	"github.com/compose-spec/compose-go/types"
	"github.com/stretchr/testify/assert"
)

func TestDisableProfiles(t *testing.T) {
	assert := assert.New(t)
	project, _ := loaderAdditional2.Load(context.Background())

	EnableAllService(project)

	assert.Len(project.Services, 4)
	assert.Len(project.DisabledServices, 0)

	DisableProfiles(project, []string{"extended"})
	assert.Len(project.Services, 3)
	assert.Len(project.DisabledServices, 1)
	assertServiceHas(t, project, "sample_service")
	assertServiceHas(t, project, "additional2")
	assertServiceHas(t, project, "no_profile")

	EnableAllService(project)

	DisableProfiles(project, []string{"extended", "extended2"})
	assert.Len(project.Services, 2)
	assert.Len(project.DisabledServices, 2)
	assertServiceHas(t, project, "sample_service")
	assertServiceHas(t, project, "no_profile")

	EnableAllService(project)

	DisableProfiles(project, []string{"*"})
	assert.Len(project.Services, 0)
	assert.Len(project.DisabledServices, 4)
}

func assertServiceHas(t *testing.T, project *types.Project, name string) {
	t.Helper()

	_, err := project.GetService(name)
	if err != nil {
		t.Errorf("must found. err = %v", err)
	}
}
