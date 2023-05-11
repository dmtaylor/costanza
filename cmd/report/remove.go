package report

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/dmtaylor/costanza/config"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove stats for specified month",
	Long:  `Deletes stats for specified month when no longer needed`,
	RunE:  removeStats,
}

func removeStats(cmd *cobra.Command, _ []string) error {
	app, err := config.LoadApp()
	if err != nil {
		return fmt.Errorf("failed to load app state: %w", err)
	}
	month, err := cmd.Flags().GetString("month")
	if err != nil {
		return fmt.Errorf("error getting month: %w", err)
	}
	log.Printf("Deleting usage activity for month %s...", month)
	err = app.Stats.RemoveMonthActivity(context.Background(), month)
	if err != nil {
		return fmt.Errorf("failed to remove stats for month %s: %w", month, err)
	}
	log.Printf("Successfully removed stats for %s", month)

	return nil
}
