package transaction

import (
	"context"
	"fmt"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/getters"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/page"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/regources/generated"
	"time"
)

const (
	streamPageLimit = 100
)

type Streamer struct {
	log *logan.Entry
	getters.TransactionHandler
}

func NewStreamer(handler getters.TransactionHandler, log *logan.Entry) *Streamer {
	return &Streamer{
		log:                log,
		TransactionHandler: handler,
	}
}
func (s *Streamer) StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
) (<-chan regources.TransactionListResponse, <-chan error) {
	txChan := make(chan regources.TransactionListResponse)
	errChan := make(chan error)
	limit := fmt.Sprintf("%d", streamPageLimit)
	s.SetFilters(query.TransactionFilters{
		ChangeTypes: changeTypes,
		EntryTypes:  entryTypes,
	})
	s.SetPageParams(page.Params{
		Limit: &limit,
	})
	s.SetIncludes(query.TransactionIncludes{
		LedgerEntryChanges: true,
	})

	txPage, err := s.List()
	if err != nil {
		errChan <- err
		return txChan, errChan
	}
	if txPage == nil {
		errChan <- errors.New("got nil page")
		return txChan, errChan
	}
	go func() {
		defer func() {
			s.log.Info("Closing channels...")
			close(txChan)
			close(errChan)
		}()
		txChan <- *txPage
		ticker := time.NewTicker(5 * time.Second)
		running.WithBackOff(ctx, s.log, "transaction-streamer", func(ctx context.Context) error {
			if txPage == nil || len(txPage.Data) == 0 {
				// TODO: Find better way
				<-ticker.C
				txPage, err = s.Self()
			} else {
				txPage, err = s.Next()
			}
			if err != nil {
				errChan <- err
				return err
			}
			if txPage != nil {
				txChan <- *txPage
			} else {
				s.log.Warn("got nil page")
			}
			return nil
		}, time.Second, 2*time.Second, 10*time.Second)
	}()

	return txChan, errChan

}
