package addrstate

import (
	"encoding/json"
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
			externalData := ExternalData{}
			databb := []byte(data.Data)
			if err := json.Unmarshal(databb, &externalData); err != nil {
				//todo add logging of some sort
				break
			}
			update.ExternalAccount = &StateExternalAccountUpdate{
				ExternalType: e.SystemType,
				State:        ExternalAccountBindingStateCreated,
				Data:         externalData,
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
