package keeper_test

import (
	"context"
	"fmt"
	"github.com/fetchai/fetchd/x/did/types"
)

func (s *KeeperTestSuite) TestGRPCQueryDidDocuments() {
	queryClient := s.queryClient
	var req *types.QueryDidDocumentsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"Pass: will return an empty array",
			func() {
				req = &types.QueryDidDocumentsRequest{}
			},
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			didsResp, err := queryClient.DidDocuments(context.Background(), req)
			if tc.expPass {
				s.NoError(err)
				s.NotNil(didsResp)

			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryDidDocument() {
	queryClient := s.queryClient
	var req *types.QueryDidDocumentRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"Fail: will fail because no id is provided",
			func() {
				req = &types.QueryDidDocumentRequest{}
			},
			false,
		},
		{
			"Fail: will fail because no did is found",
			func() {
				req = &types.QueryDidDocumentRequest{
					Id: "did:cosmos:cash:1234",
				}
			},
			false,
		},
		{
			"Pass: will pass because a did is found",
			func() {

				dd, _ := types.NewDidDocument("did:cosmos:cash:1234")

				s.app.DidKeeper.SetDidDocument(
					s.sdkCtx,
					[]byte(dd.Id),
					dd,
				)
				req = &types.QueryDidDocumentRequest{
					Id: "did:cosmos:cash:1234",
				}
			},
			true,
		},
		{
			"Pass: will pass because a address did is autoresolved",
			func() {
				req = &types.QueryDidDocumentRequest{
					Id: "did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				}
			},
			true,
		},
		{
			"Fail: will fail because the only cosmos based address are supported",
			func() {
				req = &types.QueryDidDocumentRequest{
					Id: "did:cosmos:key:0xB88F61E6FbdA83fbfffAbE364112137480398018",
				}
			},
			false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			didsResp, err := queryClient.DidDocument(context.Background(), req)
			if tc.expPass {
				s.NoError(err)
				s.NotNil(didsResp)

			} else {
				s.Require().Error(err)
			}
		})
	}
}
