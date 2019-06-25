package addrstate

import (
	"context"
	"gitlab.com/tokend/regources/generated"
	"time"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdr"
)

// StateMutator uses to get StateUpdate for specific effects and entryTypes
type StateMutator interface {
	GetStateUpdate(change xdr.LedgerEntryChange) StateUpdate
	GetEffects() []int
	GetEntryTypes() []int
}

// StreamTransactions streams transactions fetched for specified filters.
type TXStreamer interface {
	StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
	) (<-chan regources.TransactionListResponse, <-chan error)
}

// Watcher watches what comes from txStreamer and what StateMutators do
type Watcher struct {
	log        *logan.Entry
	mutators   []StateMutator
	txStreamer TXStreamer
	ctx        context.Context

	// internal state
	head       time.Time
	headUpdate chan struct{}
	state      *State
}

// New returns new watcher and start streaming transactionsV2
func New(ctx context.Context, log *logan.Entry, mutators []StateMutator, txQ TXStreamer) *Watcher {
	ctx, cancel := context.WithCancel(ctx)

	w := &Watcher{
		log:        log.WithField("service", "addrstate"),
		mutators:   mutators,
		txStreamer: txQ,
		ctx:        ctx,

		state:      newState(),
		headUpdate: make(chan struct{}),
	}

	go func() {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.WithRecover(rvr).Error("state watcher panicked")
			}
			cancel()
		}()
		w.run(ctx)
	}()

	return w
}

func (w *Watcher) ensureReached(ctx context.Context, ts time.Time) {
	for w.head.Before(ts) {
		select {
		case <-ctx.Done():
			return
		case <-w.headUpdate:
			// Make the for check again
			continue
		}
	}
}

func (w *Watcher) run(ctx context.Context) {
	var entryTypes []int
	var changes []int

	for _, mutator := range w.mutators {
		entryTypes = append(entryTypes, mutator.GetEntryTypes()...)
		changes = append(changes, mutator.GetEffects()...)
	}

	// there is intentionally no recover, it should just die in case of persistent error
	txStream, txStreamErrs := w.txStreamer.StreamTransactions(ctx, changes, entryTypes)

	for {
		select {
		case txResp := <-txStream:
			txs := txResp.Data
			for _, tx := range txs {
				// go through all ledger changes
				changes := tx.Relationships.LedgerEntryChanges
				if changes != nil {
					for _, changeKey := range changes.Data {
						// apply all mutators
						change := txResp.Included.MustLedgerEntryChange(changeKey)
						ledgerEntryChange, err := convertLedgerEntryChange(*change)
						if err != nil {
							w.log.WithError(err).Error("failed to get state update", logan.F{
								"entry_type":     change.Attributes.EntryType,
								"effect":         change.Attributes.ChangeType,
								"transaction_id": tx.ID,
							})
							return
						}
						for _, mutator := range w.mutators {
							stateUpdate := mutator.GetStateUpdate(ledgerEntryChange)
							w.state.Mutate(tx.Attributes.CreatedAt, stateUpdate)
						}
					}
				}

				// if we made it here it's safe to bump head cursor
				w.head = txResp.Meta.LatestLedgerCloseTime

				// must be in select to listen new updates.
				// If not gets all updates before time it was run and then w.headUpdate locks
				select {
				case w.headUpdate <- struct{}{}:
				default:
				}
			}
		case err := <-txStreamErrs:
			w.log.WithError(err).Warn("TXStreamer sent error into channel.")
		}
	}
}
