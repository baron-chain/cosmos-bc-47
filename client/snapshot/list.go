package snapshot

import (
	"fmt"
	"time"

	"github.com/baron-chain/cosmos-bc-47/server"
	"github.com/baron-chain/cosmos-bc-47/store/snapshots"
	"github.com/spf13/cobra"
)

type SnapshotInfo struct {
	Height    uint64
	Format    uint32
	Chunks    uint32
	Hash      []byte
	Timestamp time.Time
	Size      int64
}

func NewListSnapshotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available Baron Chain snapshots",
		Long:  "Display all available snapshots with detailed information including height, format, chunks, and size",
		RunE:  listSnapshots,
	}
	
	cmd.Flags().BoolP("detail", "d", false, "Show detailed snapshot information")
	cmd.Flags().Int64P("min-height", "m", 0, "Minimum height filter")
	return cmd
}

func listSnapshots(cmd *cobra.Command, args []string) error {
	ctx := server.GetServerContextFromCmd(cmd)
	showDetail, _ := cmd.Flags().GetBool("detail")
	minHeight, _ := cmd.Flags().GetInt64("min-height")

	store, err := server.GetSnapshotStore(ctx.Viper)
	if err != nil {
		return fmt.Errorf("failed to access snapshot store: %w", err)
	}

	snapshots, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		cmd.Println("No snapshots found")
		return nil
	}

	cmd.Println("Available snapshots:")
	cmd.Println("-------------------")

	for _, snap := range snapshots {
		if int64(snap.Height) < minHeight {
			continue
		}

		info := formatSnapshotInfo(snap, showDetail)
		if showDetail {
			cmd.Printf("Height: %d\nFormat: %d\nChunks: %d\nHash: %X\nSize: %d bytes\nTimestamp: %s\n-------------------\n",
				snap.Height, snap.Format, snap.Chunks, snap.Hash, snap.Metadata.Size, snap.Metadata.Timestamp)
		} else {
			cmd.Printf("Height: %d | Format: %d | Chunks: %d\n", 
				snap.Height, snap.Format, snap.Chunks)
		}
	}

	return nil
}

func formatSnapshotInfo(snap *snapshots.Snapshot, detailed bool) string {
	if !detailed {
		return fmt.Sprintf("Height: %d, Format: %d, Chunks: %d", 
			snap.Height, snap.Format, snap.Chunks)
	}

	return fmt.Sprintf("Height: %d\nFormat: %d\nChunks: %d\nHash: %X\nSize: %d bytes\nTimestamp: %s",
		snap.Height, snap.Format, snap.Chunks, snap.Hash, snap.Metadata.Size, snap.Metadata.Timestamp)
}

func validateSnapshot(snap *snapshots.Snapshot) error {
	if snap == nil {
		return fmt.Errorf("invalid snapshot: nil")
	}

	if snap.Height == 0 {
		return fmt.Errorf("invalid snapshot height: 0")
	}

	if snap.Format == 0 {
		return fmt.Errorf("invalid snapshot format: 0")
	}

	if snap.Chunks == 0 {
		return fmt.Errorf("invalid snapshot chunks: 0")
	}

	if len(snap.Hash) == 0 {
		return fmt.Errorf("invalid snapshot hash: empty")
	}

	return nil
}
