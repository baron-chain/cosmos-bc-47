package snapshot

import (
	"fmt"

	"github.com/baron-chain/cosmos-bc-47/server"
	servertypes "github.com/baron-chain/cosmos-bc-47/server/types"
	"github.com/baron-chain/cosmos-bc-47/store/snapshots"
	"github.com/spf13/cobra"
)

func ExportSnapshotCmd(appCreator servertypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export Baron Chain state snapshot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := server.GetServerContextFromCmd(cmd)
			height, err := cmd.Flags().GetInt64("height")
			if err != nil {
				return fmt.Errorf("failed to get height flag: %w", err)
			}

			db, err := server.OpenDB(ctx.Config.RootDir, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return fmt.Errorf("failed to open DB: %w", err)
			}
			defer db.Close()

			app := appCreator(ctx.Logger, db, nil, ctx.Viper)
			sm := app.SnapshotManager()
			if sm == nil {
				return fmt.Errorf("snapshot manager not configured")
			}

			if height == 0 {
				height = int64(app.LastBlockHeight())
			}

			if height <= 0 {
				return fmt.Errorf("height must be positive")
			}

			snapshot, err := sm.Create(uint64(height))
			if err != nil {
				return fmt.Errorf("failed to create snapshot: %w", err)
			}

			cmd.Printf("Snapshot created successfully:\nHeight: %d\nFormat: %d\nChunks: %d\nHash: %X\n",
				snapshot.Height, snapshot.Format, snapshot.Chunks, snapshot.Hash)
			return nil
		},
	}
}

type SnapshotOptions struct {
	Height  uint64
	Format  uint32
	Chunks  uint32
	Hash    []byte
	Version string
}

func ValidateSnapshot(snapshot *snapshots.Snapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot is nil")
	}
	if snapshot.Height == 0 {
		return fmt.Errorf("invalid height")
	}
	if snapshot.Format == 0 {
		return fmt.Errorf("invalid format")
	}
	if snapshot.Chunks == 0 {
		return fmt.Errorf("invalid chunks")
	}
	if len(snapshot.Hash) == 0 {
		return fmt.Errorf("invalid hash")
	}
	return nil
}
