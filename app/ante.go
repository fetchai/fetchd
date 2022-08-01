package app

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),

		// BlsPubKeyValidationDecorator must be called before SetPubKeyDecorator
		NewBlsPubKeyValidationDecorator(options.AccountKeeper),

		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// BlsPubKeyValidationDecorator will validate new BLS pubkeys when they're not already set on the account
// it will only happen once and for new keys, as the following SetPubKeyDecorator call will set the pubkey on the account
type BlsPubKeyValidationDecorator struct {
	ak ante.AccountKeeper
}

func NewBlsPubKeyValidationDecorator(ak ante.AccountKeeper) BlsPubKeyValidationDecorator {
	return BlsPubKeyValidationDecorator{
		ak: ak,
	}
}

func (d BlsPubKeyValidationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	pubkeys, err := sigTx.GetPubKeys()
	if err != nil {
		return ctx, err
	}
	signers := sigTx.GetSigners()

	for i, pk := range pubkeys {
		if pk == nil {
			continue
		}

		// Only make check if simulate=false
		if !simulate && !bytes.Equal(pk.Address(), signers[i]) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey,
				"pubKey does not match signer address %s with signer index: %d", signers[i], i)
		}

		acc, err := ante.GetSignerAcc(ctx, d.ak, signers[i])
		if err != nil {
			return ctx, err
		}
		// account already has pubkey set,no need to reset
		if acc.GetPubKey() != nil {
			continue
		}

		// Validate public key for bls12381 so that it only needs to be checked once
		// next time it will be set on the account by the SetPubKeyDecorator
		pubkey, ok := pk.(*bls12381.PubKey)
		if ok {
			if !pubkey.Validate() {
				return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "Invalid public key: either infinity or not subgroup element")
			}
		}
	}

	return next(ctx, tx, simulate)
}

// BlsDefaultSigVerificationGasConsumer extends the default ante.DefaultSigVerificationGasConsumer to add support fo bls pubkeys.
func BlsDefaultSigVerificationGasConsumer(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
	switch sig.PubKey.(type) {
	case *bls12381.PubKey:
		meter.ConsumeGas(bls12381.DefaultSigVerifyCostBls12381, "ante verify: bls12381")
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}
