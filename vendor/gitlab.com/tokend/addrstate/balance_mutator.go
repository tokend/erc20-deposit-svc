package addrstate

import (
	"gitlab.com/tokend/go/xdr"
)

type BalanceMutator struct {
	Asset string
}

func (b BalanceMutator) GetEffects() []int {
	return []int{int(xdr.LedgerEntryChangeTypeCreated)}
}

func (b BalanceMutator) GetEntryTypes() []int {
	return []int{int(xdr.LedgerEntryTypeBalance)}
}

func (b BalanceMutator) GetStateUpdate(change xdr.LedgerEntryChange) (update StateUpdate) {
	switch change.Type {
	case xdr.LedgerEntryChangeTypeCreated:
		switch change.Created.Data.Type {
		case xdr.LedgerEntryTypeBalance:
			data := change.Created.Data.Balance
			if string(data.Asset) != b.Asset {
				break
			}
			update.Balance = &StateBalanceUpdate{
				Address: data.AccountId.Address(),
				Balance: data.BalanceId.AsString(),
				Asset:   string(data.Asset),
			}
		}
	}
	return
}

