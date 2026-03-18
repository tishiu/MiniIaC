package docker

import (
	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/state"
	"context"
	"fmt"
)

type NetworkProvider struct {
	runtime DockerRuntime
}

func NewNetworkProvider(runtime DockerRuntime) *NetworkProvider {
	return &NetworkProvider{runtime: runtime}
}

func NewNetworkProviderFromEnv() (*NetworkProvider, error) {
	runtime, err := NewRuntimeFromEnv()
	if err != nil {
		return nil, err
	}
	return NewNetworkProvider(runtime), nil
}

func (p *NetworkProvider) Create(ctx context.Context, desired *config.Resource) (*state.ResourceState, error) {
	name, ok := desired.Properties["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name property required")
	}

	driver := "bridge" // Default driver
	if driveVal, ok := desired.Properties["driver"].(string); ok {
		driver = driveVal
	}

	snap, err := p.runtime.CreateNetwork(ctx, NetworkSpec{
		Name:   name,
		Driver: driver,
	})
	if err != nil {
		return nil, err
	}

	return &state.ResourceState{
		ID:   snap.ID,
		Type: desired.Type,
		Attributes: map[string]interface{}{
			"network_id": snap.ID,
			"name":       snap.Name,
			"driver":     snap.Driver,
		},
	}, nil
}

func (p *NetworkProvider) Read(ctx context.Context, resourceID string) (*state.ResourceState, error) {
	snap, found, err := p.runtime.InspectNetwork(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	return &state.ResourceState{
		ID:   snap.ID,
		Type: "docker_network",
		Attributes: map[string]interface{}{
			"network_id": snap.ID,
			"name":       snap.Name,
			"driver":     snap.Driver,
		},
	}, nil
}

func (p *NetworkProvider) Update(ctx context.Context, desired *config.Resource, resourceID string) (*state.ResourceState, error) {
	// For networks, update = recreate
	if err := p.Delete(ctx, resourceID); err != nil {
		return nil, err
	}

	return p.Create(ctx, desired)
}

func (p *NetworkProvider) Delete(ctx context.Context, resourceID string) error {
	return p.runtime.DeleteNetwork(ctx, resourceID)
}
