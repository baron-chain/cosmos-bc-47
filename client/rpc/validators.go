package rpc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	tmtypes "github.com/baron-chain/cometbft-bc/types"

	"github.com/baron-chain/cosmos-bc-47/client"
	"github.com/baron-chain/cosmos-bc-47/client/flags"
	cryptocodec "github.com/baron-chain/cosmos-bc-47/crypto/codec"
	cryptotypes "github.com/baron-chain/cosmos-bc-47/crypto/types"
	sdk "github.com/baron-chain/cosmos-bc-47/types"
	"github.com/baron-chain/cosmos-bc-47/types/query"
)

const (
	defaultNodeEndpoint = "tcp://localhost:26657"
	defaultLimit       = 100
)

type ValidatorOutput struct {
	Address          sdk.ConsAddress    `json:"address"`
	PubKey           cryptotypes.PubKey `json:"pub_key"`
	ProposerPriority int64             `json:"proposer_priority"`
	VotingPower      int64             `json:"voting_power"`
}

type ValidatorsOutput struct {
	BlockHeight int64             `json:"block_height"`
	Validators  []ValidatorOutput `json:"validators"`
	Total       uint64            `json:"total"`
}

func (vo ValidatorsOutput) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Block Height: %d\n", vo.BlockHeight)
	fmt.Fprintf(&b, "Total Validators: %d\n", vo.Total)

	for _, val := range vo.Validators {
		fmt.Fprintf(&b, "\nValidator Details:\n")
		fmt.Fprintf(&b, "  Address:           %s\n", val.Address)
		fmt.Fprintf(&b, "  Public Key:        %s\n", val.PubKey)
		fmt.Fprintf(&b, "  Proposer Priority: %d\n", val.ProposerPriority)
		fmt.Fprintf(&b, "  Voting Power:      %d\n", val.VotingPower)
	}

	return b.String()
}

func ValidatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validator-set [height]",
		Short:   "Get Baron Chain validator set at a given height",
		Example: "$ barond query validator-set 1000",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get query context: %w", err)
			}

			height, err := parseOptionalHeight(args)
			if err != nil {
				return err
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			result, err := QueryValidators(cmd.Context(), clientCtx, height, &page, &limit)
			if err != nil {
				return fmt.Errorf("failed to query validators: %w", err)
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	cmd.Flags().String(flags.FlagNode, defaultNodeEndpoint, "Baron Chain node RPC endpoint")
	cmd.Flags().StringP(flags.FlagOutput, "o", "text", "Output format (text|json)")
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Page number for paginated results")
	cmd.Flags().Int(flags.FlagLimit, defaultLimit, "Number of results per page")

	return cmd
}

func parseOptionalHeight(args []string) (*int64, error) {
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

func convertValidatorOutput(validator *tmtypes.Validator) (ValidatorOutput, error) {
	pubKey, err := cryptocodec.FromTmPubKeyInterface(validator.PubKey)
	if err != nil {
		return ValidatorOutput{}, fmt.Errorf("failed to convert validator public key: %w", err)
	}

	return ValidatorOutput{
		Address:          sdk.ConsAddress(validator.Address),
		PubKey:           pubKey,
		ProposerPriority: validator.ProposerPriority,
		VotingPower:      validator.VotingPower,
	}, nil
}

func QueryValidators(ctx context.Context, clientCtx client.Context, height *int64, page, limit *int) (ValidatorsOutput, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return ValidatorsOutput{}, fmt.Errorf("failed to get node: %w", err)
	}

	validatorsRes, err := node.Validators(ctx, height, page, limit)
	if err != nil {
		return ValidatorsOutput{}, fmt.Errorf("failed to query validators: %w", err)
	}

	total := uint64(0)
	if validatorsRes.Total > 0 {
		total = uint64(validatorsRes.Total)
	}

	validators := make([]ValidatorOutput, len(validatorsRes.Validators))
	for i, validator := range validatorsRes.Validators {
		validators[i], err = convertValidatorOutput(validator)
		if err != nil {
			return ValidatorsOutput{}, err
		}
	}

	return ValidatorsOutput{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  validators,
		Total:       total,
	}, nil
}
