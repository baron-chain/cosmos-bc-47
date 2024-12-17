package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	rpchttp "github.com/baron-chain/cometbft-bc/rpc/client/http"
	coretypes "github.com/baron-chain/cometbft-bc/rpc/core/types"
	tmtypes "github.com/baron-chain/cometbft-bc/types"
	"github.com/baron-chain/cosmos-bc-47/client"
	"github.com/baron-chain/cosmos-bc-47/client/flags"
	sdk "github.com/baron-chain/cosmos-bc-47/types"
	"github.com/baron-chain/cosmos-bc-47/types/errors"
)

const (
	defaultTimeout   = 15 * time.Second
	subscriberID    = "baron-chain-subscriber"
	websocketPath   = "/websocket"
)

func createTxResponse(res *coretypes.ResultBroadcastTxCommit, txResult *tmtypes.ResponseDeliverTx, hash []byte) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	txHash := ""
	if hash != nil {
		txHash = hex.EncodeToString(hash)
	}

	parsedLogs, _ := sdk.ParseABCILogs(txResult.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: txResult.Codespace,
		Code:      txResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(txResult.Data)),
		RawLog:    txResult.Log,
		Logs:      parsedLogs,
		Info:      txResult.Info,
		GasWanted: txResult.GasWanted,
		GasUsed:   txResult.GasUsed,
		Events:    txResult.Events,
	}
}

func createBroadcastTxResponse(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	if !res.CheckTx.IsOK() {
		return createTxResponse(res, &res.CheckTx, res.Hash)
	}
	return createTxResponse(res, &res.DeliverTx, res.Hash)
}

func QueryEventForTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "event-query-tx [hash]",
		Short:   "Query for a Baron Chain transaction by hash",
		Example: "$ barond query event-query-tx 0x123...",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get client context: %w", err)
			}

			txHash := args[0]
			return queryTxEvent(cmd.Context(), clientCtx, txHash)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func queryTxEvent(ctx context.Context, clientCtx client.Context, txHash string) error {
	wsClient, err := rpchttp.New(clientCtx.NodeURI, websocketPath)
	if err != nil {
		return fmt.Errorf("failed to create websocket client: %w", err)
	}

	if err := wsClient.Start(); err != nil {
		return fmt.Errorf("failed to start websocket client: %w", err)
	}
	defer wsClient.Stop() //nolint:errcheck

	queryCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	query := fmt.Sprintf("%s='%s' AND %s='%s'", 
		tmtypes.EventTypeKey, 
		tmtypes.EventTx, 
		tmtypes.TxHashKey, 
		txHash,
	)

	eventCh, err := wsClient.Subscribe(queryCtx, subscriberID, query)
	if err != nil {
		return fmt.Errorf("failed to subscribe to tx events: %w", err)
	}
	defer wsClient.UnsubscribeAll(context.Background(), subscriberID) //nolint:errcheck

	select {
	case evt := <-eventCh:
		txEvent, ok := evt.Data.(tmtypes.EventDataTx)
		if !ok {
			return fmt.Errorf("received invalid event data type: %T", evt.Data)
		}

		res := &coretypes.ResultBroadcastTxCommit{
			DeliverTx: txEvent.Result,
			Hash:      tmtypes.Tx(txEvent.Tx).Hash(),
			Height:    txEvent.Height,
		}

		return clientCtx.PrintProto(createBroadcastTxResponse(res))

	case <-queryCtx.Done():
		return errors.ErrLogic.Wrap("timed out waiting for transaction event")
	}
}
