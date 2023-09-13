package compose

import (
	"bytes"
	"context"
	"strings"
	"sync"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

type ComposeService struct {
	mu          sync.Mutex
	out, err    *bytes.Buffer
	projectName string
	project     *types.Project
	service     api.Service
}

// NewComposeService returns *Compose service wrapper.
//
// clientOpt is used to initialize docker client.
// If clientOpt is nil, it uses a default option which is set Context to "default" and others left empty.
//
// stdout and stderr are used to redirect compose output. If both or either is nil, outputs for that side would be simply discarded.
// Be cautious, however, compose calls os.Exit(1) in case of some errors. Redirecting values to log or somewhere is strongly advised.
func NewComposeService(
	dockerCli *command.DockerCli,
	outBuf, errBuf *bytes.Buffer,
	projectName string,
	project *types.Project,
) *ComposeService {
	serviceProxy := api.NewServiceProxy().WithService(compose.NewComposeService(dockerCli))

	// Mimicking toProject of cli/cli.
	// Without this, docker compose v2 lose track of project and therefore would not be able to recreate services.
	for i, service := range project.Services {
		service.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     service.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}
		project.Services[i] = service
	}

	return &ComposeService{
		service:     serviceProxy,
		out:         outBuf,
		err:         errBuf,
		projectName: projectName,
		project:     project,
	}
}

func (c *ComposeService) resetBuf() {
	c.out.Reset()
	c.err.Reset()
}

func (c *ComposeService) parseOutput() ComposeOutput {
	out := ComposeOutput{
		Resource: make(map[string]ComposeOutputLine),
	}
	for _, lines := range []string{c.out.String(), c.err.String()} {
		for _, line := range strings.Split(lines, "\n") {
			if line == "" {
				continue
			}
			decoded, err := DecodeComposeOutputLine(line, c.projectName, c.project)
			if err != nil {
				continue
			}
			out.Resource[decoded.Name] = decoded
		}
	}
	return out
}

// Create executes the equivalent to a `compose create`
func (c *ComposeService) Create(ctx context.Context, options api.CreateOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if err := c.service.Create(ctx, c.project, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// Start executes the equivalent to a `compose start`
func (c *ComposeService) Start(ctx context.Context, options api.StartOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Start(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// Restart restarts containers
func (c *ComposeService) Restart(ctx context.Context, options api.RestartOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Restart(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// Stop executes the equivalent to a `compose stop`
func (c *ComposeService) Stop(ctx context.Context, options api.StopOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Stop(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// Down executes the equivalent to a `compose down`
func (c *ComposeService) Down(ctx context.Context, options api.DownOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Down(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// Ps executes the equivalent to a `compose ps`
func (c *ComposeService) Ps(ctx context.Context, options api.PsOptions) ([]api.ContainerSummary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	summary, err := c.service.Ps(ctx, c.projectName, options)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// Kill executes the equivalent to a `compose kill`
func (c *ComposeService) Kill(ctx context.Context, options api.KillOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Kill(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// RunOneOffContainer creates a service oneoff container and starts its dependencies
func (c *ComposeService) RunOneOffContainer(ctx context.Context, opts api.RunOptions) (exitCode int, stdout, stderr string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if opts.Project == nil {
		opts.Project = c.project
	}
	exitCode, err = c.service.RunOneOffContainer(ctx, c.project, opts)
	if err != nil {
		return exitCode, c.out.String(), c.err.String(), err
	}
	return exitCode, c.out.String(), c.err.String(), nil
}

// Remove executes the equivalent to a `compose rm`
func (c *ComposeService) Remove(ctx context.Context, options api.RemoveOptions) (ComposeOutput, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.resetBuf()
	if options.Project == nil {
		options.Project = c.project
	}
	if err := c.service.Remove(ctx, c.projectName, options); err != nil {
		return ComposeOutput{}, err
	}
	return c.parseOutput(), nil
}

// DryRunMode switches c to dry run mode if dryRun is true.
// Implementations might not change back to normal mode even if dryRun is false.
// User must call this only once and only when the user whishes to use dry run client.
func (c *ComposeService) DryRunMode(ctx context.Context, dryRun bool) (context.Context, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.service.DryRunMode(ctx, dryRun)
}
