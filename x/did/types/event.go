package types

// NewDidDocumentCreatedEvent constructs a new did_created sdk.Event
func NewDidDocumentCreatedEvent(did, owner string) *DidDocumentCreatedEvent {
	return &DidDocumentCreatedEvent{
		Did:    did,
		Signer: owner,
	}
}

// NewDidDocumentUpdatedEvent constructs a new did_created sdk.Event
func NewDidDocumentUpdatedEvent(did, signer string) *DidDocumentUpdatedEvent {
	return &DidDocumentUpdatedEvent{
		Did:    did,
		Signer: signer,
	}
}
