package middleware

import (
	"math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

// NewFetchdTxHandler defines a TxHandler middleware stacks based on the x/auth/middleware.NewDefaultTxHandler
// with the addition of a BLS Pubkey validation handler
func NewFetchdTxHandler(options authmiddleware.TxHandlerOptions) (tx.Handler, error) {
	if options.TxDecoder == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "txDecoder is required for middlewares")
	}

	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for middlewares")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for middlewares")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for middlewares")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = authmiddleware.DefaultSigVerificationGasConsumer
	}

	var extensionOptionChecker = options.ExtensionOptionChecker
	if extensionOptionChecker == nil {
		extensionOptionChecker = rejectExtensionOption
	}

	var txFeeChecker = options.TxFeeChecker
	if txFeeChecker == nil {
		txFeeChecker = checkTxFeeWithValidatorMinGasPrices
	}

	return authmiddleware.ComposeMiddlewares(
		authmiddleware.NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		authmiddleware.NewTxDecoderMiddleware(options.TxDecoder),
		// Set a new GasMeter on sdk.Context.
		//
		// Make sure the Gas middleware is outside of all other middlewares
		// that reads the GasMeter. In our case, the Recovery middleware reads
		// the GasMeter to populate GasInfo.
		authmiddleware.GasTxMiddleware,
		// Recover from panics. Panics outside of this middleware won't be
		// caught, be careful!
		authmiddleware.RecoveryTxMiddleware,
		// Choose which events to index in Tendermint. Make sure no events are
		// emitted outside of this middleware.
		authmiddleware.NewIndexEventsTxMiddleware(options.IndexEvents),
		// Reject all extension options other than the ones needed by the feemarket.
		authmiddleware.NewExtensionOptionsMiddleware(extensionOptionChecker),
		authmiddleware.ValidateBasicMiddleware,
		authmiddleware.TxTimeoutHeightMiddleware,
		authmiddleware.ValidateMemoMiddleware(options.AccountKeeper),
		authmiddleware.ConsumeTxSizeGasMiddleware(options.AccountKeeper),
		// No gas should be consumed in any middleware above in a "post" handler part. See
		// ComposeMiddlewares godoc for details.
		// `DeductFeeMiddleware` and `IncrementSequenceMiddleware` should be put outside of `WithBranchedStore` middleware,
		// so their storage writes are not discarded when tx fails.
		authmiddleware.DeductFeeMiddleware(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, txFeeChecker),

		// ------ custom middleware -----
		// Validate the BLS PubKeys if not already set on the accounts by the next SetPubKeyMiddleware
		BlsPubKeyValidationMiddleware(options.AccountKeeper, DefaultBlsPubkKeyValidationFunc),
		// ------

		authmiddleware.SetPubKeyMiddleware(options.AccountKeeper),
		authmiddleware.ValidateSigCountMiddleware(options.AccountKeeper),
		authmiddleware.SigGasConsumeMiddleware(options.AccountKeeper, sigGasConsumer),
		authmiddleware.SigVerificationMiddleware(options.AccountKeeper, options.SignModeHandler),
		authmiddleware.IncrementSequenceMiddleware(options.AccountKeeper),
		// Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
		// These kinds of middlewares should be put under this:
		// - Could return error after messages executed successfully.
		// - Storage writes should be discarded together when tx failed.
		authmiddleware.WithBranchedStore,
		// Consume block gas. All middlewares whose gas consumption after their `next` handler
		// should be accounted for, should go below this middleware.
		authmiddleware.ConsumeBlockGasMiddleware,
		authmiddleware.NewTipMiddleware(options.BankKeeper),
	), nil
}

// rejectExtensionOption is the default extension check that reject all tx
// extensions.
func rejectExtensionOption(*codectypes.Any) bool {
	return false
}

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, can the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// This is only for local mempool purposes, if this is a DeliverTx, the `MinGasPrices` should be zero.
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	priority := getTxPriority(feeCoins)
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func getTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
