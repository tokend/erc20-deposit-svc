package addrstate

import (
	"gitlab.com/tokend/go/xdr"
)

type ExternalSystemBindingMutator struct {
	SystemType int32
}

func (e ExternalSystemBindingMutator) GetEffects() []int {
	return []int{int(xdr.LedgerEntryChangeTypeCreated), int(xdr.LedgerEntryChangeTypeRemoved)}
}

func (e ExternalSystemBindingMutator) GetEntryTypes() []int {
	return []int{int(xdr.LedgerEntryTypeExternalSystemAccountId)}
}

func (e ExternalSystemBindingMutator) GetStateUpdate(change xdr.LedgerEntryChange) (update StateUpdate) {
	switch change.Type {
	case xdr.LedgerEntryChangeTypeCreated:
		switch change.Created.Data.Type {
		case xdr.LedgerEntryTypeExternalSystemAccountId:
			data := change.Created.Data.ExternalSystemAccountId
			if int32(data.ExternalSystemType) != e.SystemType {
				break
			}
			update.ExternalAccount = &StateExternalAccountUpdate{
				ExternalType: e.SystemType,
				State:        ExternalAccountBindingStateCreated,
				Data:         string(data.Data),
				Address:      data.AccountId.Address(),
			}
		}
	case xdr.LedgerEntryChangeTypeRemoved:
		switch change.Removed.Type {
		case xdr.LedgerEntryTypeExternalSystemAccountId:
			data := change.Removed.ExternalSystemAccountId
			if int32(data.ExternalSystemType) != e.SystemType {
				break
			}
			update.ExternalAccount = &StateExternalAccountUpdate{
				ExternalType: e.SystemType,
				State:        ExternalAccountBindingStateDeleted,
				Address:      data.AccountId.Address(),
			}
		}
	}
	return
}
