package addrstate

import (
	"context"
	"time"

	"gitlab.com/tokend/go/xdr"
)

func (w *Watcher) ExternalAccountAt(ctx context.Context, ts time.Time, systemType int32, address string, payload *string) *string {
	w.ensureReached(ctx, ts)

	w.state.RLock()
	defer w.state.RUnlock()

	if _, ok := w.state.external[systemType]; !ok {
		// external system states doesn't exist (yet)
		return nil
	}

	externalDataType := "address"
	if payload != nil {
		externalDataType = "address_with_payload"
	}
	states := w.state.external[systemType][ExternalData{
		Type: externalDataType,
		Data: ExternalDataEntry{Address: address, Payload: payload},
	}]
	if len(states) == 0 {
		return nil
	}
	// iterating through the closed periods
	for i := 0; i < len(states)-1; i += 1 {
		a := states[i]
		b := states[i+1]
		if ts.After(a.UpdatedAt) && ts.Before(b.UpdatedAt) {
			// we found time interval that includes our ts,
			// first states is current one
			addr := a.Address
			return &addr
		}
	}
	// checking last known state
	lastState := states[len(states)-1]
	if ts.After(lastState.UpdatedAt) && lastState.State == ExternalAccountBindingStateCreated {
		addr := lastState.Address
		return &addr
	}
	// seems like rogue deposit, but who cares
	return nil
}

// BindExternalSystemEntities returns all known external data for systemType
func (w *Watcher) BindedExternalSystemEntities(ctx context.Context, systemType int32) (result []ExternalData) {
	w.ensureReached(ctx, time.Now())

	w.state.RLock()
	defer w.state.RUnlock()

	if _, ok := w.state.external[systemType]; !ok {
		return result
	}

	entities := w.state.external[systemType]
	for entity, _ := range entities {
		result = append(result, entity)
	}
	return result
}

func (w *Watcher) KYCData(ctx context.Context, address string) *string {
	w.ensureReached(ctx, time.Now())

	w.state.RLock()
	defer w.state.RUnlock()

	kycData, ok := w.state.accountKYC[address]
	if !ok {
		return nil
	}

	return &kycData
}

func (w *Watcher) Balance(ctx context.Context, address string, asset string) *string {
	w.state.RLock()

	// let's hope for the best and try to get balance before reaching head
	if w.state.balances[address] != nil {
		if balance, ok := w.state.balances[address][asset]; ok {
			return &balance
		}
	}

	w.state.RUnlock()

	// if we don't have balance already, let's wait for latest ledger
	w.ensureReached(ctx, time.Now())

	w.state.RLock()
	defer w.state.RUnlock()

	// now check again
	if w.state.balances[address] != nil {
		if balance, ok := w.state.balances[address][asset]; ok {
			return &balance
		}
	}

	// seems like user doesn't have balance in provided asset atm
	return nil
}

type AssetPairEvent struct {
	Pair          string
	UpdatedAt     time.Time
	Price         int64
	PhysicalPrice int64
}

// TODO add someone responsible for closing the channel. probably need to store context alongside channel
func (w *Watcher) AssetPair(ctx context.Context) chan AssetPairEvent {
	w.ensureReached(ctx, time.Now())
	w.state.RLock()

	cn := make(chan AssetPairEvent)
	go func() {
		defer w.state.RUnlock()
		for assetPair, price := range w.state.assetPair {
			for _, p := range price {
				cn <- AssetPairEvent{Pair: assetPair.String(), UpdatedAt: p.UpdatedAt, Price: p.Value, PhysicalPrice: p.PhysicalPrice}
			}
		}
		w.state.chAssetPairs = append(w.state.chAssetPairs, cn)
	}()
	return cn
}

// AssetPairPriceAt finds what price was for asset pair at a given time
func (w *Watcher) AssetPairPriceAt(ctx context.Context, base, quote string, ts time.Time) *int64 {
	w.ensureReached(ctx, time.Now())

	w.state.RLock()
	defer w.state.RUnlock()

	ap := AssetPair{Base: xdr.AssetCode(base), Quote: xdr.AssetCode(quote)}
	states := w.state.assetPair[ap]

	// iterating through the closed periods
	for i := 0; i < len(states)-1; i += 1 {
		a := states[i]
		b := states[i+1]
		if ts.After(a.UpdatedAt) && ts.Before(b.UpdatedAt) || ts.Equal(a.UpdatedAt) {
			// we found time interval that includes our ts,
			// first states is current one
			return &a.Value
		}
	}
	// checking last known state
	lastState := states[len(states)-1]
	if ts.After(lastState.UpdatedAt) || ts.Equal(lastState.UpdatedAt) {
		return &lastState.Value
	}
	return nil
}

type SaleEvent struct {
	BaseAsset string
	UpdatedAt time.Time
}

func (w *Watcher) Sale(ctx context.Context) chan SaleEvent {
	w.ensureReached(ctx, time.Now())
	w.state.RLock()
	cn := make(chan SaleEvent)
	go func() {
		defer w.state.RUnlock()
		for asset, ts := range w.state.sale {
			cn <- SaleEvent{BaseAsset: asset, UpdatedAt: ts}
		}
		w.state.chSale = append(w.state.chSale, cn)
	}()
	return cn

}
