package compose

import (
	"context"
	"os"
	"testing"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/google/go-cmp/cmp"
)

var (
	loaderBase, loaderAdditional, loaderAdditional2 *Loader
)

func init() {
	loaderBase, _ = NewLoader(
		"testdata",
		types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "./testdata/compose.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		[]func(*loader.Options){func(o *loader.Options) { o.Profiles = []string{"*"} }},
		nil,
	)
	loaderAdditional, _ = NewLoader(
		"testdata",
		types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "./testdata/compose.yml"},
				{Filename: "./testdata/additional.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		[]func(*loader.Options){func(o *loader.Options) { o.Profiles = []string{"*"} }},
		nil,
	)
	loaderAdditional2, _ = NewLoader(
		"testdata",
		types.ConfigDetails{
			WorkingDir: "./testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "./testdata/compose.yml"},
				{Filename: "./testdata/additional.yml"},
				{Filename: "./testdata/additional2.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		[]func(*loader.Options){func(o *loader.Options) { o.Profiles = []string{"*"} }},
		nil,
	)
}

func TestCompareProjectImage(t *testing.T) {
	ctx := context.Background()
	old, _ := loaderAdditional.Load(ctx)
	newer, _ := loaderAdditional2.Load(ctx)

	onlyInOld, addedInNew := CompareProjectImage(old, newer)
	if diff := cmp.Diff([]string{"debian:bookworm-20230904"}, onlyInOld); diff != "" {
		t.Errorf("not equal. diff = %s", diff)
	}
	if diff := cmp.Diff([]string(nil), addedInNew); diff != "" {
		t.Errorf("not equal. diff = %s", diff)
	}
}
