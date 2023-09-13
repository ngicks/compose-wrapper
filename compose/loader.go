package compose

import (
	"bytes"
	"context"
	"io"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

func InitializeDockerCli(
	clientOpt *flags.ClientOptions,
	stdout, stderr io.Writer,
	ops ...command.DockerCliOption,
) (cli *command.DockerCli, bufOut, bufErr *bytes.Buffer, err error) {
	bufOut = new(bytes.Buffer)
	bufErr = new(bytes.Buffer)

	var consoleStd, consoleErr io.Writer = bufOut, bufErr
	if stdout != nil {
		consoleStd = io.MultiWriter(bufOut, stdout)
	}
	if stderr != nil {
		consoleErr = io.MultiWriter(bufErr, stderr)
	}

	ops = append(ops, command.WithOutputStream(consoleStd), command.WithErrorStream(consoleErr))
	dockerCli, err := command.NewDockerCli(ops...)
	if err != nil {
		return nil, nil, nil, err
	}
	if clientOpt != nil {
		err = dockerCli.Initialize(clientOpt)
	} else {
		err = dockerCli.Initialize(&flags.ClientOptions{
			Context: "default",
		})
	}
	if err != nil {
		return nil, nil, nil, err
	}
	return dockerCli, bufOut, bufErr, nil
}

type Loader struct {
	DockerCli      *command.DockerCli
	ProjectName    string
	ConfigDetails  types.ConfigDetails
	Options        []func(*loader.Options)
	OutBuf, ErrBuf *bytes.Buffer
}

func NewLoader(
	clientOpt *flags.ClientOptions,
	redirectedOut, redirectedErr io.Writer,
	projectName string,
	configDetails types.ConfigDetails,
	options []func(*loader.Options),
	ops ...command.DockerCliOption,
) (*Loader, error) {
	var err error

	dockerCli, outBuf, errBuf, err := InitializeDockerCli(
		clientOpt,
		redirectedOut, redirectedErr,
		ops...,
	)

	if err != nil {
		return nil, err
	}
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

	return &Loader{
		DockerCli:     dockerCli,
		ProjectName:   projectName,
		ConfigDetails: configDetails,
		Options:       options,
		OutBuf:        outBuf,
		ErrBuf:        errBuf,
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

func (l *Loader) LoadComposeService(ctx context.Context, stdout io.Writer, stderr io.Writer) (*ComposeService, error) {
	project, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return NewComposeService(
		l.DockerCli,
		l.OutBuf,
		l.ErrBuf,
		l.ProjectName,
		project,
	), nil
}
