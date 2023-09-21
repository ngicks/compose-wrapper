package compose

import (
	"slices"
	"sync"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

type LoaderProxy struct {
	mu     sync.RWMutex
	loader *Loader
}

func NewLoaderProxy(
	projectName string,
	configDetails types.ConfigDetails,
	options []func(*loader.Options),
	clientOpt *flags.ClientOptions,
	ops ...command.DockerCliOption,
) (*LoaderProxy, error) {
	loader, err := NewLoader(projectName, configDetails, options, clientOpt, ops...)
	if err != nil {
		return nil, err
	}

	return &LoaderProxy{
		loader: loader,
	}, nil
}

func (p *LoaderProxy) PreloadConfigDetails() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	loaded, err := PreloadConfigDetails(p.loader.ConfigDetails)
	if err != nil {
		return err
	}
	p.loader.ConfigDetails = loaded
	return nil
}

func (p *LoaderProxy) DockerCli() *command.DockerCli {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loader.DockerCli
}

func (p *LoaderProxy) ProjectName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loader.ProjectName
}

func (p *LoaderProxy) ConfigDetails() types.ConfigDetails {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return cloneConfigDetails(p.loader.ConfigDetails)
}

func (p *LoaderProxy) Options() []func(*loader.Options) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return slices.Clone(p.loader.Options)
}

func (p *LoaderProxy) UpdateDockerCli(dockerCli *command.DockerCli) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loader.DockerCli = dockerCli
}
func (p *LoaderProxy) UpdateProjectName(projectName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loader.ProjectName = projectName
}
func (p *LoaderProxy) UpdateConfigDetails(configDetails types.ConfigDetails) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loader.ConfigDetails = configDetails
}
func (p *LoaderProxy) UpdateOptions(options []func(*loader.Options)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.loader.Options = options
}
