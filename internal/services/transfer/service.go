package transfer

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"math/big"
	"time"
)

type ERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log
}

//Run starts service
func (s *Service) Run(ctx context.Context) {
	err := s.prepare(ctx)
	if err != nil {
		s.log.WithError(err).Fatal("failed to prepare transfer listener")
	}
	go running.WithBackOff(ctx, s.log, "old-transfer-listener", func(ctx context.Context) error {
		err = s.processOld(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to process old transfers")
		}
		s.log.Info("Finished streaming old transfers")
		return nil
	}, time.Minute, time.Minute, time.Hour)

	go running.WithBackOff(ctx, s.log, "new-transfer-listener", func(ctx context.Context) error {
		err = s.processNew(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to process new transfers")
		}
		s.log.Info("Finished streaming new transfers")
		return nil
	}, time.Minute, time.Minute, time.Hour)

}

func (s *Service) processOld(ctx context.Context) error {
	oldCh, oldSubscription, err := s.contract.FilterLogs(
		&bind.FilterOpts{
			Context: ctx,
			Start:   s.cfg.Checkpoint,
		}, "Transfer")
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to old logs")
	}
	defer oldSubscription.Unsubscribe()

runner:
	for {
		select {
		case event, ok := <-oldCh:
			if !ok {
				break runner
			}
			err = s.processTransfer(ctx, event)
			if err != nil {
				return errors.Wrap(err, "failed to process transfer")
			}
		case err := <-oldSubscription.Err():
			if err != nil {
				return errors.Wrap(err, "got error from subscription")
			}
		case <-ctx.Done():
			break runner
		}
	}
	return nil
}

func (s *Service) processNew(ctx context.Context) error {
	newCh, newSubscription, err := s.contract.WatchLogs(
		&bind.WatchOpts{
			Context: ctx,
		}, "Transfer")
	if err != nil {
		s.log.WithError(err).Error("failed to subscribe to new logs")
	}

	defer newSubscription.Unsubscribe()
runner:
	for {
		select {
		case event, ok := <-newCh:
			if !ok {
				return errors.New("channel closed unexpectedly")
			}
			err = s.processTransfer(ctx, event)
			if err != nil {
				return errors.Wrap(err, "failed to process transfer")
			}

		case err := <-newSubscription.Err():
			if err != nil {
				return errors.Wrap(err, "subscription returned error")
			}
		case <-ctx.Done():
			break runner
		}
	}

	return nil
}

func (s *Service) processTransfer(ctx context.Context, event types.Log) error {
	s.Lock()
	defer s.Unlock()
	parsed := new(ERC20Transfer)
	err := s.contract.UnpackLog(parsed, "Transfer", event)
	if err != nil {
		return errors.Wrap(err, "failed to unpack log", logan.F{
			"event_data":   event.Data,
			"event_topics": event.Topics,
		})
	}

	block, err := s.client.BlockByHash(ctx, event.BlockHash)
	if err != nil {
		return errors.Wrap(err, "failed to get block", logan.F{
			"block_hash":   event.BlockHash.String(),
			"block_number": event.BlockNumber,
		})
	}
	s.log.WithFields(logan.F{
		"amount":       parsed.Value,
		"tx_hash":      event.TxHash.String(),
		"block_number": event.BlockNumber,
	}).Info("Got transfer")

	s.ch <- Details{
		TransactionHash: event.TxHash.String(),
		Amount:          parsed.Value,
		Destination:     parsed.To,
		Decimals:        int64(s.decimals),
		BlockTime:       time.Unix(int64(block.Time()), 0),
	}

	return nil
}
