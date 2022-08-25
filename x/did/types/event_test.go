package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	didtypes "github.com/fetchai/fetchd/x/did/types"
)

func TestNewDidDocumentCreatedEvent(t *testing.T) {
	type args struct {
		did   string
		owner string
	}
	tests := []struct {
		name string
		args args
		want *didtypes.DidDocumentCreatedEvent
	}{
		{
			"PASS: did created event",
			args{
				did:   "did:cosmos:net:foochain:123",
				owner: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			&didtypes.DidDocumentCreatedEvent{
				Did:    "did:cosmos:net:foochain:123",
				Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, didtypes.NewDidDocumentCreatedEvent(tt.args.did, tt.args.owner), "NewDidDocumentCreatedEvent(%v, %v)", tt.args.did, tt.args.owner)
		})
	}
}

func TestNewDidDocumentUpdatedEvent(t *testing.T) {
	type args struct {
		did    string
		signer string
	}
	tests := []struct {
		name string
		args args
		want *didtypes.DidDocumentUpdatedEvent
	}{
		{
			"PASS: did updated event",
			args{
				did:    "did:cosmos:net:foochain:123",
				signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
			&didtypes.DidDocumentUpdatedEvent{
				Did:    "did:cosmos:net:foochain:123",
				Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, didtypes.NewDidDocumentUpdatedEvent(tt.args.did, tt.args.signer), "NewDidDocumentUpdatedEvent(%v, %v)", tt.args.did, tt.args.signer)
		})
	}
}
