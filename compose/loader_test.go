package compose

import (
	"context"
	"os"
	"testing"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/google/go-cmp/cmp"
)

func TestPreloadConfigDetails(t *testing.T) {
	projectName := "testdata"
	loadedNormaly, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "./testdata/compose.yml"},
				{Filename: "./testdata/additional.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		func(o *loader.Options) {
			o.SetProjectName(projectName, true)
		},
	)
	if err != nil {
		panic(err)
	}

	confDetail, err := PreloadConfigDetails(types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{
			{Filename: "./testdata/compose.yml"},
			{Filename: "./testdata/additional.yml"},
		},
		Environment: types.NewMapping(os.Environ()),
	})
	if err != nil {
		panic(err)
	}

	cachedConfig, err := loader.LoadWithContext(
		context.Background(),
		confDetail,
		func(o *loader.Options) {
			o.SetProjectName(projectName, true)
		},
	)
	if err != nil {
		panic(err)
	}

	if diff := cmp.Diff(loadedNormaly, cachedConfig); diff != "" {
		t.Errorf("not equal. diff = %s", diff)
	}
}
