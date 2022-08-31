package keeper_test

import (
	"fmt"
	"github.com/fetchai/fetchd/x/verifiable-credential/types"
)

func (s *KeeperTestSuite) TestVerifiableCredentialsKeeperSetAndGet() {
	testCases := []struct {
		msg string
		vc  types.VerifiableCredential
		// TODO: add mallate func and clean up test
		expPass bool
	}{
		//{
		//	"data stored successfully",
		//	types.NewUserVerifiableCredential(
		//		"did:cash:1111",
		//		"",
		//		time.Now(),
		//		types.NewUserCredentialSubject("", "root", true),
		//	),
		//	true,
		//},
	}
	for _, tc := range testCases {
		s.app.VcKeeper.SetVerifiableCredential(
			s.sdkCtx,
			[]byte(tc.vc.Id),
			tc.vc,
		)
		s.app.VcKeeper.SetVerifiableCredential(
			s.sdkCtx,
			[]byte(tc.vc.Id+"1"),
			tc.vc,
		)
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.expPass {
				_, found := s.app.VcKeeper.GetVerifiableCredential(
					s.sdkCtx,
					[]byte(tc.vc.Id),
				)
				s.Require().True(found)

				array := s.app.VcKeeper.GetAllVerifiableCredentials(
					s.sdkCtx,
				)

				s.Require().Equal(2, len(array))
			} else {
				// TODO write failure cases
				s.Require().False(tc.expPass)
			}
		})
	}
}
