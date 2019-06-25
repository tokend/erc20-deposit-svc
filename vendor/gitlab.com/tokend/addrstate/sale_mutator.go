package addrstate

import "gitlab.com/tokend/go/xdr"

type SaleMutator struct {
}

func (b SaleMutator) GetEffects() []int {
	return []int{int(xdr.LedgerEntryChangeTypeCreated)}
}

func (b SaleMutator) GetEntryTypes() []int {
	return []int{int(xdr.LedgerEntryTypeSale)}
}

func (b SaleMutator) GetStateUpdate(change xdr.LedgerEntryChange) (update StateUpdate) {
	switch change.Type {
	case xdr.LedgerEntryChangeTypeCreated:
		data := change.Created.Data
		switch data.Type {
		case xdr.LedgerEntryTypeSale:
			update.Sale = &StateSaleUpdate{
				BaseAsset: data.Sale.BaseAsset,
			}
		}
	}
	return
}
