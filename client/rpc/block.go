package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/baron-chain/cosmos-bc-47/client"
	"github.com/baron-chain/cosmos-bc-47/client/flags"
	"github.com/baron-chain/cometbft-bc/types"
)

const (
	defaultNodeEndpoint = "tcp://localhost:26657"
	flagNode           = "node"
)

func BlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block [height]",
		Short:   "Get verified data for the Baron Chain block at given height",
		Example: "$ barond query block 100",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get query context: %w", err)
			}

			height, err := parseHeight(args)
			if err != nil {
				return err
			}

			output, err := fetchBlockData(clientCtx, height)
			if err != nil {
				return fmt.Errorf("failed to fetch block data: %w", err)
			}

			return clientCtx.PrintBytes(output)
		},
	}

	cmd.Flags().StringP(flagNode, "n", defaultNodeEndpoint, "Baron Chain node to connect to")
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func parseHeight(args []string) (*int64, error) {
	if len(args) == 0 {
		return nil, nil
	}

	height, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid height '%s': %w", args[0], err)
	}

	if height <= 0 {
		return nil, fmt.Errorf("height must be greater than 0, got %d", height)
	}

	return &height, nil
}

func fetchBlockData(clientCtx client.Context, height *int64) (json.RawMessage, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	block, err := node.Block(ctx, height)
	if err != nil {
		return nil, err
	}

	return json.Marshal(block)
}

func GetChainHeight(clientCtx client.Context) (int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return -1, fmt.Errorf("failed to get node: %w", err)
	}

	ctx := context.Background()
	status, err := node.Status(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to get node status: %w", err)
	}

	return status.SyncInfo.LatestBlockHeight, nil
}
