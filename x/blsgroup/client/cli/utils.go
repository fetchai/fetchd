package cli

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupcli "github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/fetchai/fetchd/x/blsgroup"
)

func parseGroupMembers(clientCtx client.Context, membersFile string) ([]*group.GroupMember, error) {
	res := group.QueryGroupMembersResponse{}

	if membersFile == "" {
		return res.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}
	err = clientCtx.Codec.UnmarshalJSON(contents, &res)
	if err != nil {
		return nil, err
	}

	if res.Pagination.NextKey != nil {
		return nil, fmt.Errorf("require all the group members")
	}

	return res.Members, nil
}

func sortGroupMembersFunc(groupMembers []*group.GroupMember) func(i, j int) bool {
	return func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(groupMembers[i].Member.Address)
		if err != nil {
			panic(err)
		}
		addrj, err := sdk.AccAddressFromBech32(groupMembers[j].Member.Address)
		if err != nil {
			panic(err)
		}
		return bytes.Compare(addri, addrj) < 0
	}
}

func parseBlsVote(clientCtx client.Context, voteFile string) (blsgroup.MsgVoteResponse, error) {
	vote := blsgroup.MsgVoteResponse{}

	if voteFile == "" {
		return vote, nil
	}

	contents, err := ioutil.ReadFile(voteFile)
	if err != nil {
		return vote, err
	}

	err = clientCtx.Codec.UnmarshalJSON(contents, &vote)
	if err != nil {
		return vote, err
	}

	return vote, nil
}

func execFromString(execStr string) group.Exec {
	exec := group.Exec_EXEC_UNSPECIFIED
	if execStr == groupcli.ExecTry {
		exec = group.Exec_EXEC_TRY
	}
	return exec
}
