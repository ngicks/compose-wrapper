package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/google/go-cmp/cmp"
	"github.com/ngicks/compose-wrapper/compose"
)

func main() {
	projectName := "testdata"

	loadedNormaly, err := loader.LoadWithContext(
		context.Background(),
		types.ConfigDetails{
			WorkingDir: "../compose/testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "../compose/testdata/compose.yml"},
				{Filename: "../compose/testdata/additional.yml"},
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

	confDetail, err := compose.PreloadConfigDetails(types.ConfigDetails{
		WorkingDir: "../compose/testdata",
		ConfigFiles: []types.ConfigFile{
			{Filename: "../compose/testdata/compose.yml"},
			{Filename: "../compose/testdata/additional.yml"},
		},
		Environment: types.NewMapping(os.Environ()),
	})
	if err != nil {
		panic(err)
	}
	loader, err := compose.NewLoader(
		projectName,
		confDetail,
		[](func(*loader.Options)){},
		nil,
		command.WithOutputStream(io.Discard),
		command.WithErrorStream(io.Discard),
	)
	if err != nil {
		panic(err)
	}

	project, _ := loader.Load(context.Background())

	// no diff.
	fmt.Printf("diff = %s\n", cmp.Diff(loadedNormaly, project))

	service, err := loader.LoadComposeService(context.Background(), func(p *types.Project) error {
		compose.EnableAllService(p)
		return nil
	})
	if err != nil {
		panic(err)
	}

	_, err = service.DryRunMode(context.Background(), true)
	if err != nil {
		panic(err)
	}

	fmt.Println("compose service created")
	list, err := service.Ps(context.Background(), api.PsOptions{All: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+#v\n\n", list)
	out, err := service.Create(context.Background(), api.CreateOptions{
		Recreate:             api.RecreateDiverged,
		RecreateDependencies: api.RecreateDiverged,
		RemoveOrphans:        true,
		Inherit:              true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+#v\n\n", out)

	list, err = service.Ps(context.Background(), api.PsOptions{All: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+#v\n\n", list)

	exitCode, stdout, stderr, err := service.RunOneOffContainer(context.Background(), api.RunOptions{Service: "additional", AutoRemove: true})
	if err != nil {
		fmt.Printf("err = %v\n\n", err)
	}
	fmt.Printf("exitCode = %d, stdout = %s, stderr = %s\n\n", exitCode, stdout, stderr)

	// out, err = service.Down(context.Background(), api.DownOptions{RemoveOrphans: true, Volumes: true})
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%+#v\n\n", out)
}
