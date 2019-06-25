package addrstate

import "gitlab.com/tokend/go/xdr"

type AssetPairMutator struct {
}

func (b AssetPairMutator) GetEffects() []int {
	return []int{int(xdr.LedgerEntryChangeTypeUpdated), int(xdr.LedgerEntryChangeTypeCreated)}
}

func (b AssetPairMutator) GetEntryTypes() []int {
	return []int{int(xdr.LedgerEntryTypeAssetPair)}
}

func (b AssetPairMutator) GetStateUpdate(change xdr.LedgerEntryChange) (update StateUpdate) {
	switch change.Type {
	case xdr.LedgerEntryChangeTypeCreated:
		data := change.Created.Data
		switch data.Type {
		case xdr.LedgerEntryTypeAssetPair:
			update.AssetPair = &StateAssetPairUpdate{
				Base:         data.AssetPair.Base,
				Quote:        data.AssetPair.Quote,
				CurrentPrice: int64(data.AssetPair.CurrentPrice),
				PhysicalPrice: int64(data.AssetPair.PhysicalPrice),
			}
		}

	case xdr.LedgerEntryChangeTypeUpdated:
		data := change.Updated.Data
		switch data.Type {
		case xdr.LedgerEntryTypeAssetPair:
			update.AssetPair = &StateAssetPairUpdate{
				Base:         data.AssetPair.Base,
				Quote:        data.AssetPair.Quote,
				CurrentPrice: int64(data.AssetPair.CurrentPrice),
				PhysicalPrice: int64(data.AssetPair.PhysicalPrice),
			}
		}
	}

	return
}
