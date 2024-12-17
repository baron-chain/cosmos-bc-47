package snapshot

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/baron-chain/cosmos-bc-47/server"
)

const (
	deleteCommandUse     = "delete <height> <format>"
	deleteCommandShort   = "Delete a local Baron Chain snapshot"
	deleteCommandLong    = `Delete a local snapshot from the Baron Chain node.

Arguments:
  height - The height of the snapshot to delete
  format - The format number of the snapshot to delete (usually 1 or 2)`
	deleteCommandExample = `  # Delete a snapshot at height 1000000 with format 1
  barond snapshots delete 1000000 1

  # Delete a snapshot at height 2000000 with format 2
  barond snapshots delete 2000000 2`
)

// DeleteSnapshotCmd returns a command to delete local snapshots
func DeleteSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     deleteCommandUse,
		Short:   deleteCommandShort,
		Long:    deleteCommandLong,
		Example: deleteCommandExample,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			height, err := parseHeight(args[0])
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}

			format, err := parseFormat(args[1])
			if err != nil {
				return fmt.Errorf("invalid format: %w", err)
			}

			ctx := server.GetServerContextFromCmd(cmd)
			snapshotStore, err := server.GetSnapshotStore(ctx.Viper)
			if err != nil {
				return fmt.Errorf("failed to get snapshot store: %w", err)
			}

			if err := snapshotStore.Delete(height, format); err != nil {
				return fmt.Errorf("failed to delete snapshot at height %d format %d: %w", height, format, err)
			}

			cmd.Printf("Successfully deleted snapshot at height %d format %d\n", height, format)
			return nil
		},
	}

	return cmd
}

func parseHeight(heightStr string) (uint64, error) {
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if height == 0 {
		return 0, fmt.Errorf("height must be greater than 0")
	}
	return height, nil
}

func parseFormat(formatStr string) (uint32, error) {
	format, err := strconv.ParseUint(formatStr, 10, 32)
	if err != nil {
		return 0, err
	}
	if format == 0 {
		return 0, fmt.Errorf("format must be greater than 0")
	}
	return uint32(format), nil
}
