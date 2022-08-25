package keeper_test

import (
	"fmt"
	"github.com/fetchai/fetchd/x/did/types"
)

func (s *KeeperTestSuite) TestDidDocumentKeeperSetAndGet() {
	testCases := []struct {
		msg     string
		didFn   func() types.DidDocument
		expPass bool
	}{
		{
			"data stored successfully",
			func() types.DidDocument {
				dd, _ := types.NewDidDocument("did:cash:subject")
				return dd
			},
			true,
		},
	}
	for _, tc := range testCases {
		dd := tc.didFn()

		s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(dd.Id), dd)
		s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(dd.Id+"1"), dd)
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.expPass {
				_, found := s.app.DidKeeper.GetDidDocument(
					s.sdkCtx,
					[]byte(dd.Id),
				)
				s.Require().True(found)

				allEntities := s.app.DidKeeper.GetAllDidDocuments(
					s.sdkCtx,
				)
				s.Require().Equal(2, len(allEntities))
			} else {
				// TODO write failure cases
				s.Require().False(tc.expPass)
			}
		})
	}
}
