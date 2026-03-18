package cli

import (
	"fmt"

	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/provider"
	"github.com/tishiu/MiniIac/pkg/provider/docker"
	"github.com/tishiu/MiniIac/pkg/provider/local"
	"github.com/tishiu/MiniIac/pkg/reconciler"
	"github.com/tishiu/MiniIac/pkg/state"
)

type CLI struct {
	stateManager *state.Manager
	parser       *config.Parser
	catalog      *provider.Catalog
	reconciler   *reconciler.Reconciler
}

func NewCLI() (*CLI, error) {
	stateManager := state.NewManager()
	parser := config.NewParser()
	catalog := provider.NewCatalog()
	runtime, err := docker.NewRuntimeFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker runtime: %w", err)
	}

	if err := registerProviders(catalog, runtime); err != nil {
		return nil, fmt.Errorf("failed to register providers: %w", err)
	}

	reconcilerInstance := reconciler.NewReconciler(stateManager, catalog)

	return &CLI{
		stateManager: stateManager,
		parser:       parser,
		catalog:      catalog,
		reconciler:   reconcilerInstance,
	}, nil
}

// registerProviders registers all built-in providers
func registerProviders(catalog *provider.Catalog, runtime docker.DockerRuntime) error {
	if err := catalog.Register(provider.Definition{
		Type:     "local_file",
		Provider: local.NewFileProvider(),
		Schema: provider.Schema{
			Required: []string{"path", "content"},
		},
	}); err != nil {
		return err
	}

	if err := catalog.Register(provider.Definition{
		Type:     "docker_container",
		Provider: docker.NewContainerProvider(runtime),
		Schema: provider.Schema{
			Required: []string{"image"},
			Optional: []string{"port", "network_id"},
			Coerce: map[string]func(interface{}) (interface{}, error){
				"port": coerceInt,
			},
		},
	}); err != nil {
		return err
	}

	if err := catalog.Register(provider.Definition{
		Type:     "docker_network",
		Provider: docker.NewNetworkProvider(runtime),
		Schema: provider.Schema{
			Required: []string{"name"},
			Optional: []string{"driver"},
			Defaults: map[string]interface{}{
				"driver": "bridge",
			},
		},
	}); err != nil {
		return err
	}

	return nil
}

func coerceInt(v interface{}) (interface{}, error) {
	switch value := v.(type) {
	case int:
		return value, nil
	case float64:
		return int(value), nil
	default:
		return nil, fmt.Errorf("must be a number")
	}
}
