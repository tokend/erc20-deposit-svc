package transfer

import (
	"context"
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
	s.processOld(ctx)
	s.log.Info("Finished streaming old transfers")
	s.processNew(ctx)
	s.log.Info("Finished streaming new transfers")
	s.oldSubscription.Unsubscribe()
	s.newSubscription.Unsubscribe()
}

func (s *Service) processOld(ctx context.Context) {
	running.UntilSuccess(ctx, s.log, "old-transfer-streamer", func(ctx context.Context) (bool, error) {
	runner:
		for {
			if len(s.old) == 0 {
				break runner
			}

			select {
			case event, ok := <-s.old:
				if !ok {
					break runner
				}
				s.processTransfer(ctx, event)
			case err := <-s.oldSubscription.Err():
				if err != nil {
					s.log.WithError(err).Warn("got error from subscription")
				}
			case <-ctx.Done():
				break runner
			}
		}
		return true, nil
	}, time.Minute, time.Hour)
}

func (s *Service) processNew(ctx context.Context) {
	running.WithBackOff(ctx, s.log, "transfer-streamer", func(ctx context.Context) error {
		select {
		case event, ok := <-s.new:
			if !ok {
				return errors.New("Channel closed unexpectedly")
			}
			s.processTransfer(ctx, event)
		case err := <-s.newSubscription.Err():
			if err != nil {
				return errors.Wrap(err, "Subscription returned error")
			}
		}
		return nil
	}, time.Second, 20*time.Second, 5*time.Minute)
}

func (s *Service) processTransfer(ctx context.Context, event types.Log) {
	parsed := new(ERC20Transfer)
	err := s.contract.UnpackLog(parsed, "Transfer", event)
	if err != nil {
		s.log.WithError(err).Error("failed to unpack log")
		return
	}
	block, err := s.client.BlockByHash(ctx, event.BlockHash)
	if err != nil {
		s.log.WithFields(logan.F{
			"block_hash":   event.BlockHash.String(),
			"block_number": event.BlockNumber,
		}).WithError(err).Error("failed to get block")
		return
	}
	s.log.WithField("amount", parsed.Value).Info("got event")

	s.ch <- Details{
		TransactionHash: event.TxHash.String(),
		Amount:          parsed.Value,
		Destination:     parsed.To,
		Decimals:        int64(s.decimals),
		BlockTime:       time.Unix(int64(block.Time()), 0),
	}
}
