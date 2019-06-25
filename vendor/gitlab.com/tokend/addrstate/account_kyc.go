package addrstate

import "gitlab.com/tokend/go/xdr"

type AccountKYCMutator struct{}

func (m AccountKYCMutator) GetEffects() []int {
	return []int{int(xdr.LedgerEntryChangeTypeUpdated), int(xdr.LedgerEntryChangeTypeCreated)}
}

func (m AccountKYCMutator) GetEntryTypes() []int {
	return []int{int(xdr.LedgerEntryTypeAccountKyc)}
}

func (m *AccountKYCMutator) GetStateUpdate(change xdr.LedgerEntryChange) (update StateUpdate) {
	switch change.Type {
	case xdr.LedgerEntryChangeTypeCreated:
		data := change.Created.Data
		switch data.Type {
		case xdr.LedgerEntryTypeAccountKyc:
			update.AccountKYC = &StateAccountKYCUpdate{
				Address: data.AccountKyc.AccountId.Address(),
				KYCData: string(data.AccountKyc.KycData),
			}
		}
	case xdr.LedgerEntryChangeTypeUpdated:
		data := change.Updated.Data
		switch data.Type {
		case xdr.LedgerEntryTypeAccountKyc:
			update.AccountKYC = &StateAccountKYCUpdate{
				Address: data.AccountKyc.AccountId.Address(),
				KYCData: string(data.AccountKyc.KycData),
			}
		}
	}
	return update
}
