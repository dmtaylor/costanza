package report

import (
	"context"
	"log"

	"github.com/pkg/errors"
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

func removeStats(cmd *cobra.Command, args []string) error {
	app, err := config.LoadApp()
	if err != nil {
		return errors.Wrap(err, "failed to load app state")
	}
	month, err := cmd.Flags().GetString("month")
	if err != nil {
		return errors.Wrap(err, "error getting month")
	}
	log.Printf("Deleting usage activity for month %s...", month)
	err = app.Stats.RemoveMonthActivity(context.Background(), month)
	if err != nil {
		return errors.Wrap(err, "failed to remove stats for month "+month)
	}
	log.Printf("Successfully removed stats for %s", month)

	return nil
}
