package cli

import (
	"context"
	"fmt"

	"github.com/tishiu/MiniIac/pkg/reconciler"
)

func (c *CLI) Destroy(autoApprove bool) error {
	plan, err := c.reconciler.Prepare(context.Background(), reconciler.Request{
		Mode: reconciler.ModeDestroy,
	})
	if err != nil {
		return fmt.Errorf("failed to prepare destroy: %w", err)
	}
	defer plan.Discard()

	if len(plan.Changes()) == 0 {
		fmt.Println("No resources to destroy.")
		return nil
	}

	// Display resources to be destroyed
	fmt.Println("\n=== Resources to Destroy ===")
	for _, change := range plan.Changes() {
		if change.Type != reconciler.ChangeTypeDelete {
			continue
		}
		id := change.OldState.ID
		typ := change.OldState.Type
		if change.Resource != nil {
			id = change.Resource.ID
			typ = change.Resource.Type
		}
		fmt.Printf("  - %s (%s)\n", id, typ)
	}

	if !autoApprove {
		fmt.Print("\nDo you want to destroy all resources? (yes/no): ")
		var response string
		fmt.Scanln(&response)

		if response != "yes" {
			fmt.Println("Destroy cancelled.")
			return nil
		}
	}

	fmt.Println("\nDestroying resources...")

	if err := plan.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to destroy resources: %w", err)
	}

	fmt.Println("\nAll resources destroyed successfully!")

	return nil
}
