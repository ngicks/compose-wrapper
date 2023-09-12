package main

import (
	"context"
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/client"
	"github.com/ngicks/compose-wrapper/compose"
)

func main() {
	projectName := "testdata"
	project, err := loader.LoadWithContext(
		context.Background(), types.ConfigDetails{
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

	dockerClient, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	if err := client.FromEnv(dockerClient); err != nil {
		panic(err)
	}

	service, err := compose.NewComposeService(
		projectName,
		project,
		nil,
		os.Stdout, os.Stderr,
		command.WithAPIClient(dockerClient),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("compose service created")
	list, err := service.Ps(context.Background(), api.PsOptions{Project: project, All: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+#v\n", list)
	if _, err := service.Create(context.Background(), api.CreateOptions{
		Recreate:             api.RecreateDiverged,
		RecreateDependencies: api.RecreateDiverged,
		Inherit:              true,
	}); err != nil {
		panic(err)
	}
	list, err = service.Ps(context.Background(), api.PsOptions{Project: project, All: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+#v\n", list)
}
