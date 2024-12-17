package snapshot

import (
	"github.com/spf13/cobra"
	servertypes "github.com/baron-chain/cosmos-bc-47/server/types"
)

const (
	snapshotCmdName        = "snapshots"
	snapshotCmdShortDesc   = "Manage Baron Chain local snapshots"
	snapshotCmdLongDesc    = `Manage local snapshots for Baron Chain state sync and backup.
This command provides functionality to list, restore, export, and manage snapshots.`
)

// Cmd returns the snapshots management command group for Baron Chain
func Cmd(appCreator servertypes.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:     snapshotCmdName,
		Short:   snapshotCmdShortDesc,
		Long:    snapshotCmdLongDesc,
		Example: getExamples(),
	}

	cmd.AddCommand(
		ListSnapshotsCmd,
		RestoreSnapshotCmd(appCreator),
		ExportSnapshotCmd(appCreator),
		DumpArchiveCmd(),
		LoadArchiveCmd(),
		DeleteSnapshotCmd(),
	)

	return cmd
}

func getExamples() string {
	return `  # List all available snapshots
  barond snapshots list

  # Export a snapshot at a specific height
  barond snapshots export --height 1000000

  # Restore from a snapshot
  barond snapshots restore <snapshot-file>

  # Dump snapshot to archive
  barond snapshots dump <snapshot-name>

  # Load snapshot from archive
  barond snapshots load <archive-name>

  # Delete a snapshot
  barond snapshots delete <snapshot-name>`
}
