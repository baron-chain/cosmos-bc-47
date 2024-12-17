package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/baron-chain/cometbft-bc/libs/bytes"
	"github.com/baron-chain/cometbft-bc/p2p"
	coretypes "github.com/baron-chain/cometbft-bc/rpc/core/types"
	"github.com/baron-chain/cosmos-bc-47/client"
	"github.com/baron-chain/cosmos-bc-47/client/flags"
	cryptocodec "github.com/baron-chain/cosmos-bc-47/crypto/codec"
	cryptotypes "github.com/baron-chain/cosmos-bc-47/crypto/types"
)

const (
	defaultNodeEndpoint = "tcp://localhost:26657"
	flagNode           = "node"
)

type ValidatorInfo struct {
	Address     bytes.HexBytes      `json:"address"`
	PubKey      cryptotypes.PubKey  `json:"pub_key"`
	VotingPower int64              `json:"voting_power"`
}

type NodeStatus struct {
	NodeInfo      p2p.DefaultNodeInfo `json:"node_info"`
	SyncInfo      coretypes.SyncInfo  `json:"sync_info"`
	ValidatorInfo ValidatorInfo       `json:"validator_info"`
}

func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Query Baron Chain node status",
		Example: "$ barond query status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get query context: %w", err)
			}

			status, err := queryNodeStatus(clientCtx)
			if err != nil {
				return err
			}

			pubKey, err := convertValidatorPubKey(status)
			if err != nil {
				return err
			}

			nodeStatus := NodeStatus{
				NodeInfo: status.NodeInfo,
				SyncInfo: status.SyncInfo,
				ValidatorInfo: ValidatorInfo{
					Address:     status.ValidatorInfo.Address,
					PubKey:     pubKey,
					VotingPower: status.ValidatorInfo.VotingPower,
				},
			}

			output, err := json.MarshalIndent(nodeStatus, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal status: %w", err)
			}

			return clientCtx.PrintBytes(output)
		},
	}

	cmd.Flags().StringP(flagNode, "n", defaultNodeEndpoint, "Baron Chain node to connect to")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func queryNodeStatus(clientCtx client.Context) (*coretypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to query node status: %w", err)
	}

	return status, nil
}

func convertValidatorPubKey(status *coretypes.ResultStatus) (cryptotypes.PubKey, error) {
	if status.ValidatorInfo.PubKey == nil {
		return nil, nil
	}

	pubKey, err := cryptocodec.FromTmPubKeyInterface(status.ValidatorInfo.PubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert validator public key: %w", err)
	}

	return pubKey, nil
}
