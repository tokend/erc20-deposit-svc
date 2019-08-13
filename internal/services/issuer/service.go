package issuer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/tokend/erc20-deposit-svc/internal/horizon/submit"
	"github.com/tokend/erc20-deposit-svc/internal/services/transfer"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/addrstate"
	"gitlab.com/tokend/go/hash"
	"gitlab.com/tokend/go/xdrbuild"
)

const (
	taskCheckTxConfirmed uint32 = 2048 // 2^11
)

func (s *Service) prepare(ctx context.Context) {
	s.addressProvider = addrstate.New(
		ctx,
		s.log,
		[]addrstate.StateMutator{
			addrstate.ExternalSystemBindingMutator{SystemType: s.asset.ExternalSystemType},
			addrstate.BalanceMutator{Asset: s.asset.ID},
		},
		s.streamer,
	)
}

func (s *Service) Run(ctx context.Context) {
	s.prepare(ctx)
	s.log.WithField("asset", s.asset.ID).Info("Started issuer service")
	running.WithBackOff(ctx, s.log, "issuer", func(ctx context.Context) error {
		for t := range s.ch {
			err := s.processTransfer(ctx, t)
			if err != nil {
				return errors.Wrap(err, "failed to process transfer")
			}
		}
		return nil
	}, 10*time.Second, 20*time.Second, time.Minute)
}

func (s *Service) processTransfer(ctx context.Context, transfer transfer.Details) error {
	destination := strings.ToLower(transfer.Destination.String())
	address := s.addressProvider.ExternalAccountAt(ctx, transfer.BlockTime, s.asset.ExternalSystemType, destination, nil)
	if address == nil {
		return nil
	}
	balance := s.addressProvider.Balance(ctx, *address, s.asset.ID)
	if balance == nil {
		s.log.Debug("Unable to find valid balance to issue tokens to")
		return nil
	}

	issueDetails := details{
		TxHash:      transfer.TransactionHash,
		Destination: transfer.Destination.String(),
	}
	detailsbb, err := json.Marshal(issueDetails)
	if err != nil {
		return errors.Wrap(err, "failed to marshal transfer details")
	}

	refHash := hash.Hash([]byte(transfer.TransactionHash))

	reference := hex.EncodeToString(refHash[:])

	amountToIssue := amountToIssue(transfer.Amount, transfer.Decimals, int64(s.asset.Attributes.TrailingDigits))

	if amountToIssue == 0 {
		s.log.WithFields(logan.F{
			"amount":      transfer.Amount.String(),
			"transaction": transfer.TransactionHash,
		}).Warn("amount to issue is too small")
		return nil
	}

	tasks := taskCheckTxConfirmed
	envelope, err := s.builder.Transaction(s.owner).Op(xdrbuild.CreateIssuanceRequest{
		Reference: reference,
		Asset:     s.asset.ID,
		Amount:    amountToIssue,
		Receiver:  *balance,
		Details:   json.RawMessage(detailsbb),
		AllTasks:  &tasks,
	}).Sign(s.issuer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to craft transaction")
	}
	err = s.submitEnvelope(ctx, envelope)
	if err != nil {
		return errors.Wrap(err, "failed to submit issuance tx")
	}
	s.log.WithFields(logan.F{
		"amount": amountToIssue,
	}).Debug("Successfully submitted issuance")
	return nil
}

func (s *Service) submitEnvelope(ctx context.Context, envelope string) error {
	_, err := s.txSubmitter.Submit(ctx, envelope, false)
	if submitFailure, ok := err.(submit.TxFailure); ok {
		if len(submitFailure.OperationResultCodes) == 1 &&
			submitFailure.OperationResultCodes[0] == "op_reference_duplication" {
			return nil
		}
	}
	if err != nil {
		return errors.Wrap(err, "Horizon SubmitResult has error")
	}
	return nil
}

func amountToIssue(am *big.Int, decimals int64, trailingDigits int64) uint64 {
	var toIssue uint64
	if decimals > trailingDigits {
		toIssue = am.Uint64() / big.NewInt(1).Exp(big.NewInt(10), big.NewInt(decimals-trailingDigits), nil).Uint64()
	} else {
		toIssue = am.Uint64() * big.NewInt(1).Exp(big.NewInt(10), big.NewInt(trailingDigits-decimals), nil).Uint64()
	}
	return toIssue
}
