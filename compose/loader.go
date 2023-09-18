package compose

import (
	"context"
	"os"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

// InitializeDockerCli initializes DockerCli.
//
// If clientOpt is nil, cli will be initialized with &flags.ClientOptions{Context: "default"}.
//
// ops will be applied twice, therefore all ops must be idempotent and/or must be aware of it.
// This is to encounter the case where passing malformed *flag.ClientOptions may cause it to exit by calling os.Exit(1).
// To prevent it from silently dying, this function sets err output stream to os.Stderr if it is not set.
// After initialization, it re-applies ops to ensure err output stream is what the caller wants to be.
func InitializeDockerCli(
	clientOpt *flags.ClientOptions,
	ops ...command.DockerCliOption,
) (cli *command.DockerCli, err error) {
	dockerCli, err := command.NewDockerCli(ops...)
	if err != nil {
		return nil, err
	}

	_ = dockerCli.Apply(command.WithErrorStream(os.Stderr))

	if clientOpt != nil {
		err = dockerCli.Initialize(clientOpt)
	} else {
		err = dockerCli.Initialize(&flags.ClientOptions{
			Context: "default",
		})
	}
	if err != nil {
		return nil, err
	}

	// This calls os.Exit(1) in case of initialization error.
	// That's why we've set stdout / stderr to output streams.
	_ = dockerCli.Client()

	_ = dockerCli.Apply(command.WithErrorStream(nil))
	if err := dockerCli.Apply(ops...); err != nil {
		return nil, err
	}

	if dockerCli.Err() == nil {
		_ = dockerCli.Apply(command.WithErrorStream(os.Stderr))
	}

	return dockerCli, nil
}

type Loader struct {
	DockerCli     *command.DockerCli
	ProjectName   string
	ConfigDetails types.ConfigDetails
	Options       []func(*loader.Options)
}

func NewLoader(
	projectName string,
	configDetails types.ConfigDetails,
	options []func(*loader.Options),
	clientOpt *flags.ClientOptions,
	ops ...command.DockerCliOption,
) (*Loader, error) {
	dockerCli, err := InitializeDockerCli(clientOpt, ops...)
	if err != nil {
		return nil, err
	}

	return &Loader{
		DockerCli:     dockerCli,
		ProjectName:   projectName,
		ConfigDetails: configDetails,
		Options:       options,
	}, nil
}

func (l *Loader) Load(ctx context.Context) (*types.Project, error) {
	return loader.LoadWithContext(
		ctx,
		l.ConfigDetails,
		append(
			l.Options,
			func(o *loader.Options) {
				o.SetProjectName(l.ProjectName, true)
			},
		)...,
	)
}

func (l *Loader) LoadComposeService(ctx context.Context, ops ...func(p *types.Project) error) (*ComposeService, error) {
	project, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	for _, op := range ops {
		if err := op(project); err != nil {
			return nil, err
		}
	}

	return NewComposeService(
		l.ProjectName,
		project,
		l.DockerCli,
	), nil
}
