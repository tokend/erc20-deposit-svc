package addrstate

import "gitlab.com/tokend/go/xdr"

type ExternalAccountBindingState int32

const (
	ExternalAccountBindingStateCreated ExternalAccountBindingState = iota + 1
	ExternalAccountBindingStateDeleted
)

// StateUpdate is a connector between LedgerEntryChange and Watcher state for specific consumers
type StateUpdate struct {
	ExternalAccount     *StateExternalAccountUpdate
	Balance             *StateBalanceUpdate
	AccountKYC          *StateAccountKYCUpdate
	AssetPair           *StateAssetPairUpdate
	Sale                *StateSaleUpdate
}

type StateAccountKYCUpdate struct {
	Address string
	KYCData string
}

type StateExternalAccountUpdate struct {
	// ExternalType external system accound id type
	ExternalType int32
	// Data external system pool entity data
	Data string
	// Address is a TokenD account address
	Address string
	// State shows current external pool entity binding state
	State ExternalAccountBindingState
}

type StateBalanceUpdate struct {
	Address string
	Balance string
	Asset   string
}

type StateAssetPairUpdate struct {
	Base         xdr.AssetCode
	Quote        xdr.AssetCode
	CurrentPrice int64
	PhysicalPrice int64
}

type StateSaleUpdate struct {
	BaseAsset xdr.AssetCode
}
