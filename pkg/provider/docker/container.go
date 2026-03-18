package docker

import (
	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/state"
	"context"
	"fmt"
)

type ContainerProvider struct {
	runtime DockerRuntime
}

func NewContainerProvider(runtime DockerRuntime) *ContainerProvider {
	return &ContainerProvider{runtime: runtime}
}

func NewContainerProviderFromEnv() (*ContainerProvider, error) {
	runtime, err := NewRuntimeFromEnv()
	if err != nil {
		return nil, err
	}
	return NewContainerProvider(runtime), nil
}

func (p *ContainerProvider) Create(ctx context.Context, desired *config.Resource) (*state.ResourceState, error) {
	image, ok := desired.Properties["image"].(string)
	if !ok {
		return nil, fmt.Errorf("image property required")
	}

	// Handle YAML number type (float64)
	var port int
	if portVal, ok := desired.Properties["port"]; ok {
		switch v := portVal.(type) {
		case float64:
			port = int(v)
		case int:
			port = v
		default:
			return nil, fmt.Errorf("port must be a number")
		}
	}
	// Optional network_id
	networkID, _ := desired.Properties["network_id"].(string)

	snap, err := p.runtime.CreateContainer(ctx, ContainerSpec{
		Name:      desired.ID,
		Image:     image,
		Port:      port,
		NetworkID: networkID,
	})
	if err != nil {
		return nil, err
	}

	attributes := map[string]interface{}{
		"container_id": snap.ID,
		"image":        image,
		"status":       snap.Status,
	}

	if port > 0 {
		attributes["port"] = port
	}
	if networkID != "" {
		attributes["network_id"] = networkID
	}
	if snap.IPAddress != "" {
		attributes["ip_address"] = snap.IPAddress
	}

	return &state.ResourceState{
		ID:         snap.ID,
		Type:       desired.Type,
		Attributes: attributes,
	}, nil
}

func (p *ContainerProvider) Read(ctx context.Context, resourceID string) (*state.ResourceState, error) {
	snap, found, err := p.runtime.InspectContainer(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	attributes := map[string]interface{}{
		"container_id": snap.ID,
		"image":        snap.Image,
		"status":       snap.Status,
	}

	if snap.IPAddress != "" {
		attributes["ip_address"] = snap.IPAddress
	}
	if snap.Port > 0 {
		attributes["port"] = snap.Port
	}
	if snap.NetworkID != "" {
		attributes["network_id"] = snap.NetworkID
	}

	return &state.ResourceState{
		ID:         snap.ID,
		Type:       "docker_container",
		Attributes: attributes,
	}, nil
}

func (p *ContainerProvider) Update(ctx context.Context, desired *config.Resource, resourceID string) (*state.ResourceState, error) {
	if err := p.Delete(ctx, resourceID); err != nil {
		return nil, err
	}

	return p.Create(ctx, desired)
}

func (p *ContainerProvider) Delete(ctx context.Context, resourceID string) error {
	return p.runtime.DeleteContainer(ctx, resourceID)
}
