package client

import (
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/fetchai/fetchd/x/group"
)

func parseMembers(clientCtx client.Context, membersFile string) ([]group.Member, error) {
	members := group.Members{}

	if membersFile == "" {
		return members.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}

	err = clientCtx.JSONMarshaler.UnmarshalJSON(contents, &members)
	if err != nil {
		return nil, err
	}

	return members.Members, nil
}

func execFromString(execStr string) group.Exec {
	var exec group.Exec
	switch execStr {
	case ExecTry:
		exec = group.Exec_EXEC_TRY
	default:
		exec = group.Exec_EXEC_UNSPECIFIED
	}
	return exec
}

func parseGroupMembers(clientCtx client.Context, membersFile string) ([]*group.GroupMember, error) {
	res := group.QueryGroupMembersResponse{}

	if membersFile == "" {
		return res.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}
	err = clientCtx.JSONMarshaler.UnmarshalJSON(contents, &res)
	if err != nil {
		return nil, err
	}

	if res.Pagination.NextKey != nil {
		return nil, fmt.Errorf("require all the group members")
	}

	return res.Members, nil
}

func parseVoteBasic(clientCtx client.Context, voteFile string) (group.MsgVoteBasicResponse, error) {
	vote := group.MsgVoteBasicResponse{}

	if voteFile == "" {
		return vote, nil
	}

	contents, err := ioutil.ReadFile(voteFile)
	if err != nil {
		return vote, err
	}

	err = clientCtx.JSONMarshaler.UnmarshalJSON(contents, &vote)
	if err != nil {
		return vote, err
	}

	return vote, nil
}

func parseVotePollBasic(clientCtx client.Context, voteFile string) (group.MsgVotePollBasicResponse, error) {
	vote := group.MsgVotePollBasicResponse{}

	if voteFile == "" {
		return vote, nil
	}

	contents, err := ioutil.ReadFile(voteFile)
	if err != nil {
		return vote, err
	}

	err = clientCtx.JSONMarshaler.UnmarshalJSON(contents, &vote)
	if err != nil {
		return vote, err
	}

	return vote, nil
}
