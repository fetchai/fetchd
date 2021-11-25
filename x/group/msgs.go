package group

import (
	bytes "bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"

	"github.com/fetchai/fetchd/types/math"
	"github.com/fetchai/fetchd/types/module/server"
	prototypes "github.com/gogo/protobuf/types"
)

var _ sdk.Msg = &MsgCreateGroup{}

// Route Implements Msg.
func (m MsgCreateGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if err := assertMetadataLength(m.Metadata, "group metadata"); err != nil {
		return err
	}

	members := Members{Members: m.Members}
	if err := members.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "members")
	}
	for i := range m.Members {
		member := m.Members[i]
		if _, err := math.NewPositiveDecFromString(member.Weight); err != nil {
			return sdkerrors.Wrap(err, "member weight")
		}
	}
	return nil
}

func (m Member) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	if _, err := math.NewNonNegativeDecFromString(m.Weight); err != nil {
		return sdkerrors.Wrap(err, "weight")
	}

	if err := assertMetadataLength(m.Metadata, "member metadata"); err != nil {
		return err
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupAdmin) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAdmin.
func (m MsgUpdateGroupAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAdmin) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")
	}

	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(ErrInvalid, "new and old admin are the same")
	}
	return nil
}

func (m *MsgUpdateGroupAdmin) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupMetadata) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupMetadata.
func (m MsgUpdateGroupMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMetadata) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	if err = assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}
	return nil
}

func (m *MsgUpdateGroupMetadata) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}

// Route Implements Msg.
func (m MsgUpdateGroupMembers) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupMembers) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMembers) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}

// GetSigners returns the expected signers for a MsgUpdateGroupMembers.
func (m MsgUpdateGroupMembers) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMembers) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if len(m.MemberUpdates) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "member updates")
	}
	members := Members{Members: m.MemberUpdates}
	if err := members.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "members")
	}
	return nil
}

func (m *MsgUpdateGroupMembers) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgCreateGroupAccount{}

// Route Implements Msg.
func (m MsgCreateGroupAccount) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroupAccount) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroupAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupAccount.
func (m MsgCreateGroupAccount) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroupAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")
	}

	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAccountAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupAccountAdmin) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountAdmin.
func (m MsgUpdateGroupAccountAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountAdmin) ValidateBasic() error {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(ErrInvalid, "new and old admin are the same")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAccountDecisionPolicy{}
var _ types.UnpackInterfacesMessage = MsgUpdateGroupAccountDecisionPolicy{}

func NewMsgUpdateGroupAccountDecisionPolicyRequest(admin sdk.AccAddress, address sdk.AccAddress, decisionPolicy DecisionPolicy) (*MsgUpdateGroupAccountDecisionPolicy, error) {
	m := &MsgUpdateGroupAccountDecisionPolicy{
		Admin:   admin.String(),
		Address: address.String(),
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MsgUpdateGroupAccountDecisionPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// Route Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) Type() string {
	return sdk.MsgTypeURL(&m)
}

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountDecisionPolicy.
func (m MsgUpdateGroupAccountDecisionPolicy) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountDecisionPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}

	return nil
}

func (m *MsgUpdateGroupAccountDecisionPolicy) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgUpdateGroupAccountDecisionPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var _ sdk.Msg = &MsgUpdateGroupAccountMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupAccountMetadata) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountMetadata.
func (m MsgUpdateGroupAccountMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if err := assertMetadataLength(m.Metadata, "group account metadata"); err != nil {
		return err
	}

	return nil
}

var _ sdk.Msg = &MsgCreateGroupAccount{}
var _ types.UnpackInterfacesMessage = MsgCreateGroupAccount{}

// NewMsgCreateGroupAccount creates a new MsgCreateGroupAccount.
func NewMsgCreateGroupAccount(admin sdk.AccAddress, group uint64, metadata []byte, decisionPolicy DecisionPolicy) (*MsgCreateGroupAccount, error) {
	m := &MsgCreateGroupAccount{
		Admin:    admin.String(),
		GroupId:  group,
		Metadata: metadata,
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MsgCreateGroupAccount) GetAdmin() string {
	return m.Admin
}

func (m *MsgCreateGroupAccount) GetGroupID() uint64 {
	return m.GroupId
}

func (m *MsgCreateGroupAccount) GetMetadata() []byte {
	return m.Metadata
}

func (m *MsgCreateGroupAccount) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

func (m *MsgCreateGroupAccount) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateGroupAccount) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var _ sdk.Msg = &MsgCreateProposal{}

// NewMsgCreateProposalRequest creates a new MsgCreateProposal.
func NewMsgCreateProposalRequest(address string, proposers []string, msgs []sdk.Msg, metadata []byte, exec Exec) (*MsgCreateProposal, error) {
	m := &MsgCreateProposal{
		Address:   address,
		Proposers: proposers,
		Metadata:  metadata,
		Exec:      exec,
	}
	err := m.SetMsgs(msgs)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Route Implements Msg.
func (m MsgCreateProposal) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateProposal) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateProposal.
func (m MsgCreateProposal) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			panic(err)
		}
		addrs[i] = addr
	}
	return addrs
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}

	if len(m.Proposers) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposers")
	}
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			return sdkerrors.Wrap(err, "proposers")
		}
		addrs[i] = addr
	}
	if err := AccAddresses(addrs).ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proposers")
	}

	msgs := m.GetMsgs()
	for i, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "msg %d", i)
		}
	}

	return nil
}

// SetMsgs packs msgs into Any's
func (m *MsgCreateProposal) SetMsgs(msgs []sdk.Msg) error {
	anys, err := server.SetMsgs(msgs)
	if err != nil {
		return err
	}
	m.Msgs = anys
	return nil
}

// GetMsgs unpacks m.Msgs Any's into sdk.Msg's
func (m MsgCreateProposal) GetMsgs() []sdk.Msg {
	msgs, err := server.GetMsgs(m.Msgs)
	if err != nil {
		panic(err)
	}
	return msgs
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return server.UnpackInterfaces(unpacker, m.Msgs)
}

var _ sdk.Msg = &MsgVote{}

// Route Implements Msg.
func (m MsgVote) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVote) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVote.
func (m MsgVote) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}
	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}
	return nil
}

var _ sdk.Msg = &MsgVoteBasic{}

// Route Implements Msg.
func (m MsgVoteBasic) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteBasic) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteBasic) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVoteBasic) GetSigners() []sdk.AccAddress {
	panic("this message does not include signers")
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteBasic) ValidateBasic() error {
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}
	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "vote expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "vote expiry")
	}
	return nil
}

var _ sdk.Msg = &MsgVoteBasicResponse{}
var _ types.UnpackInterfacesMessage = MsgVoteBasicResponse{}

// Route Implements Msg.
func (m MsgVoteBasicResponse) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteBasicResponse) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteBasicResponse) GetSignBytes() []byte {
	res := MsgVoteBasic{
		ProposalId: m.ProposalId,
		Choice:     m.Choice,
		Expiry:     m.Expiry,
	}
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&res))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVoteBasicResponse) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteBasicResponse) ValidateBasic() error {
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}

	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}

	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter account")
	}

	if m.PubKey == nil {
		return sdkerrors.Wrap(ErrEmpty, "public key")
	}

	if len(m.Sig) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "voter signature")
	}

	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "expiry")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgVoteBasicResponse) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(m.PubKey, &pk)
}

func (m MsgVoteBasicResponse) VerifySignature() error {
	msgBytes := m.GetSignBytes()
	voterAddress := m.GetSigners()[0]

	pubKey, ok := m.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return sdkerrors.Wrap(ErrInvalid, "public key")
	}

	if !bytes.Equal(pubKey.Address(), voterAddress) {
		return sdkerrors.Wrapf(ErrInvalid, "public key does not match the voter's address %s", m.Voter)
	}

	if !pubKey.VerifySignature(msgBytes, m.Sig) {
		return sdkerrors.Wrap(ErrInvalid, "sigature verification failed")
	}

	return nil
}

var _ sdk.Msg = &MsgVoteAgg{}

// Route Implements Msg.
func (m MsgVoteAgg) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteAgg) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteAgg) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
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
		if _, ok := Choice_name[int32(c)]; !ok {
			return sdkerrors.Wrap(ErrInvalid, "choice")
		}
	}

	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}

	if len(m.AggSig) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "voter signature")
	}

	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "expiry")
	}
	return nil
}

var _ sdk.Msg = &MsgExec{}

// Route Implements Msg.
func (m MsgExec) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgExec) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgExec) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgExec.
func (m MsgExec) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgExec) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return sdkerrors.Wrap(err, "signer")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	return nil
}

var _ sdk.Msg = &MsgCreatePoll{}

// Route Implements Msg.
func (m MsgCreatePoll) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreatePoll) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreatePoll) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateProposal.
func (m MsgCreatePoll) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreatePoll) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		return sdkerrors.Wrap(err, "creator account")
	}

	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")
	}

	if len(m.Title) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "poll title")
	}

	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}
	if err := assertTitleLength(m.Title, "poll title"); err != nil {
		return err
	}

	if err := m.Options.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "options")
	}

	if m.VoteLimit <= 0 {
		return sdkerrors.Wrap(ErrInvalid, "vote limit must be positive")
	}

	if int(m.VoteLimit) > len(m.Options.Titles) {
		return sdkerrors.Wrap(ErrInvalid, "vote limit exceeds the number of options")
	}

	t, err := prototypes.TimestampFromProto(&m.Timeout)
	if err != nil {
		return sdkerrors.Wrap(err, "timeout")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "timeout")
	}

	return nil
}

var _ sdk.Msg = &MsgVotePoll{}

// Route Implements Msg.
func (m MsgVotePoll) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVotePoll) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVotePoll) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateProposal.
func (m MsgVotePoll) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVotePoll) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "creator account")
	}

	if m.PollId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "poll")
	}

	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}

	if err := m.Options.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "options")
	}

	return nil
}

var _ sdk.Msg = &MsgVotePollBasic{}

// Route Implements Msg.
func (m MsgVotePollBasic) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVotePollBasic) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVotePollBasic) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateProposal.
func (m MsgVotePollBasic) GetSigners() []sdk.AccAddress {
	panic("not implemented for this message")
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVotePollBasic) ValidateBasic() error {
	if m.PollId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "poll")
	}

	if len(m.Option) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "vote option")
	}

	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "vote expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "vote expiry")
	}
	return nil
}

var _ sdk.Msg = &MsgVotePollBasicResponse{}
var _ types.UnpackInterfacesMessage = MsgVotePollBasicResponse{}

// Route Implements Msg.
func (m MsgVotePollBasicResponse) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVotePollBasicResponse) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVotePollBasicResponse) GetSignBytes() []byte {
	panic("not implemented")
}

func (m MsgVotePollBasicResponse) GetSignBytesMany() [][]byte {
	signBytesMany := make([][]byte, 0, len(m.Options.Titles))
	for _, x := range m.Options.Titles {
		y := MsgVotePollBasic{
			PollId: m.PollId,
			Option: x,
			Expiry: m.Expiry,
		}
		signBytesMany = append(signBytesMany, y.GetSignBytes())
	}
	return signBytesMany
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVotePollBasicResponse) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVotePollBasicResponse) ValidateBasic() error {
	if m.PollId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if err := m.Options.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "options")
	}
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter account")
	}
	if m.PubKey == nil {
		return sdkerrors.Wrap(ErrEmpty, "public key")
	}
	if len(m.Sig) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "voter signature")
	}
	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "expiry")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgVotePollBasicResponse) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(m.PubKey, &pk)
}

func (m MsgVotePollBasicResponse) VerifySignature() error {
	msgsBytes := m.GetSignBytesMany()
	voterAddress := m.GetSigners()[0]

	pubKey, ok := m.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return sdkerrors.Wrap(ErrInvalid, "public key")
	}

	if !bytes.Equal(pubKey.Address(), voterAddress) {
		return sdkerrors.Wrapf(ErrInvalid, "public key does not match the voter's address %s", m.Voter)
	}

	pkBls, ok := pubKey.(*bls12381.PubKey)
	if !ok {
		return fmt.Errorf("only support bls public key")
	}

	// todo: repeated public keys can be coalesced in pairings
	pkss := make([][]*bls12381.PubKey, len(m.Options.Titles))
	for i := range m.Options.Titles {
		pkss[i] = []*bls12381.PubKey{pkBls}
	}

	if err := bls12381.VerifyAggregateSignature(msgsBytes, false, m.Sig, pkss); err != nil {
		return err
	}

	return nil
}

var _ sdk.Msg = &MsgVotePollAgg{}

// Route Implements Msg.
func (m MsgVotePollAgg) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVotePollAgg) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVotePollAgg) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVotePollAgg) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVotePollAgg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(err, "sender")
	}
	if m.PollId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "poll")
	}
	if len(m.Votes) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "votes")
	}
	for _, v := range m.Votes {
		if len(v.Titles) != 0 {
			if err := v.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(ErrInvalid, "options")
			}
		}
	}
	if err := assertMetadataLength(m.Metadata, "metadata"); err != nil {
		return err
	}
	if len(m.AggSig) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "signature")
	}
	t, err := prototypes.TimestampFromProto(&m.Expiry)
	if err != nil {
		return sdkerrors.Wrap(err, "expiry")
	}
	if t.IsZero() {
		return sdkerrors.Wrap(ErrEmpty, "expiry")
	}
	return nil
}
