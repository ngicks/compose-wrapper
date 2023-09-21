package compose

import (
	"context"
	"testing"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestComposeService_dind(t *testing.T) {
	require := require.New(t)
	composeService, err := loaderAdditional.LoadComposeService(context.Background())
	require.NoError(err)

	dryRunCtx, err := composeService.DryRunMode(context.Background(), true)
	require.NoError(err)

	out, err := composeService.Create(dryRunCtx, api.CreateOptions{})
	require.NoError(err)

	delete(out.Resource, "Network:default")
	if diff := cmp.Diff(createDryRunOutputResourceMap, out.Resource); diff != "" {
		t.Errorf("not equal. diff =%s", diff)
	}
}
