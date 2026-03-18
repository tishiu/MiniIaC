package docker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerRuntime interface {
	CreateContainer(ctx context.Context, spec ContainerSpec) (ContainerSnapshot, error)
	InspectContainer(ctx context.Context, id string) (ContainerSnapshot, bool, error)
	DeleteContainer(ctx context.Context, id string) error

	CreateNetwork(ctx context.Context, spec NetworkSpec) (NetworkSnapshot, error)
	InspectNetwork(ctx context.Context, id string) (NetworkSnapshot, bool, error)
	DeleteNetwork(ctx context.Context, id string) error
}

type ContainerSpec struct {
	Name      string
	Image     string
	Port      int
	NetworkID string
}

type ContainerSnapshot struct {
	ID        string
	Image     string
	Status    string
	IPAddress string
	Port      int
	NetworkID string
}

type NetworkSpec struct {
	Name   string
	Driver string
}

type NetworkSnapshot struct {
	ID     string
	Name   string
	Driver string
}

type sdkRuntime struct {
	client *client.Client
}

func NewRuntimeFromEnv() (DockerRuntime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &sdkRuntime{client: cli}, nil
}

func (r *sdkRuntime) CreateContainer(ctx context.Context, spec ContainerSpec) (ContainerSnapshot, error) {
	containerConfig := &container.Config{
		Image: spec.Image,
	}
	hostConfig := &container.HostConfig{}

	if spec.Port > 0 {
		portStr := fmt.Sprintf("%d/tcp", spec.Port)
		containerConfig.ExposedPorts = nat.PortSet{
			nat.Port(portStr): struct{}{},
		}
		hostConfig.PortBindings = nat.PortMap{
			nat.Port(portStr): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(spec.Port),
				},
			},
		}
	}

	if spec.NetworkID != "" {
		hostConfig.NetworkMode = container.NetworkMode(spec.NetworkID)
	}

	resp, err := r.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, spec.Name)
	if err != nil {
		return ContainerSnapshot{}, fmt.Errorf("failed to create container: %w", err)
	}

	if err := r.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return ContainerSnapshot{}, fmt.Errorf("failed to start container: %w", err)
	}

	return ContainerSnapshot{
		ID:        resp.ID,
		Image:     spec.Image,
		Status:    "running",
		Port:      spec.Port,
		NetworkID: spec.NetworkID,
	}, nil
}

func (r *sdkRuntime) InspectContainer(ctx context.Context, id string) (ContainerSnapshot, bool, error) {
	inspect, err := r.client.ContainerInspect(ctx, id)
	if err != nil {
		if client.IsErrNotFound(err) {
			return ContainerSnapshot{}, false, nil
		}
		return ContainerSnapshot{}, false, fmt.Errorf("failed to inspect container: %w", err)
	}

	snapshot := ContainerSnapshot{
		ID:     inspect.ID,
		Image:  inspect.Config.Image,
		Status: inspect.State.Status,
	}
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.IPAddress != "" {
		snapshot.IPAddress = inspect.NetworkSettings.IPAddress
	}

	return snapshot, true, nil
}

func (r *sdkRuntime) DeleteContainer(ctx context.Context, id string) error {
	timeout := 10
	err := r.client.ContainerStop(ctx, id, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	err = r.client.ContainerRemove(ctx, id, container.RemoveOptions{
		Force: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

func (r *sdkRuntime) CreateNetwork(ctx context.Context, spec NetworkSpec) (NetworkSnapshot, error) {
	driver := spec.Driver
	if driver == "" {
		driver = "bridge"
	}

	resp, err := r.client.NetworkCreate(ctx, spec.Name, network.CreateOptions{
		Driver: driver,
	})
	if err != nil {
		return NetworkSnapshot{}, fmt.Errorf("failed to create network: %w", err)
	}

	return NetworkSnapshot{
		ID:     resp.ID,
		Name:   spec.Name,
		Driver: driver,
	}, nil
}

func (r *sdkRuntime) InspectNetwork(ctx context.Context, id string) (NetworkSnapshot, bool, error) {
	inspect, err := r.client.NetworkInspect(ctx, id, network.InspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return NetworkSnapshot{}, false, nil
		}
		return NetworkSnapshot{}, false, fmt.Errorf("failed to inspect network: %w", err)
	}
	return NetworkSnapshot{
		ID:     inspect.ID,
		Name:   inspect.Name,
		Driver: inspect.Driver,
	}, true, nil
}

func (r *sdkRuntime) DeleteNetwork(ctx context.Context, id string) error {
	err := r.client.NetworkRemove(ctx, id)
	if err != nil && !client.IsErrNotFound(err) {
		return fmt.Errorf("failed to remove network: %w", err)
	}
	return nil
}
