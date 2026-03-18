package reconciler

import (
	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/reference"
	"github.com/tishiu/MiniIac/pkg/state"
)

func InterpolateReferences(resource *config.Resource, currentState *state.State) *config.Resource {
	return reference.New().Resolve(resource, currentState)
}
