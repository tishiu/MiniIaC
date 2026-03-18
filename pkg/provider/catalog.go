package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/state"
)

type Action uint8

const (
	ActionCreate Action = iota
	ActionUpdate
	ActionDelete
	ActionNoop
)

type Request struct {
	Type      string
	Action    Action
	Desired   *config.Resource
	CurrentID string
	Current   *state.ResourceState
}

type Schema struct {
	Required []string
	Optional []string
	Defaults map[string]interface{}
	Coerce   map[string]func(interface{}) (interface{}, error)
}

type Definition struct {
	Type     string
	Schema   Schema
	Provider Provider
}

type PreparedResource struct {
	Resource *config.Resource
	Provider Provider
}

type Catalog struct {
	mu   sync.RWMutex
	defs map[string]Definition
}

func NewCatalog() *Catalog {
	return &Catalog{
		defs: make(map[string]Definition),
	}
}

func (c *Catalog) Register(def Definition) error {
	if def.Type == "" {
		return fmt.Errorf("definition type is required")
	}
	if def.Provider == nil {
		return fmt.Errorf("definition provider is required for type %s", def.Type)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.defs[def.Type]; exists {
		return fmt.Errorf("provider already registered for type %s", def.Type)
	}
	c.defs[def.Type] = def
	return nil
}

func (c *Catalog) Lookup(resourceType string) (Provider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	def, ok := c.defs[resourceType]
	if !ok {
		return nil, fmt.Errorf("provider not found for resource type: %s", resourceType)
	}
	return def.Provider, nil
}

func (c *Catalog) Prepare(resource *config.Resource) (*PreparedResource, error) {
	if resource == nil {
		return nil, fmt.Errorf("resource is required")
	}

	c.mu.RLock()
	def, ok := c.defs[resource.Type]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resource.Type)
	}

	props := cloneProperties(resource.Properties)

	for key, value := range def.Schema.Defaults {
		if _, exists := props[key]; !exists {
			props[key] = value
		}
	}

	allowed := make(map[string]bool)
	for _, k := range def.Schema.Required {
		allowed[k] = true
	}
	for _, k := range def.Schema.Optional {
		allowed[k] = true
	}
	for k := range def.Schema.Defaults {
		allowed[k] = true
	}
	for k := range def.Schema.Coerce {
		allowed[k] = true
	}

	for _, req := range def.Schema.Required {
		if _, exists := props[req]; !exists {
			return nil, fmt.Errorf("resource %s (%s): missing required property %q", resource.ID, resource.Type, req)
		}
	}

	for key, value := range props {
		if !allowed[key] {
			return nil, fmt.Errorf("resource %s (%s): unknown property %q", resource.ID, resource.Type, key)
		}
		if coerceFn, ok := def.Schema.Coerce[key]; ok && coerceFn != nil {
			coerced, err := coerceFn(value)
			if err != nil {
				return nil, fmt.Errorf("resource %s (%s): invalid %q: %w", resource.ID, resource.Type, key, err)
			}
			props[key] = coerced
		}
	}

	return &PreparedResource{
		Resource: &config.Resource{
			ID:         resource.ID,
			Type:       resource.Type,
			Properties: props,
		},
		Provider: def.Provider,
	}, nil
}

func (c *Catalog) Execute(ctx context.Context, req Request) (*state.ResourceState, error) {
	switch req.Action {
	case ActionNoop:
		return req.Current, nil
	case ActionCreate:
		prepared, err := c.prepareFromRequest(req)
		if err != nil {
			return nil, err
		}
		newState, err := prepared.Provider.Create(ctx, prepared.Resource)
		if err != nil {
			return nil, fmt.Errorf("create failed for %s: %w", prepared.Resource.ID, err)
		}
		return newState, nil
	case ActionUpdate:
		if req.CurrentID == "" {
			return nil, fmt.Errorf("current resource id is required for update")
		}
		prepared, err := c.prepareFromRequest(req)
		if err != nil {
			return nil, err
		}
		newState, err := prepared.Provider.Update(ctx, prepared.Resource, req.CurrentID)
		if err != nil {
			return nil, fmt.Errorf("update failed for %s: %w", prepared.Resource.ID, err)
		}
		return newState, nil
	case ActionDelete:
		resourceType := req.Type
		if resourceType == "" && req.Current != nil {
			resourceType = req.Current.Type
		}
		if req.CurrentID == "" {
			return nil, fmt.Errorf("current resource id is required for delete")
		}
		prov, err := c.Lookup(resourceType)
		if err != nil {
			return nil, err
		}
		if err := prov.Delete(ctx, req.CurrentID); err != nil {
			return nil, fmt.Errorf("delete failed for %s: %w", req.CurrentID, err)
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown action: %d", req.Action)
	}
}

func (c *Catalog) prepareFromRequest(req Request) (*PreparedResource, error) {
	resource := req.Desired
	if resource == nil {
		return nil, fmt.Errorf("desired resource is required")
	}
	if req.Type != "" && resource.Type == "" {
		resource = &config.Resource{
			ID:         resource.ID,
			Type:       req.Type,
			Properties: cloneProperties(resource.Properties),
		}
	}
	return c.Prepare(resource)
}

func cloneProperties(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
