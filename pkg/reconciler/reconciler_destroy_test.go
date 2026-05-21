package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tishiu/MiniIac/pkg/graph"
	"github.com/tishiu/MiniIac/pkg/state"
)

func TestBuildDestroyResources_InfersContainerNetworkDependency(t *testing.T) {
	currentState := state.NewState()
	currentState.Resources["backend_net"] = &state.ResourceState{
		ID:   "network-1",
		Type: "docker_network",
		Attributes: map[string]interface{}{
			"network_id": "network-1",
			"name":       "backend-net",
		},
	}
	currentState.Resources["app"] = &state.ResourceState{
		ID:   "container-1",
		Type: "docker_container",
		Attributes: map[string]interface{}{
			"network_id": "network-1",
			"image":      "nginx:alpine",
		},
	}

	resources := buildDestroyResources(currentState)
	require.Len(t, resources, 2)

	byID := make(map[string]map[string]interface{}, len(resources))
	for _, resource := range resources {
		byID[resource.ID] = resource.Properties
	}

	assert.Equal(t, "${docker_network.backend_net.network_id}", byID["app"]["network_id"])
	assert.Equal(t, "network-1", byID["backend_net"]["network_id"])

	g := graph.NewGraph()
	require.NoError(t, g.Build(resources))

	order, err := g.TopologicalSort()
	require.NoError(t, err)

	appIndex := indexOf(order, "app")
	netIndex := indexOf(order, "backend_net")
	require.NotEqual(t, -1, appIndex)
	require.NotEqual(t, -1, netIndex)
	assert.Less(t, appIndex, netIndex, "dependent container must be destroyed before network")
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}
