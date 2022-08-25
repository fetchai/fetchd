package types_test

import (
	"fmt"

	"github.com/fetchai/fetchd/app"
	didtypes "github.com/fetchai/fetchd/x/did/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/crypto/types"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const contextDIDBase = "https://www.w3.org/ns/did/v1"

func TestNewChainDID(t *testing.T) {

	tests := []struct {
		did   string
		chain string
		want  didtypes.DID
	}{
		{
			"subject",
			"cash",
			didtypes.DID("did:cosmos:net:cash:subject"),
		},
		{
			"",
			"cash",
			didtypes.DID("did:cosmos:net:cash:"),
		},
		{
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			"cosmoshub",
			didtypes.DID("did:cosmos:net:cosmoshub:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75"),
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDID#", i), func(t *testing.T) {
			if got := didtypes.NewChainDID(tt.chain, tt.did); got != tt.want {
				t.Errorf("DID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeyDID(t *testing.T) {

	tests := []struct {
		name    string
		account string
		want    didtypes.DID
	}{
		{
			"PASS: valid account",
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, didtypes.NewKeyDID(tt.account), "NewKeyDID(%v)", tt.account)
		})
	}
}

func TestDID_NewVerificationMethodID(t *testing.T) {

	tests := []struct {
		name string
		did  didtypes.DID
		vmID string
		want string
	}{
		{
			"PASS: generated vmId for network DID",
			didtypes.DID("did:cosmos:net:foochain:whatever"),
			"123456",
			"did:cosmos:net:foochain:whatever#123456",
		},
		{
			"PASS: generated vmId for key DID",
			didtypes.DID("did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
			"123456",
			"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8#123456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.did.NewVerificationMethodID(tt.vmID), "NewVerificationMethodID(%v)", tt.vmID)
		})
	}
}

func TestIsValidDID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"did:cash:net:subject", true},
		{"did:cash:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75", true},
		{"did:cash:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75#key-1", false},
		{"did:subject", false},
		{"DID:cash:subject", false},
		{"d1d:cash:subject", false},
		{"d1d:#:subject", false},
		{"", false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestIsValidDID#", i), func(t *testing.T) {
			if got := didtypes.IsValidDID(tt.input); got != tt.want {
				t.Errorf("IsValidDID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidDIDURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"did:cash:subject", true},
		{"did:cash:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75", true},
		{"did:cash:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75#key-1", true},
		{"did:cash:cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75?key=1", true},
		{"did:cosmos:net:cosmoscash-testnet:575d062c-d110-42a9-9c04-cb1ff8c01f06#Z46DAL1MrJlVW_WmJ19WY8AeIpGeFOWl49Qwhvsnn2M", true},
		{"did:subject", false},
		{"DID:cash:subject", false},
		{"d1d:cash:subject", false},
		{"d1d:#:subject", false},
		{"", false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestIsValidDIDURL#", i), func(t *testing.T) {
			if got := didtypes.IsValidDIDURL(tt.input); got != tt.want {
				t.Errorf("IsValidDIDURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRFC3986Uri(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{
			"[][àséf",
			true,
		},
		{
			"# \u007e // / / ///// //// // / / ??? ?? 不",
			true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestIsValidRFC3986Uri#", i), func(t *testing.T) {
			if got := didtypes.IsValidRFC3986Uri(tt.input); got != tt.want {
				t.Errorf("IsValidRFC3986Uri() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidDIDDocument(t *testing.T) {
	tests := []struct {
		name  string
		didFn func() *didtypes.DidDocument
		want  bool
	}{
		{
			"PASS: document is valid",
			func() *didtypes.DidDocument {
				return &didtypes.DidDocument{
					Context: []string{contextDIDBase},
					Id:      "did:cosmos:net:cash:1",
				}
			},
			true, // all good
		},
		{
			"FAIL: empty context",
			func() *didtypes.DidDocument {
				return &didtypes.DidDocument{
					Context: []string{},
					Id:      "did:cosmos:net:cash:1",
				}
			},
			false, // missing context
		},
		{
			"PASS: minimal did document",
			func() *didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cosmos:cash:1")
				assert.NoError(t, err)
				return &dd
			},
			true, // all good
		},
		{
			"FAIL: empty did",
			func() *didtypes.DidDocument {
				return &didtypes.DidDocument{
					Context: []string{contextDIDBase},
					Id:      "",
				}
			},
			false, // empty id
		},
		{
			"FAIL: nil did document",
			func() *didtypes.DidDocument {
				return nil
			},
			false, // nil pointer
		},
		{
			"PASS: did with valid controller",
			func() *didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cosmos:key:cas:1", didtypes.WithControllers(
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				))
				assert.NoError(t, err)
				return &dd
			},
			true,
		},
		{
			"FAIL: invalid controller",
			func() *didtypes.DidDocument {
				return &didtypes.DidDocument{
					Context: []string{contextDIDBase},
					Id:      "did:cosmos:net:foochain:1",
					Controller: []string{
						"did:cosmos:net:foochain:whatever",
					},
				}
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprint("TestIsValidDIDDocument#", tt.name), func(t *testing.T) {
			if got := didtypes.IsValidDIDDocument(tt.didFn()); got != tt.want {
				t.Errorf("TestIsValidDIDDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidDIDMetadata(t *testing.T) {

	tests := []struct {
		didMetaFn func() *didtypes.DidMetadata
		want      bool
	}{
		{
			func() *didtypes.DidMetadata {
				now := time.Now()
				return &didtypes.DidMetadata{
					VersionId: "d95daac05a36f93d1494208d02d1522d758466c62ea6b64c50b78999d2021f51",
					Created:   &now,
				}
			},
			true, // all good
		},
		{
			func() *didtypes.DidMetadata {
				now := time.Now()
				return &didtypes.DidMetadata{
					VersionId: "",
					Created:   &now,
				}
			},
			false, // missing version
		},
		{
			func() *didtypes.DidMetadata {
				now := time.Now()
				return &didtypes.DidMetadata{
					VersionId: "d95daac05a36f93d1494208d02d1522d758466c62ea6b64c50b78999d2021f51",
					Updated:   &now,
				}
			},
			false, // null created
		},
		{
			func() *didtypes.DidMetadata {
				var now time.Time
				return &didtypes.DidMetadata{
					VersionId: "d95daac05a36f93d1494208d02d1522d758466c62ea6b64c50b78999d2021f51",
					Created:   &now,
				}
			},
			false, // zero created
		},
		{
			func() *didtypes.DidMetadata {
				return nil
			},
			false, // nil pointer
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestIsValidDIDMetadata#", i), func(t *testing.T) {
			if got := didtypes.IsValidDIDMetadata(tt.didMetaFn()); got != tt.want {
				t.Errorf("TestIsValidDIDMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVerification(t *testing.T) {
	tests := []struct {
		v       *didtypes.Verification
		wantErr bool
	}{
		{
			v: didtypes.NewVerification(
				didtypes.NewVerificationMethod(
					"did:cash:subject#key-1",
					"did:cash:subject",
					didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
				),
				nil,
				nil,
			),
			wantErr: true, // relationships are nil
		},
		{
			v:       nil,
			wantErr: true,
		},
		{
			v: didtypes.NewVerification(
				didtypes.NewVerificationMethod(
					"did:cash:subject#key-1",
					didtypes.DID("did:cash:subject"),
					didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
				),
				[]string{string(didtypes.AssertionMethod)},
				nil,
			),
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestValidateVerification#", i), func(t *testing.T) {
			if err := didtypes.ValidateVerification(tt.v); (err != nil) != tt.wantErr {
				t.Errorf("TestValidateVerification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateService(t *testing.T) {

	tests := []struct {
		s       *didtypes.Service
		wantErr bool
	}{
		{
			s:       didtypes.NewService("agent:abc", "DIDCommMessaging", "https://agent.abc/abc"),
			wantErr: false,
		},
		{
			s:       nil,
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestValidateService#", i), func(t *testing.T) {
			if err := didtypes.ValidateService(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("ValidateService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"    a    ", false},
		{"\t", true},
		{"\n", true},
		{"   ", true},
		{"  \t \n", true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestIsEmpty#", i), func(t *testing.T) {
			if got := didtypes.IsEmpty(tt.input); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDidDocument(t *testing.T) {
	type params struct {
		id      string
		options []didtypes.DidDocumentOption
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument
		wantErr bool
	}{
		{
			params: params{
				"did:cash:subject",
				[]didtypes.DidDocumentOption{
					didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								string(didtypes.Authentication),
								string(didtypes.KeyAgreement),
								string(didtypes.KeyAgreement), // test duplicated relationship
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					),
					didtypes.WithVerifications( // multiple verifications in separate entity
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-2",
								"did:cash:subject",
								didtypes.NewBlockchainAccountID("cash", "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"),
							),
							[]string{
								string(didtypes.Authentication),
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					),
					didtypes.WithServices(&didtypes.Service{
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					}),
					didtypes.WithControllers("did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"),
				},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{
					"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
					contextDIDBase,
				},
				Id:         "did:cash:subject",
				Controller: []string{"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"},
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyMultibase{"F03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-2",
						string(didtypes.DIDVMethodTypeCosmosAccountAddress),
						"did:cash:subject",
						&didtypes.VerificationMethod_BlockchainAccountID{"cosmos:cash:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"},
					},
				},
				Service: []*didtypes.Service{
					{
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					},
				},
				Authentication: []string{"did:cash:subject#key-1", "did:cash:subject#key-2"},
				KeyAgreement:   []string{"did:cash:subject#key-1"},
			},
			wantErr: false,
		},
		{
			params: params{
				"did:cash:subject",
				[]didtypes.DidDocumentOption{
					didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1", // duplicate key
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{},
						),
					),
					didtypes.WithServices(&didtypes.Service{
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					}),
				},
			},
			wantDid: didtypes.DidDocument{},
			wantErr: true, // the error is caused by duplicated verification method id
		},
		{
			params: params{
				"did:cash:subject",
				[]didtypes.DidDocumentOption{
					didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte("02503c8ace59c085b15c5f9c2474e9235bcb9694f07516bdc06f7caec788c3dd2c"), didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					),
					didtypes.WithServices(
						&didtypes.Service{
							"agent:xyz",
							"DIDCommMessaging",
							"https://agent.xyz/1234",
						},
						&didtypes.Service{
							"agent:xyz",
							"DIDCommMessaging",
							"https://agent.xyz/1234",
						},
					),
				},
			},
			wantDid: didtypes.DidDocument{},
			wantErr: true, //duplicated service id
		},
		{
			wantErr: true, // invalid did
			params: params{
				id:      "something not right",
				options: []didtypes.DidDocumentOption{},
			},
			wantDid: didtypes.DidDocument{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestNewDidDocument#", i), func(t *testing.T) {
			gotDid, err := didtypes.NewDidDocument(tt.params.id, tt.params.options...)

			if tt.wantErr {
				require.NotNil(t, err, "test: TestNewDidDocument#%v", i)
				return
			}

			require.Nil(t, err, "test: TestNewDidDocument#%v", i)
			assert.Equal(t, tt.wantDid, gotDid)
		})
	}
}

func TestDidDocument_AddControllers(t *testing.T) {

	tests := []struct {
		name                string
		malleate            func() didtypes.DidDocument
		controllers         []string
		expectedControllers []string
		wantErr             error
	}{
		{
			"PASS: controllers added",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", // duplicate controllers
					),
				)
				return dd
			},
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			nil,
		},
		{
			"FAIL: invalid controller did",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
					),
				)
				return dd
			},
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				"not a did:cosmos:key:cosmos100000000000000000000000000000000000004",
			},
			[]string{},
			sdkerrors.Wrapf(didtypes.ErrInvalidDIDFormat, "did document controller validation error 'not a did:cosmos:key:cosmos100000000000000000000000000000000000004'"),
		},
		{
			"FAIL: controller did is not type key",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
					),
				)
				return dd
			},
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				"did:cosmos:net:foochain:1234",
			},
			[]string{},
			sdkerrors.Wrapf(didtypes.ErrInvalidInput, "did document controller 'did:cosmos:net:foochain:1234' must be of type key"),
		},
		{
			"PASS: controllers empty",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					),
				)
				return dd
			},
			nil,
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_AddControllers#", i), func(t *testing.T) {
			didDoc := tt.malleate()
			err := didDoc.AddControllers(tt.controllers...)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDidDocument_DeleteControllers(t *testing.T) {

	tests := []struct {
		name                string
		malleate            func() didtypes.DidDocument
		controllers         []string
		expectedControllers []string
		wantErr             error
	}{
		{
			"PASS: controllers deleted",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					),
				)
				return dd
			},
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			[]string{
				"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			nil,
		},
		{
			"FAIL: invalid controller did",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
					),
				)
				return dd
			},
			[]string{
				"not a did:cosmos:key:cosmos100000000000000000000000000000000000004",
			},
			[]string{},
			sdkerrors.Wrapf(didtypes.ErrInvalidDIDFormat, "did document controller validation error 'not a did:cosmos:key:cosmos100000000000000000000000000000000000004'"),
		},
		{
			"PASS: controllers empty",
			func() didtypes.DidDocument {
				dd, _ := didtypes.NewDidDocument("did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					),
				)
				return dd
			},
			nil,
			[]string{
				"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
				"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_AddControllers#", i), func(t *testing.T) {
			didDoc := tt.malleate()
			err := didDoc.DeleteControllers(tt.controllers...)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDidDocument_AddVerifications(t *testing.T) {
	type params struct {
		malleate      func() didtypes.DidDocument // build a did document
		verifications []*didtypes.Verification    // input list of verifications
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument // expected result
		wantErr bool
	}{
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject")
					return d
				},
				[]*didtypes.Verification{
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-2",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.CapabilityInvocation,
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					),
				},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{
					"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
					contextDIDBase,
				},
				Id:         "did:cash:subject",
				Controller: nil,
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyMultibase{"F03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-2",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyMultibase{"F03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},
				Service:              nil,
				Authentication:       []string{"did:cash:subject#key-1", "did:cash:subject#key-2"},
				KeyAgreement:         []string{"did:cash:subject#key-1"},
				CapabilityInvocation: []string{"did:cash:subject#key-2"},
			},
		},
		{
			wantErr: true, // duplicated existing method id
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
								didtypes.KeyAgreement, // test duplicated relationship
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					))
					return d
				},
				[]*didtypes.Verification{
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							string(didtypes.CapabilityDelegation),
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					),
				},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: true, // duplicated new method id
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
								didtypes.KeyAgreement, // test duplicated relationship
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					))
					return d
				},
				[]*didtypes.Verification{
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-2",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.KeyAgreement,
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					),
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-2",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					),
				},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: true, // fail validation
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
								didtypes.KeyAgreement, // test duplicated relationship
							},
							[]string{
								"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
							},
						),
					))
					return d
				},
				[]*didtypes.Verification{
					{
						[]string{
							string(didtypes.Authentication),
							string(didtypes.KeyAgreement),
							string(didtypes.KeyAgreement), // test duplicated relationship
						},
						&didtypes.VerificationMethod{
							"invalid method url",
							didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
							"did:cash:subject",
							&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					},
				},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: true, // verification relationship does not exists
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject")
					return d
				},
				[]*didtypes.Verification{
					{
						[]string{
							didtypes.Authentication,
							"UNSUPPORTED RELATIONSHIP",
							didtypes.KeyAgreement,
						},
						&didtypes.VerificationMethod{
							"did:cash:subject#key1",
							didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
							"did:cash:subject",
							&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
						},
						[]string{
							"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
						},
					},
				},
			},
			wantDid: didtypes.DidDocument{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_AddVerifications#", i), func(t *testing.T) {
			gotDid := tt.params.malleate()

			err := gotDid.AddVerifications(tt.params.verifications...)

			if tt.wantErr {
				require.NotNil(t, err, "test: TestDidDocument_AddVerifications#%v", i)
				return
			}

			require.Nil(t, err, "test: TestDidDocument_AddVerifications#%v", i)
			assert.Equal(t, tt.wantDid, gotDid)
		})
	}
}

func TestDidDocument_RevokeVerification(t *testing.T) {
	type params struct {
		malleate func() didtypes.DidDocument // build a did document
		methodID string                      // input list of verifications
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument // expected result
		wantErr bool
	}{
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithVerifications(
							didtypes.NewVerification(
								didtypes.NewVerificationMethod(
									"did:cash:subject#key-1",
									"did:cash:subject",
									didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
								),
								[]string{
									didtypes.Authentication,
									didtypes.KeyAgreement,
								},
								nil,
							),
							didtypes.NewVerification(
								didtypes.NewVerificationMethod(
									"did:cash:subject#key-2",
									"did:cash:subject",
									didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
								),
								[]string{
									didtypes.Authentication,
									didtypes.CapabilityInvocation,
								},
								[]string{
									"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
								},
							),
						),
					)
					return d
				},
				"did:cash:subject#key-2",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{
					"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
					contextDIDBase,
				},
				Id:         "did:cash:subject",
				Controller: nil,
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyMultibase{"F03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},
				Service:        nil,
				Authentication: []string{"did:cash:subject#key-1"},
				KeyAgreement:   []string{"did:cash:subject#key-1"},
			},
		},
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithVerifications(
							didtypes.NewVerification(
								didtypes.VerificationMethod{
									"did:cash:subject#key-1",
									didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
									"did:cash:subject",
									&didtypes.VerificationMethod_PublicKeyMultibase{"F03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
								},
								[]string{
									didtypes.Authentication,
									didtypes.KeyAgreement,
								},
								nil,
							),
						),
					)
					return d
				},
				"did:cash:subject#key-1",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{
					contextDIDBase,
				},
				Id: "did:cash:subject",
			},
		},
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithVerifications(
							didtypes.NewVerification(
								didtypes.VerificationMethod{
									"did:cash:subject#key-1",
									didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
									"did:cash:subject",
									&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
								},
								[]string{
									didtypes.Authentication,
									didtypes.KeyAgreement,
								},
								nil,
							),
							didtypes.NewVerification(
								didtypes.VerificationMethod{
									"did:cash:subject#key-2",
									didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
									"did:cash:subject",
									&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
								},
								[]string{
									didtypes.Authentication,
									didtypes.CapabilityInvocation,
								},
								nil,
							),
							didtypes.NewVerification(
								didtypes.VerificationMethod{
									"did:cash:subject#key-3",
									didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
									"did:cash:subject",
									&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
								},
								[]string{
									didtypes.Authentication,
									didtypes.KeyAgreement,
									didtypes.AssertionMethod,
								},
								nil,
							),
						),
					)
					return d
				},
				"did:cash:subject#key-2",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{
					contextDIDBase,
				},
				Id:         "did:cash:subject",
				Controller: nil,
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-3",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},
				Service:         nil,
				Authentication:  []string{"did:cash:subject#key-1", "did:cash:subject#key-3"},
				KeyAgreement:    []string{"did:cash:subject#key-1", "did:cash:subject#key-3"},
				AssertionMethod: []string{"did:cash:subject#key-3"},
			},
		},
		{
			wantErr: true, // verification method not found
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithVerifications(
							didtypes.NewVerification(
								didtypes.NewVerificationMethod(
									"did:cash:subject#key-1",
									"did:cash:subject",
									didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
								),
								[]string{
									didtypes.Authentication,
									didtypes.KeyAgreement,
								},
								nil,
							),
							didtypes.NewVerification(
								didtypes.NewVerificationMethod(
									"did:cash:subject#key-2",
									"did:cash:subject",
									didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
								),
								[]string{
									didtypes.Authentication,
									didtypes.CapabilityInvocation,
								},
								[]string{
									"https://gpg.jsld.org/contexts/lds-gpg2020-v0.0.jsonld",
								},
							),
						),
					)
					return d
				},
				"did:cash:subject#key-3",
			},
			wantDid: didtypes.DidDocument{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_RevokeVerification#", i), func(t *testing.T) {
			gotDid := tt.params.malleate()

			err := gotDid.RevokeVerification(tt.params.methodID)

			if tt.wantErr {
				require.NotNil(t, err, "test: TestDidDocument_RevokeVerification#%v", i)
				return
			}

			require.Nil(t, err, "test: TestDidDocument_RevokeVerification#%v", i)

			assert.Equal(t, tt.wantDid, gotDid)
		})
	}
}

func TestDidDocument_SetVerificationRelationships(t *testing.T) {
	type params struct {
		malleate      func() didtypes.DidDocument
		methodID      string
		relationships []string
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument // expected result
		wantErr bool
	}{
		{
			wantErr: true, // empty relationships
			params: params{
				malleate: func() didtypes.DidDocument {
					dd, _ := didtypes.NewDidDocument("did:cash:subject")
					return dd
				},
				methodID:      "did:cash:subject#key-1",
				relationships: []string{},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
			},
		},
		{
			wantErr: true, //invalid method id
			params: params{
				malleate: func() didtypes.DidDocument {
					dd, _ := didtypes.NewDidDocument("did:cash:subject")
					return dd
				},
				methodID:      "did:cash:subject#key-1 invalid ",
				relationships: []string{},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: false,
			params: params{
				malleate: func() didtypes.DidDocument {
					dd, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-1",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{},
						),
					))
					return dd
				},
				methodID: "did:cash:subject#key-1",
				relationships: []string{
					string(didtypes.AssertionMethod),
					string(didtypes.AssertionMethod), // test duplicated relationship
					string(didtypes.AssertionMethod), // test duplicated relationship
					string(didtypes.AssertionMethod), // test duplicated relationship
				},
			},

			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},
				AssertionMethod: []string{"did:cash:subject#key-1"},
			},
		},
		{
			wantErr: false, // different delete scenarios
			params: params{
				malleate: func() didtypes.DidDocument {
					dd, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-1",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{},
						),
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-2",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
							},
							[]string{},
						),
					))
					return dd
				},
				methodID:      "did:cash:subject#key-1",
				relationships: []string{string(didtypes.AssertionMethod)},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-2",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},
				Authentication:  []string{"did:cash:subject#key-2"},
				AssertionMethod: []string{"did:cash:subject#key-1"},
			},
		},
		{
			wantErr: false, // different delete scenarios
			params: params{
				malleate: func() didtypes.DidDocument {
					dd, _ := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-2",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
							},
							[]string{},
						),
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-3",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
							},
							[]string{},
						),
						didtypes.NewVerification(
							didtypes.VerificationMethod{
								"did:cash:subject#key-1",
								didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
								"did:cash:subject",
								&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
							},
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							[]string{},
						),
					))
					return dd
				},
				methodID:      "did:cash:subject#key-1",
				relationships: []string{string(didtypes.AssertionMethod)},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				VerificationMethod: []*didtypes.VerificationMethod{
					{
						"did:cash:subject#key-2",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-3",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					{
						"did:cash:subject#key-1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:cash:subject",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
				},

				Authentication:  []string{"did:cash:subject#key-2", "did:cash:subject#key-3"},
				AssertionMethod: []string{"did:cash:subject#key-1"},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_SetVerificationRelationships#", i), func(t *testing.T) {
			didDoc := tt.params.malleate()
			err := didDoc.SetVerificationRelationships(tt.params.methodID, tt.params.relationships...)

			if tt.wantErr {
				require.NotNil(t, err, "test: TestDidDocument_SetVerificationRelationships#%v", i)
				return
			}

			require.Nil(t, err, "test: TestDidDocument_SetVerificationRelationships#%v", i)
			assert.Equal(t, tt.wantDid, didDoc)

		})
	}
}

func TestDidDocument_HasRelationship(t *testing.T) {

	type params struct {
		didFn         func() didtypes.DidDocument
		signer        didtypes.BlockchainAccountID
		relationships []string
	}
	tests := []struct {
		name                    string
		params                  params
		expectedHasRelationship bool
	}{
		{
			name:                    "PASS: has relationships (multibase)",
			expectedHasRelationship: true,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								string(didtypes.Authentication),
								string(didtypes.KeyAgreement),
							},
							nil,
						),
					))
					assert.NoError(t, err)
					return dd
				},
				signer: didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
				relationships: []string{
					string(didtypes.AssertionMethod),
					string(didtypes.Authentication),
				},
			},
		},
		{
			name:                    "PASS: relationships missing (multibase, blockchainaccountid, hex)",
			expectedHasRelationship: false,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							nil,
						),
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:controller-1#key-2",
								"did:cash:controller-1",
								didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
							),
							[]string{
								didtypes.CapabilityDelegation,
							},
							nil,
						),
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-3",
								"did:cash:subject",
								didtypes.NewPublicKeyHex([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							nil,
						),
					))
					assert.NoError(t, err)
					return dd
				},
				signer: didtypes.NewBlockchainAccountID("cash", "subject"),
				relationships: []string{
					string(didtypes.CapabilityDelegation),
				},
			},
		},
		{
			name:                    "PASS: relationships missing (blockchainaccountid)",
			expectedHasRelationship: false,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject")
					assert.NoError(t, err)
					return dd
				},
				signer: didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
				relationships: []string{
					string(didtypes.CapabilityDelegation),
				},
			},
		},
		{
			name:                    "PASS: relationships missing (Multibase)",
			expectedHasRelationship: false,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							nil,
						),
					))
					assert.NoError(t, err)
					return dd
				},
				signer:        didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
				relationships: nil,
			},
		},
		{
			name:                    "PASS: has relationship (BlockchainAccountID)",
			expectedHasRelationship: true,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyMultibase([]byte("00dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"), didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
							},
							nil,
						),
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-2",
								"did:cash:subject",
								didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
							),
							[]string{
								didtypes.KeyAgreement,
							},
							nil,
						),
					))
					assert.NoError(t, err)
					return dd
				},
				signer: didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
				relationships: []string{
					string(didtypes.KeyAgreement),
				},
			},
		},
		{
			name:                    "PASS:  missing relationship (PublicKeyHex)",
			expectedHasRelationship: false,
			params: params{
				didFn: func() didtypes.DidDocument {
					dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
						didtypes.NewVerification(
							didtypes.NewVerificationMethod(
								"did:cash:subject#key-1",
								"did:cash:subject",
								didtypes.NewPublicKeyHex([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{
								didtypes.Authentication,
								didtypes.KeyAgreement,
							},
							nil,
						),
					))
					assert.NoError(t, err)
					return dd
				},
				signer:        didtypes.NewBlockchainAccountID("cash", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
				relationships: nil,
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_SetVerificationRelationships#", i), func(t *testing.T) {
			didDoc := tt.params.didFn()
			gotHasRelationship := didDoc.HasRelationship(tt.params.signer, tt.params.relationships...)
			assert.Equal(t, tt.expectedHasRelationship, gotHasRelationship)
		})
	}
}

func TestDidDocument_AddServices(t *testing.T) {
	type params struct {
		malleate func() didtypes.DidDocument // build a did document
		services []*didtypes.Service         // input list of verifications
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument // expected result
		wantErr bool
	}{
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject")
					return d
				},
				[]*didtypes.Service{
					didtypes.NewService(
						"agent:abc",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					),
					didtypes.NewService(
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					),
				},
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				Service: []*didtypes.Service{
					didtypes.NewService(
						"agent:abc",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					),
					didtypes.NewService(
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					),
				},
			},
		},
		{
			wantErr: true, // duplicated existing service id
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument(
						"did:cash:subject",
						didtypes.WithServices(
							didtypes.NewService(
								"agent:xyz",
								"DIDCommMessaging",
								"https://agent.xyz/1234",
							),
						),
					)
					return d
				},
				[]*didtypes.Service{
					{
						"agent:abc",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					}, {
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					},
				},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: true, // duplicated new service id
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject")
					return d
				},
				[]*didtypes.Service{
					{
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					}, {
						"agent:xyz",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					},
				},
			},
			wantDid: didtypes.DidDocument{},
		},
		{
			wantErr: true, // fail validation
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject")
					return d
				},
				[]*didtypes.Service{
					{
						"agent:abc",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					}, {
						"",
						"DIDCommMessaging",
						"https://agent.xyz/1234",
					},
				},
			},
			wantDid: didtypes.DidDocument{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_AddServices#", i), func(t *testing.T) {
			gotDid := tt.params.malleate()

			err := gotDid.AddServices(tt.params.services...)

			if tt.wantErr {
				require.NotNil(t, err, "test: TestDidDocument_AddServices#%v", i)
				return
			}

			require.Nil(t, err, "test: TestDidDocument_AddServices#%v", i)
			assert.Equal(t, tt.wantDid, gotDid)
		})
	}
}

func TestDidDocument_DeleteService(t *testing.T) {
	type params struct {
		didFn    func() didtypes.DidDocument // build a did document
		methodID string                      // input list of verifications
	}
	tests := []struct {
		params  params
		wantDid didtypes.DidDocument // expected result
		wantErr bool
	}{
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithServices(
							&didtypes.Service{
								"agent:abc",
								"DIDCommMessaging",
								"https://agent.abc/1234",
							},
						),
					)
					return d
				},
				"agent:abc",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
			},
		},
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithServices(
							&didtypes.Service{
								"agent:zyz",
								"DIDCommMessaging",
								"https://agent.abc/1234",
							},
							&didtypes.Service{
								"agent:abc",
								"DIDCommMessaging",
								"https://agent.abc/1234",
							},
						),
					)
					return d
				},
				"agent:abc",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				Service: []*didtypes.Service{
					{
						"agent:zyz",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					},
				},
			},
		},
		{
			wantErr: false,
			params: params{
				func() didtypes.DidDocument {
					d, _ := didtypes.NewDidDocument("did:cash:subject",
						didtypes.WithServices(
							&didtypes.Service{
								"agent:zyz",
								"DIDCommMessaging",
								"https://agent.abc/1234",
							},
							&didtypes.Service{
								"agent:abc",
								"DIDCommMessaging",
								"https://agent.abc/1234",
							},
							&didtypes.Service{
								"agent:007",
								"DIDCommMessaging",
								"https://agent.abc/007",
							},
						),
					)
					return d
				},
				"agent:abc",
			},
			wantDid: didtypes.DidDocument{
				Context: []string{contextDIDBase},
				Id:      "did:cash:subject",
				Service: []*didtypes.Service{
					{
						"agent:zyz",
						"DIDCommMessaging",
						"https://agent.abc/1234",
					}, {
						"agent:007",
						"DIDCommMessaging",
						"https://agent.abc/007",
					},
				},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("TestDidDocument_DeleteService#", i), func(t *testing.T) {
			gotDid := tt.params.didFn()

			gotDid.DeleteService(tt.params.methodID)

			assert.Equal(t, tt.wantDid, gotDid)
		})
	}
}

func TestBlockchainAccountID_GetAddress(t *testing.T) {
	tests := []struct {
		name string
		baID didtypes.BlockchainAccountID
		want string
	}{
		{
			"PASS: can get address",
			didtypes.BlockchainAccountID("cosmos:foochain:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
		},
		{
			// TODO: this should result in an error
			"PASS: address is empty",
			didtypes.BlockchainAccountID("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
			"",
		},
		{
			// TODO: this should result in an error
			"PASS: can get address (but address is wrong)",
			didtypes.BlockchainAccountID("cosmos:foochain:whatever"),
			"whatever",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.baID.GetAddress(), "GetAddress()")
		})
	}
}

func TestNewPublicKeyMultibaseFromHex(t *testing.T) {
	type args struct {
		pubKeyHex string
		vmType    didtypes.VerificationMaterialType
	}
	tests := []struct {
		name    string
		args    args
		wantPkm didtypes.PublicKeyMultibase
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"PASS: key match",
			args{
				pubKeyHex: "03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7",
				vmType:    didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019,
			},
			didtypes.NewPublicKeyMultibase(
				[]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215},
				didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
			assert.NoError,
		},
		{
			"FAIL: invalid hex key",
			args{
				pubKeyHex: "not hex string",
				vmType:    didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019,
			},
			didtypes.NewPublicKeyMultibase(nil, ""),
			assert.Error, // TODO: check the error message
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPkm, err := didtypes.NewPublicKeyMultibaseFromHex(tt.args.pubKeyHex, tt.args.vmType)
			if !tt.wantErr(t, err, fmt.Sprintf("NewPublicKeyMultibaseFromHex(%v, %v)", tt.args.pubKeyHex, tt.args.vmType)) {
				return
			}
			assert.Equalf(t, tt.wantPkm, gotPkm, "NewPublicKeyMultibaseFromHex(%v, %v)", tt.args.pubKeyHex, tt.args.vmType)

		})
	}
}

func TestDidDocument_HasPublicKey(t *testing.T) {

	tests := []struct {
		name   string
		didFn  func() didtypes.DidDocument
		pubkey func() types.PubKey
		want   bool
	}{
		{
			"PASS: has public key (multibase)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{2, 201, 95, 248, 187, 133, 206, 97, 166, 70, 229, 226, 88, 124, 29, 43, 70, 3, 244, 225, 19, 128, 44, 132, 110, 15, 15, 35, 40, 189, 237, 71, 245}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			true,
		},
		{
			"PASS: doesn't have public key (multibase)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			false,
		},
		{
			"PASS: has public key (blockchainAccount)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewBlockchainAccountID("foochain", "cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf"),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			true,
		},
		{
			"PASS: doesn't have public key (blockchainAccountId)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewBlockchainAccountID("foochain", "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			false,
		},
		{
			"PASS: has public key (publicKeyHex)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyHex([]byte{2, 201, 95, 248, 187, 133, 206, 97, 166, 70, 229, 226, 88, 124, 29, 43, 70, 3, 244, 225, 19, 128, 44, 132, 110, 15, 15, 35, 40, 189, 237, 71, 245}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			true,
		},
		{
			"PASS: doesn't have public key (pubKeyHex)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyHex([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			func() types.PubKey {
				var pk types.PubKey
				c := app.MakeEncodingConfig().Codec
				err := c.UnmarshalInterfaceJSON([]byte(`{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aslf+LuFzmGmRuXiWHwdK0YD9OETgCyEbg8PIyi97Uf1"}`), &pk)
				assert.NoError(t, err)
				return pk

			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			didDoc := tt.didFn()
			pubKey := tt.pubkey()
			assert.Equalf(t, tt.want, didDoc.HasPublicKey(pubKey), "HasPublicKey(%v)", pubKey)
		})
	}
}

func TestDidDocument_GetVerificationMethodBlockchainAddress(t *testing.T) {
	tests := []struct {
		name        string
		didFn       func() didtypes.DidDocument
		methodID    string
		wantAddress string
		wantErr     error
	}{
		{
			"PASS: get address (PublicKeyMultibase)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			"did:cash:subject#key-1",
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			nil,
		},
		{
			"PASS: get address (PublicKeyHex)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewPublicKeyHex([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			"did:cash:subject#key-1",
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			nil,
		},
		{
			"PASS: get address (BlockchainAccountID)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewBlockchainAccountID("foochain", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			"did:cash:subject#key-1",
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			nil,
		},
		{
			"PASS: get address (BlockchainAccountID)",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument("did:cash:subject", didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cash:subject#key-1",
							"did:cash:subject",
							didtypes.NewBlockchainAccountID("foochain", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
						),
						[]string{
							didtypes.Authentication,
							didtypes.KeyAgreement,
						},
						nil,
					),
				))
				assert.NoError(t, err)
				return dd
			},
			"did:cash:subject#key-2",
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			didtypes.ErrVerificationMethodNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			didDoc := tt.didFn()
			gotAddress, err := didDoc.GetVerificationMethodBlockchainAddress(tt.methodID)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equalf(t, tt.wantAddress, gotAddress, "GetVerificationMethodBlockchainAddress(%v)", tt.methodID)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDidDocument_HasController(t *testing.T) {

	tests := []struct {
		name          string
		didFn         func() didtypes.DidDocument
		controllerDID string
		want          bool
	}{
		{
			"PASS: controller found",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument(
					"did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
						"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
					),
				)
				assert.NoError(t, err)
				return dd
			},
			"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
			true,
		},
		{
			"PASS: controller not found",
			func() didtypes.DidDocument {
				dd, err := didtypes.NewDidDocument(
					"did:cash:subject",
					didtypes.WithControllers(
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
						"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
					),
				)
				assert.NoError(t, err)
				return dd
			},
			"did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			didDoc := tt.didFn()
			assert.Equalf(t, tt.want, didDoc.HasController(didtypes.DID(tt.controllerDID)), "HasController(%v)", tt.controllerDID)
		})
	}
}
