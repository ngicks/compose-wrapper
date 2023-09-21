package compose

import (
	"context"
	"os"
	"testing"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

const reverseComposeYaml = `services:
  enabled:
    image: ubuntu:jammy-20230624
    profiles:
      - enabled
    depends_on:
      dependency:
        condition: service_healthy
        restart: true
  dependency:
    image: ubuntu:jammy-20230624
    profiles:
     - disabled
  disabled:
    image: ubuntu:jammy-20230624
    profiles:
      - disabled
    depends_on:
      dependency:
        condition: service_healthy
        restart: true
`

func TestReverse(t *testing.T) {
	assert := assert.New(t)

	type testCase struct {
		enabled      []string
		forServices  []string
		enabledInSrc []string
		enabledInDst []string
	}
	for idx, tc := range []testCase{
		{
			enabled:      []string{"enabled", "dependency"},
			forServices:  []string{"enabled"},
			enabledInSrc: []string{"dependency", "enabled"},
			enabledInDst: []string{"disabled"},
		},
		{
			enabled:      []string{"enabled", "dependency", "disabled"},
			forServices:  []string{"enabled", "disabled"},
			enabledInSrc: []string{"dependency", "disabled", "enabled"},
			enabledInDst: nil,
		},
		{
			enabled:      []string{},
			forServices:  []string{},
			enabledInSrc: nil,
			enabledInDst: []string{"dependency", "disabled", "enabled"},
		},
	} {
		src := loadFromString(reverseComposeYaml)
		dst := loadFromString(reverseComposeYaml)
		assert.NoError(src.EnableServices(tc.enabled...))
		assert.NoError(src.ForServices(tc.forServices, types.IncludeDependencies))
		assert.NoError(dst.EnableServices(tc.enabled...))
		assert.NoError(dst.ForServices(tc.forServices, types.IncludeDependencies))
		assert.NoError(Reverse(src, dst))

		assertNoDiff := func(expected, actual any) {
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("case = %d. not equal. diff = %s", idx, diff)
			}
		}
		assertNoDiff(tc.enabledInSrc, src.ServiceNames())
		assertNoDiff(tc.enabledInDst, dst.ServiceNames())
		assertNoDiff(
			map[string]struct{}{"dependency": {}, "disabled": {}, "enabled": {}},
			toSet(src.AllServices()),
		)
		assertNoDiff(
			map[string]struct{}{"dependency": {}, "disabled": {}, "enabled": {}},
			toSet(dst.AllServices()),
		)
	}
}

func loadFromString(composeYmlStr string) *types.Project {
	loaded, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{
					Filename: "./testdata/whatever.yml",
					Content:  []byte(composeYmlStr),
				},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		func(o *loader.Options) {
			o.SetProjectName("example_compose", true)
		},
	)
	if err != nil {
		panic(err)
	}
	return loaded
}

func toSet(services []types.ServiceConfig) map[string]struct{} {
	out := make(map[string]struct{}, len(services))
	for _, s := range services {
		out[s.Name] = struct{}{}
	}
	return out
}
