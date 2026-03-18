package reference

import (
	"fmt"
	"regexp"

	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/state"
)

var refPattern = regexp.MustCompile(`\$\{(\w+)\.(\w+)\.(\w+)\}`)

type Engine struct{}

type Result struct {
	Dependencies []string
	Resolved     *config.Resource
}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Dependencies(resource *config.Resource) []string {
	if resource == nil {
		return nil
	}
	return ExtractDependencies(resource.Properties)
}

func (e *Engine) Resolve(resource *config.Resource, currentState *state.State) *config.Resource {
	if resource == nil {
		return nil
	}
	return &config.Resource{
		ID:         resource.ID,
		Type:       resource.Type,
		Properties: ResolveProperties(resource.Properties, currentState),
	}
}

func (e *Engine) Process(resource *config.Resource, currentState *state.State) (*Result, error) {
	if resource == nil {
		return &Result{}, nil
	}
	return &Result{
		Dependencies: e.Dependencies(resource),
		Resolved:     e.Resolve(resource, currentState),
	}, nil
}

func ExtractDependencies(properties map[string]interface{}) []string {
	refs := make([]string, 0)
	seen := make(map[string]bool)

	var scan func(interface{})
	scan = func(v interface{}) {
		switch val := v.(type) {
		case string:
			matches := refPattern.FindAllStringSubmatch(val, -1)
			for _, match := range matches {
				resourceID := match[2]
				if !seen[resourceID] {
					refs = append(refs, resourceID)
					seen[resourceID] = true
				}
			}
		case map[string]interface{}:
			for _, nested := range val {
				scan(nested)
			}
		case []interface{}:
			for _, item := range val {
				scan(item)
			}
		}
	}

	for _, v := range properties {
		scan(v)
	}

	return refs
}

func ResolveProperties(properties map[string]interface{}, currentState *state.State) map[string]interface{} {
	if properties == nil {
		return map[string]interface{}{}
	}

	resolved := make(map[string]interface{}, len(properties))
	for key, value := range properties {
		resolved[key] = resolveValue(value, currentState)
	}
	return resolved
}

func resolveValue(value interface{}, currentState *state.State) interface{} {
	switch v := value.(type) {
	case string:
		return resolveString(v, currentState)
	case map[string]interface{}:
		return ResolveProperties(v, currentState)
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = resolveValue(item, currentState)
		}
		return out
	default:
		return value
	}
}

func resolveString(input string, currentState *state.State) string {
	return refPattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := refPattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		resourceID := parts[2]
		attribute := parts[3]

		if currentState != nil {
			if resourceState, ok := currentState.Resources[resourceID]; ok {
				if attrValue, ok := resourceState.Attributes[attribute]; ok {
					return fmt.Sprint(attrValue)
				}
			}
		}

		return match
	})
}
