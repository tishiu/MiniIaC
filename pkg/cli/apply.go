package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tishiu/MiniIac/pkg/reconciler"
)

func (c *CLI) Apply(configPath string, autoApprove bool) error {
	desired, err := c.parser.Parse(configPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	plan, err := c.reconciler.Prepare(context.Background(), reconciler.Request{
		Mode:    reconciler.ModeApply,
		Desired: desired,
	})
	if err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}
	defer plan.Discard()

	c.printPlan(plan.Changes())

	if !autoApprove {
		fmt.Print("\nDo you want to apply these changes? (yes/no): ")
		var response string
		fmt.Scanln(&response)

		if response != "yes" {
			fmt.Println("Apply cancelled.")
			return nil
		}
	}

	fmt.Println("\nApplying changes...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Handle graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, cancelling apply...")
		cancel()
	}()
	defer signal.Stop(sigChan)

	if err := plan.Commit(ctx); err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	fmt.Println("\nInfrastructure updated successfully!")

	return nil
}
