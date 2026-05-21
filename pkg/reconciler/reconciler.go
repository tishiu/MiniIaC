package reconciler

import (
	"context"
	"fmt"

	"github.com/tishiu/MiniIac/pkg/config"
	"github.com/tishiu/MiniIac/pkg/graph"
	"github.com/tishiu/MiniIac/pkg/logger"
	"github.com/tishiu/MiniIac/pkg/provider"
	"github.com/tishiu/MiniIac/pkg/state"
)

type Reconciler struct {
	stateManager *state.Manager
	catalog      *provider.Catalog
}

type Mode uint8

const (
	ModePreview Mode = iota
	ModeApply
	ModeDestroy
)

type Request struct {
	Mode    Mode
	Desired []*config.Resource
}

type PlanSummary struct {
	Create int
	Update int
	Delete int
	Noop   int
}

type PreparedPlan struct {
	reconciler *Reconciler
	mode       Mode
	desired    []*config.Resource
	changes    []*Change
	order      []string
}

func NewReconciler(stateManager *state.Manager, catalog *provider.Catalog) *Reconciler {
	return &Reconciler{
		stateManager: stateManager,
		catalog:      catalog,
	}
}

func (p *PreparedPlan) Changes() []*Change {
	return p.changes
}

func (p *PreparedPlan) Summary() PlanSummary {
	out := PlanSummary{}
	for _, change := range p.changes {
		switch change.Type {
		case ChangeTypeCreate:
			out.Create++
		case ChangeTypeUpdate:
			out.Update++
		case ChangeTypeDelete:
			out.Delete++
		case ChangeTypeNoop:
			out.Noop++
		}
	}
	return out
}

func (p *PreparedPlan) Discard() error {
	return nil
}

func (p *PreparedPlan) Commit(ctx context.Context) error {
	switch p.mode {
	case ModePreview:
		return nil
	case ModeApply:
		return p.reconciler.commitApply(ctx, p)
	case ModeDestroy:
		return p.reconciler.commitDestroy(ctx, p)
	default:
		return fmt.Errorf("unknown plan mode: %d", p.mode)
	}
}

func (r *Reconciler) Prepare(ctx context.Context, req Request) (*PreparedPlan, error) {
	switch req.Mode {
	case ModeDestroy:
		return r.prepareDestroy(ctx)
	case ModePreview, ModeApply:
		return r.prepareDesired(ctx, req.Mode, req.Desired)
	default:
		return nil, fmt.Errorf("unsupported mode: %d", req.Mode)
	}
}

func (r *Reconciler) Plan(desired []*config.Resource) ([]*Change, error) {
	plan, err := r.Prepare(context.Background(), Request{
		Mode:    ModePreview,
		Desired: desired,
	})
	if err != nil {
		return nil, err
	}
	return plan.Changes(), nil
}

func (r *Reconciler) Apply(ctx context.Context, desired []*config.Resource) error {
	plan, err := r.Prepare(ctx, Request{
		Mode:    ModeApply,
		Desired: desired,
	})
	if err != nil {
		return err
	}
	return plan.Commit(ctx)
}

func (r *Reconciler) Destroy(ctx context.Context) error {
	plan, err := r.Prepare(ctx, Request{
		Mode: ModeDestroy,
	})
	if err != nil {
		return err
	}
	return plan.Commit(ctx)
}

func (r *Reconciler) prepareDesired(ctx context.Context, mode Mode, desired []*config.Resource) (*PreparedPlan, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	log := logger.Get()

	normalized := make([]*config.Resource, 0, len(desired))
	for _, resource := range desired {
		prepared, err := r.catalog.Prepare(resource)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		normalized = append(normalized, prepared.Resource)
	}

	log.Info("loading current state")

	currentState, err := r.stateManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	log.Info("computing diff", "desired_count", len(normalized), "current_count", len(currentState.Resources))
	changes := ComputeDiff(normalized, currentState)

	g := graph.NewGraph()
	if err := g.Build(normalized); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	if err := g.ValidateDAG(); err != nil {
		return nil, fmt.Errorf("failed to validate dag: %w", err)
	}

	if err := g.ValidateReferences(); err != nil {
		return nil, fmt.Errorf("failed to validate references: %w", err)
	}

	var order []string
	if mode == ModeApply {
		// Use reverse topological order so dependencies are created before dependents
		// (edges go dependent→dependency, so reverse gives dependency-first order).
		order, err = g.TopologicalSortReverse()
		if err != nil {
			return nil, fmt.Errorf("failed to build topological order: %w", err)
		}
	}

	log.Info("plan complete", "changes", len(changes))
	return &PreparedPlan{
		reconciler: r,
		mode:       mode,
		desired:    normalized,
		changes:    changes,
		order:      order,
	}, nil
}

func (r *Reconciler) prepareDestroy(ctx context.Context) (*PreparedPlan, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	currentState, err := r.stateManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	if len(currentState.Resources) == 0 {
		return &PreparedPlan{
			reconciler: r,
			mode:       ModeDestroy,
			changes:    []*Change{},
			order:      []string{},
		}, nil
	}

	resources := buildDestroyResources(currentState)

	g := graph.NewGraph()
	if err := g.Build(resources); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}
	order, err := g.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to build destroy order: %w", err)
	}

	changes := make([]*Change, 0, len(order))
	for _, logicalID := range order {
		rs, ok := currentState.Resources[logicalID]
		if !ok {
			continue
		}
		changes = append(changes, &Change{
			Type: ChangeTypeDelete,
			Resource: &config.Resource{
				ID:   logicalID,
				Type: rs.Type,
			},
			OldState: rs,
			Reason:   "resource scheduled for destroy",
		})
	}

	return &PreparedPlan{
		reconciler: r,
		mode:       ModeDestroy,
		changes:    changes,
		order:      order,
	}, nil
}

func buildDestroyResources(currentState *state.State) []*config.Resource {
	resources := make([]*config.Resource, 0, len(currentState.Resources))

	networkProviderIDToLogicalID := make(map[string]string)
	for logicalID, rs := range currentState.Resources {
		if rs == nil || rs.Type != "docker_network" {
			continue
		}
		if rs.ID != "" {
			networkProviderIDToLogicalID[rs.ID] = logicalID
		}
		if networkID, ok := rs.Attributes["network_id"].(string); ok && networkID != "" {
			networkProviderIDToLogicalID[networkID] = logicalID
		}
	}

	for logicalID, rs := range currentState.Resources {
		if rs == nil {
			continue
		}

		props := make(map[string]interface{}, len(rs.Attributes))
		for k, v := range rs.Attributes {
			props[k] = v
		}

		// State stores resolved provider IDs; rebuild docker references so
		// destroy keeps dependents ahead of their dependencies.
		if rs.Type == "docker_container" {
			if networkID, ok := props["network_id"].(string); ok && networkID != "" {
				if networkLogicalID, exists := networkProviderIDToLogicalID[networkID]; exists {
					props["network_id"] = fmt.Sprintf("${docker_network.%s.network_id}", networkLogicalID)
				}
			}
		}

		resources = append(resources, &config.Resource{
			ID:         logicalID,
			Type:       rs.Type,
			Properties: props,
		})
	}

	return resources
}

func (r *Reconciler) commitApply(ctx context.Context, plan *PreparedPlan) error {
	return r.stateManager.Transact(ctx, func(tx *state.Txn) error {
		log := logger.Get()

		for _, resourceID := range plan.order {
			var resourceChanges []*Change
			for _, change := range plan.changes {
				if change.Type != ChangeTypeCreate && change.Type != ChangeTypeUpdate {
					continue
				}
				if change.Resource != nil && change.Resource.ID == resourceID {
					resourceChanges = append(resourceChanges, change)
				}
			}

			if len(resourceChanges) == 0 {
				continue
			}

			log.Info("executing changes", "resource", resourceID, "count", len(resourceChanges))
			for _, change := range resourceChanges {
				result := executeChangeWithCatalog(ctx, change, tx.State, r.catalog)
				if result.Err != nil {
					log.Error("change failed", "resource", change.Resource.ID, "error", result.Err)
					return fmt.Errorf("failed to execute change for %s: %w", change.Resource.ID, result.Err)
				}
				if result.NewState != nil {
					tx.Index.Put(change.Resource.ID, result.NewState)
				}
			}
		}

		for _, change := range plan.changes {
			if change.Type != ChangeTypeDelete {
				continue
			}
			_, err := r.catalog.Execute(ctx, provider.Request{
				Type:      change.OldState.Type,
				Action:    provider.ActionDelete,
				CurrentID: change.OldState.ID,
				Current:   change.OldState,
			})
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", change.OldState.ID, err)
			}
			tx.Index.DeleteByProviderID(change.OldState.ID)
		}

		return nil
	})
}

func (r *Reconciler) commitDestroy(ctx context.Context, plan *PreparedPlan) error {
	return r.stateManager.Transact(ctx, func(tx *state.Txn) error {
		for _, change := range plan.changes {
			if change.Type != ChangeTypeDelete {
				continue
			}

			_, err := r.catalog.Execute(ctx, provider.Request{
				Type:      change.OldState.Type,
				Action:    provider.ActionDelete,
				CurrentID: change.OldState.ID,
				Current:   change.OldState,
			})
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", change.OldState.ID, err)
			}
			tx.Index.DeleteByProviderID(change.OldState.ID)
		}
		return nil
	})
}

func executeChangeWithCatalog(
	ctx context.Context,
	change *Change,
	currentState *state.State,
	catalog *provider.Catalog,
) ExecutionResult {
	switch change.Type {
	case ChangeTypeCreate:
		resolved := InterpolateReferences(change.Resource, currentState)
		newState, err := catalog.Execute(ctx, provider.Request{
			Type:    resolved.Type,
			Action:  provider.ActionCreate,
			Desired: resolved,
		})
		if err != nil {
			return ExecutionResult{Change: change, Err: err}
		}
		return ExecutionResult{Change: change, NewState: newState}
	case ChangeTypeUpdate:
		resolved := InterpolateReferences(change.Resource, currentState)
		newState, err := catalog.Execute(ctx, provider.Request{
			Type:      resolved.Type,
			Action:    provider.ActionUpdate,
			Desired:   resolved,
			CurrentID: change.OldState.ID,
			Current:   change.OldState,
		})
		if err != nil {
			return ExecutionResult{Change: change, Err: err}
		}
		return ExecutionResult{Change: change, NewState: newState}
	case ChangeTypeDelete:
		_, err := catalog.Execute(ctx, provider.Request{
			Type:      change.OldState.Type,
			Action:    provider.ActionDelete,
			CurrentID: change.OldState.ID,
			Current:   change.OldState,
		})
		if err != nil {
			return ExecutionResult{Change: change, Err: err}
		}
		return ExecutionResult{Change: change}
	case ChangeTypeNoop:
		return ExecutionResult{Change: change, NewState: change.OldState}
	default:
		return ExecutionResult{Change: change, Err: fmt.Errorf("unknown change type: %d", change.Type)}
	}
}
