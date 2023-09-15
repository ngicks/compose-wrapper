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
	"github.com/ngicks/compose-wrapper/compose"
)

func main() {
	projectName := "testdata"
	loader, err := compose.NewLoader(
		projectName,
		types.ConfigDetails{
			WorkingDir: "../compose/testdata",
			ConfigFiles: []types.ConfigFile{
				{Filename: "../compose/testdata/compose.yml"},
				{Filename: "../compose/testdata/additional.yml"},
			},
			Environment: types.NewMapping(os.Environ()),
		},
		[](func(*loader.Options)){},
		nil,
		command.WithOutputStream(io.Discard),
		command.WithErrorStream(io.Discard),
	)
	if err != nil {
		panic(err)
	}

	service, err := loader.LoadComposeService(context.Background())
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

	exitCode, stdout, stderr, err := service.RunOneOffContainer(context.Background(), api.RunOptions{Service: "additional"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("exitCode = %d, stdout = %s, stderr = %s\n\n", exitCode, stdout, stderr)
}
