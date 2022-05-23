package blsgroup

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
)

var _ sdk.Msg = &MsgVoteAgg{}

// Route Implements Msg.
func (m MsgVoteAgg) Route() string {
	return sdk.MsgTypeURL(&m)
}

// Type Implements Msg.
func (m MsgVoteAgg) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteAgg) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVoteAgg.
func (m MsgVoteAgg) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteAgg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(err, "sender")
	}

	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}

	if len(m.Votes) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "votes")
	}
	for _, c := range m.Votes {
		if _, ok := group.VoteOption_name[int32(c)]; !ok {
			return sdkerrors.Wrap(ErrInvalid, "choice")
		}
	}

	if len(m.AggSig) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "voter signature")
	}

	return nil
}
