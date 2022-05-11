package middleware

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

type BlsPubKeyValidationFunc func(pk *bls12381.PubKey) error

func DefaultBlsPubkKeyValidationFunc(pk *bls12381.PubKey) error {
	if !pk.Validate() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "invalid public key: either infinity or not subgroup element")
	}
	return nil
}

// BlsPubKeyValidationMiddleware creates a new middleware validating BLS public keys
// using provided BlsPubKeyValidationFunc. When BlsPubKeyValidationFunc is nil, it uses the DefaultBlsPubkKeyValidationFunc.
// This middleware must be placed before the SetPubKeyMiddleware as the validation is only done
// for new public keys, which are not already set on the accounts.
func BlsPubKeyValidationMiddleware(ak authmiddleware.AccountKeeper, valFn BlsPubKeyValidationFunc) tx.Middleware {
	if valFn == nil {
		valFn = DefaultBlsPubkKeyValidationFunc
	}

	return func(txh tx.Handler) tx.Handler {
		return blsPubKeyValidationTxHandler{
			ak:    ak,
			valFn: valFn,
			next:  txh,
		}
	}
}

// blsPubKeyValidationTxHandler will validate new BLS pubkeys when they're not already set on the account
// it will only happen once and for new keys, as the following SetPubKeyMiddleware call will set the pubkey on the account
type blsPubKeyValidationTxHandler struct {
	ak    authmiddleware.AccountKeeper
	valFn BlsPubKeyValidationFunc

	next tx.Handler
}

var _ tx.Handler = blsPubKeyValidationTxHandler{}

// CheckTx implements tx.Handler.CheckTx.
func (h blsPubKeyValidationTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	if err := h.verifyPubKey(ctx, req, false); err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return h.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (h blsPubKeyValidationTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := h.verifyPubKey(ctx, req, false); err != nil {
		return tx.Response{}, err
	}
	return h.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (h blsPubKeyValidationTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := h.verifyPubKey(ctx, req, true); err != nil {
		return tx.Response{}, err
	}
	return h.next.SimulateTx(ctx, req)
}

func (h blsPubKeyValidationTxHandler) verifyPubKey(ctx context.Context, req tx.Request, simulate bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sigTx, ok := req.Tx.(authsigning.SigVerifiableTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	pubkeys, err := sigTx.GetPubKeys()
	if err != nil {
		return err
	}
	signers := sigTx.GetSigners()

	for i, pk := range pubkeys {
		if pk == nil {
			continue
		}

		// Only make check if simulate=false
		if !simulate && !bytes.Equal(pk.Address(), signers[i]) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey,
				"pubKey does not match signer address %s with signer index: %d", signers[i], i)
		}

		acc, err := authmiddleware.GetSignerAcc(sdkCtx, h.ak, signers[i])
		if err != nil {
			return err
		}
		// account already has pubkey set, no need to validate again
		if acc.GetPubKey() != nil {
			continue
		}

		// Validate public key for bls12381 so that it only needs to be checked once
		// next time it will be set on the account by the SetPubKeyMiddleware
		pubkey, ok := pk.(*bls12381.PubKey)
		if ok {
			if err := h.valFn(pubkey); err != nil {
				return err
			}
		}
	}

	return nil
}
