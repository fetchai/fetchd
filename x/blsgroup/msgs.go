package blsgroup

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
)

var _ sdk.Msg = &MsgVoteAgg{}
var _ sdk.Msg = &MsgRegisterBlsGroup{}
var _ sdk.Msg = &MsgUnregisterBlsGroup{}

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

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m MsgRegisterBlsGroup) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrInvalid, "group_id")
	}

	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	return nil
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
func (m MsgRegisterBlsGroup) GetSigners() []types.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{addr}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m MsgUnregisterBlsGroup) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrInvalid, "group_id")
	}

	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	return nil
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
func (m MsgUnregisterBlsGroup) GetSigners() []types.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
