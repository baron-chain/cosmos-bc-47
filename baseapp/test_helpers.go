package baseapp
//BC MOD
import (
	"fmt"

	tmproto "github.com/baron-chain/cometbft-bc/proto/tendermint/types"
	sdk "github.com/baron-chain/cosmos-bc-47/types"
	sdkerrors "github.com/baron-chain/cosmos-bc-47/types/errors"
)

// SimCheck performs a CheckTx simulation and returns gas info and result.
// It's used primarily in tests and simulations.
func (app *BaseApp) SimCheck(txEncoder sdk.TxEncoder, tx sdk.Tx) (gasInfo sdk.GasInfo, result *sdk.Result, err error) {
	txBytes, err := encodeTx(txEncoder, tx)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}
	
	return app.runTxSimulation(runTxModeCheck, txBytes)
}

// Simulate executes a transaction in simulation mode to estimate gas usage.
func (app *BaseApp) Simulate(txBytes []byte) (gasInfo sdk.GasInfo, result *sdk.Result, err error) {
	return app.runTxSimulation(runTxModeSimulate, txBytes)
}

// SimDeliver simulates a transaction delivery and returns execution results.
func (app *BaseApp) SimDeliver(txEncoder sdk.TxEncoder, tx sdk.Tx) (gasInfo sdk.GasInfo, result *sdk.Result, err error) {
	txBytes, err := encodeTx(txEncoder, tx)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	return app.runTxSimulation(runTxModeDeliver, txBytes)
}

// NewContext creates a new context for transaction processing.
func (app *BaseApp) NewContext(isCheckTx bool, header tmproto.Header) sdk.Context {
	if isCheckTx {
		return sdk.NewContext(app.checkState.ms, header, true, app.logger).
			WithMinGasPrices(app.minGasPrices)
	}
	
	return sdk.NewContext(app.deliverState.ms, header, false, app.logger)
}

// NewUncachedContext creates a new context without caching.
func (app *BaseApp) NewUncachedContext(isCheckTx bool, header tmproto.Header) sdk.Context {
	return sdk.NewContext(app.cms, header, isCheckTx, app.logger)
}

// GetContextForDeliverTx returns the context for transaction delivery.
func (app *BaseApp) GetContextForDeliverTx(txBytes []byte) sdk.Context {
	return app.getContextForTx(runTxModeDeliver, txBytes)
}

// Helper functions

// encodeTx encodes a transaction using the provided encoder.
func encodeTx(txEncoder sdk.TxEncoder, tx sdk.Tx) ([]byte, error) {
	txBytes, err := txEncoder(tx)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to encode tx: %v", err)
	}
	return txBytes, nil
}

// runTxSimulation executes a transaction in the specified mode and returns relevant info.
func (app *BaseApp) runTxSimulation(mode runTxMode, txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	gasInfo, result, _, _, err := app.runTx(mode, txBytes)
	return gasInfo, result, err
}
