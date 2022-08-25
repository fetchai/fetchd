package cli_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fetchai/fetchd/x/did/client/cli"
	"github.com/fetchai/fetchd/x/did/types"

	"github.com/fetchai/fetchd/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	dbm "github.com/tendermint/tm-db"
)

// NewAppConstructor returns a new simapp AppConstructor
func NewAppConstructor(encodingCfg app.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return app.New(
			val.Ctx.Logger,
			dbm.NewMemDB(), nil, true, make(map[int64]bool),
			val.Ctx.Config.RootDir,
			0,
			encodingCfg,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

type IntegrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
}

// SetupSuite executes bootstrapping logic before all the tests, i.e. once before
// the entire suite, start executing.
func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	cfg := network.DefaultConfig()
	types.RegisterInterfaces(cfg.InterfaceRegistry)
	cfg.AppConstructor = NewAppConstructor(app.MakeEncodingConfig())
	cfg.NumValidators = 2
	s.cfg = cfg
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownSuite performs cleanup logic after all the tests, i.e. once after the
// entire suite, has finished executing.
func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func name() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func addnewdiddoc(s *IntegrationTestSuite, identifier string, val *network.Validator) {

	clientCtx := val.ClientCtx
	args := []string{
		identifier,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf(
			"--%s=%s",
			flags.FlagFees,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
		),
	}

	cmd := cli.NewCreateDidDocumentCmd()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
	// wait for blocks
	for i := 0; i < 2; i++ {
		netError := s.network.WaitForNextBlock()
		s.Require().NoError(netError)
	}
	response := &sdk.TxResponse{}
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response), out.String())
}

func (s *IntegrationTestSuite) TestGetCmdQueryDidDocuments() {
	identifier := "123456789abcdefghijka"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name     string
		args     []string
		respType proto.Message
		malleate func()
	}{
		{
			name() + "_1",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			&types.QueryDidDocumentsResponse{},
			func() {},
		},
		{
			name() + "_2",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			&types.QueryDidDocumentsResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	var first bool = true
	var size = 0
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.GetCmdQueryIdentifers()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			queryresponse := tc.respType.(*types.QueryDidDocumentsResponse)
			diddocs := queryresponse.GetDidDocuments()
			if first {
				first = false
				size = len(diddocs)
			} else {
				s.Require().Greater(len(diddocs), 0)
				s.Require().Equal(size+1, len(diddocs))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryDidDocument() {
	identifier := "123456789abcdefghijkb"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		expectErr codes.Code
		respType  proto.Message
		malleate  func()
	}{
		{
			name() + "_1",
			codes.NotFound,
			&types.QueryDidDocumentResponse{},
			func() {},
		},
		{
			name() + "_2",
			codes.OK,
			&types.QueryDidDocumentResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.GetCmdQueryIdentifer()
			identifiertoquery := "did:cosmos:net:" + clientCtx.ChainID + ":" + identifier
			args := []string{
				identifiertoquery,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr != codes.OK {
				s.Require().Error(err)
				s.Equal(tc.expectErr, status.Code(err))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				queryresponse := tc.respType.(*types.QueryDidDocumentResponse)
				diddoc := queryresponse.GetDidDocument()
				s.Require().Equal(identifiertoquery, diddoc.Id)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewCreateDidDocumentCmd() {

	identifier := "123456789abcdefghijkc"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name     string
		args     []string
		respType proto.Message
	}{
		{
			name(),
			[]string{
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				)},
			&sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {

		s.Run(tc.name, func() {
			var size = 0
			for i := 0; i < 3; i++ {
				cmd := cli.NewCreateDidDocumentCmd()
				tc.args[0] = identifier + fmt.Sprint(i)
				out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
				s.Require().NoError(err)
				// wait for blocks
				for i := 0; i < 2; i++ {
					netError := s.network.WaitForNextBlock()
					s.Require().NoError(netError)
				}
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				//pull out the just created document
				cmd = cli.GetCmdQueryIdentifer()
				identifiertoquery := "did:cosmos:net:" + clientCtx.ChainID + ":" + tc.args[0]
				args_temp := []string{
					identifiertoquery,
					fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				}
				out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
				s.Require().NoError(err)
				response1 := &types.QueryDidDocumentResponse{}
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response1))
				s.Require().Equal(response1.GetDidDocument().Id, identifiertoquery)

				//pull out the set of created documentsq
				cmd = cli.GetCmdQueryIdentifers()
				args_temp = []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
				out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
				s.Require().NoError(err)
				response2 := &types.QueryDidDocumentsResponse{}
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response2))
				diddocs := response2.GetDidDocuments()
				if i == 0 {
					size = len(diddocs)
				} else {
					s.Require().Equal(size+i, len(diddocs))
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewUpdateDidDocumentCmd() {
	identifier1 := "123456789abcdefghijkd"
	identifier2 := "cosmos1kslgpxklq75aj96cz3qwsczr95vdtrd3p0fslp"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name     string
		args     []string
		respType proto.Message
		malleate func()
	}{
		{
			name(),
			[]string{
				identifier1,
				identifier2,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				),
			},
			&sdk.TxResponse{},
			func() { addnewdiddoc(s, identifier1, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.NewAddControllerCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier1,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			controller := response.GetDidDocument().Controller
			s.Require().Equal(len(controller), 1)
			s.Require().Equal(controller[0], "did:cosmos:key:"+identifier2)
		})
	}
}

func (s *IntegrationTestSuite) TestNewAddVerificationCmd() {
	identifier := "123456789abcdefghijke"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr codes.Code
		respType  proto.Message
		malleate  func()
	}{
		{
			name(),
			[]string{
				identifier,
				`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AhJhB4NzRr2+pRpW4jDfajpML2h9yuBONsSqz6aXKZ6s"}`,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				),
			},
			codes.OK,
			&sdk.TxResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.NewAddVerificationCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(err)
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			authentications := response.GetDidDocument().Authentication
			verificationmethods := response.GetDidDocument().VerificationMethod
			s.Require().Equal(2, len(authentications))
			s.Require().Equal(2, len(verificationmethods))
			for i := 0; i < 2; i++ {
				s.Require().Equal(authentications[i], verificationmethods[i].Id)
			}

			verificationmethod := verificationmethods[1]
			s.Require().Equal("F02126107837346bdbea51a56e230df6a3a4c2f687dcae04e36c4aacfa697299eac", verificationmethod.GetPublicKeyMultibase())
		})
	}
}

func (s *IntegrationTestSuite) TestNewSetVerificationRelationshipsCmd() {
	identifier := "123456789abcdefghijkf"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr codes.Code
		respType  proto.Message
		malleate  func()
	}{
		{
			name(),
			[]string{
				identifier,
				"",
				fmt.Sprintf("--relationship=%s", types.CapabilityDelegation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				),
			},
			codes.OK,
			&sdk.TxResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			verificationmethods := response.GetDidDocument().VerificationMethod
			s.Require().Greater(len(verificationmethods), 0)
			temp := strings.Split(verificationmethods[0].Id, "#")
			tc.args[1] = temp[len(temp)-1]
			cmd = cli.NewSetVerificationRelationshipCmd()

			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(err)
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response = &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			capabilitydelegation := response.GetDidDocument().CapabilityDelegation
			s.Require().Equal(1, len(capabilitydelegation))
			s.Require().Equal(verificationmethods[0].Id, capabilitydelegation[0])
		})
	}
}

func (s *IntegrationTestSuite) TestNewRevokeVerificationCmd() {
	identifier := "123456789abcdefghijkg"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr codes.Code
		respType  proto.Message
		malleate  func()
	}{
		{
			name(),
			[]string{
				identifier,
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				),
			},
			codes.OK,
			&sdk.TxResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))

			verificationmethods := response.GetDidDocument().VerificationMethod
			s.Require().Greater(len(verificationmethods), 0)
			temp := strings.Split(verificationmethods[0].Id, "#")
			tc.args[1] = temp[len(temp)-1]
			cmd = cli.NewRevokeVerificationCmd()

			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response = &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			s.Require().Equal(0, len(response.GetDidDocument().VerificationMethod))
			s.Require().Equal(0, len(response.GetDidDocument().Authentication))
		})
	}
}

func (s *IntegrationTestSuite) TestNewAddServiceCmd() {
	identifier := "123456789abcdefghijkh"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr codes.Code
		respType  proto.Message
		malleate  func()
	}{
		{
			name(),
			[]string{
				identifier,
				"service:seuro",
				"DIDComm",
				"service:euro/SIGNATURE",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf(
					"--%s=%s",
					flags.FlagFees,
					sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				),
			},
			codes.OK,
			&sdk.TxResponse{},
			func() { addnewdiddoc(s, identifier, val) },
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()
			cmd := cli.NewAddServiceCmd()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			s.Require().Equal(1, len(response.GetDidDocument().Service))
			s.Require().Equal(tc.args[1], response.GetDidDocument().Service[0].Id)
			s.Require().Equal(tc.args[2], response.GetDidDocument().Service[0].Type)
			s.Require().Equal(tc.args[3], response.GetDidDocument().Service[0].ServiceEndpoint)
		})
	}
}

func (s *IntegrationTestSuite) TestNewDeleteServiceCmd() {
	identifier := "123456789abcdefghijki"
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	args := []string{
		identifier,
		"service:seuro",
		"DIDComm",
		"service:euro/SIGNATURE",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf(
			"--%s=%s",
			flags.FlagFees,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
		),
	}

	testCases := []struct {
		name     string
		respType proto.Message
		malleate func()
	}{
		{
			name(),
			&sdk.TxResponse{},
			func() {
				addnewdiddoc(s, identifier, val)
				cmd := cli.NewAddServiceCmd()
				out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
				s.Require().NoError(err)
				// wait for blocks
				for i := 0; i < 2; i++ {
					netError := s.network.WaitForNextBlock()
					s.Require().NoError(netError)
				}
				response := &sdk.TxResponse{}
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response), out.String())
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {

			tc.malleate()
			cmd := cli.NewDeleteServiceCmd()

			args = append(args[:2], args[4:]...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			s.Require().NoError(err)
			// wait for blocks
			for i := 0; i < 2; i++ {
				netError := s.network.WaitForNextBlock()
				s.Require().NoError(netError)
			}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

			//check for update
			cmd = cli.GetCmdQueryIdentifer()
			args_temp := []string{
				"did:cosmos:net:" + clientCtx.ChainID + ":" + identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args_temp)
			s.Require().NoError(err)
			response := &types.QueryDidDocumentResponse{}
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), response))
			s.Require().Equal(0, len(response.GetDidDocument().Service))
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
