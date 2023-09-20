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
	src := loadFromString(reverseComposeYaml)
	dst := loadFromString(reverseComposeYaml)
	assert.NoError(src.EnableServices("enabled", "dependency"))
	assert.NoError(src.ForServices([]string{"enabled"}, types.IncludeDependencies))
	assert.NoError(Reverse(src, dst))

	if diff := cmp.Diff([]string{"dependency", "enabled"}, src.ServiceNames()); diff != "" {
		t.Errorf("not equal. diff = %s", diff)
	}
	if diff := cmp.Diff([]string{"disabled"}, dst.ServiceNames()); diff != "" {
		t.Errorf("not equal. diff = %s", diff)
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
