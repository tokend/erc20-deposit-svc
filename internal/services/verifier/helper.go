package verifier

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
	"math/big"
)

const (
	taskCheckTxConfirmed uint32 = 2048 // 2^11

	//Request state
	reviewableRequestStatePending = 1
	//page size
	requestPageSizeLimit = 10

	invalidDetails = "Invalid external details"
	invalidTXHash  = "Invalid ethereum transaction hash"
	txFailed       = "Transaction failed"
)

type SentDetails struct {
	EthTxHash string `json:"eth_tx_hash"`
}

func (s *Service) confirmIssuanceSuccessful(ctx context.Context, request regources.ReviewableRequest, details *regources.CreateIssuanceRequest) error {
	detailsbb := []byte(details.Attributes.CreatorDetails)
	issuanceDetails := SentDetails{}
	err := json.Unmarshal(detailsbb, &issuanceDetails)
	if err != nil {
		s.log.WithField("request_id", request.ID).WithError(err).Warn("Unable to unmarshal creator details")
		return s.permanentReject(ctx, request, invalidDetails)
	}

	if issuanceDetails.EthTxHash == "" {
		s.log.
			WithField("external_details", request.Attributes.ExternalDetails).
			Warn("tx hash missing")
		return s.permanentReject(ctx, request, invalidTXHash)
	}

	receipt, err := s.client.TransactionReceipt(ctx, common.HexToHash(issuanceDetails.EthTxHash))
	if err == ethereum.NotFound {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to get transaction receipt", logan.F{
			"tx_hash": issuanceDetails.EthTxHash,
		})
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return s.permanentReject(ctx, request, txFailed)
	}

	if !s.ensureEnoughConfirmations(ctx, receipt.BlockNumber.Int64()) {
		return nil
	}

	err = s.approveRequest(ctx, request, 0, taskCheckTxConfirmed, map[string]interface{}{
		"eth_block_number": receipt.BlockNumber.Int64(),
	})

	if err != nil {
		return errors.Wrap(err, "failed to review request second time", logan.F{"request_id": request.ID})
	}

	return nil
}

func (s *Service) ensureEnoughConfirmations(ctx context.Context, blockNumber int64) bool {
	_, err := s.client.BlockByNumber(ctx, big.NewInt(blockNumber+s.ethCfg.Confirmations))
	if err == ethereum.NotFound {
		return false
	}
	if err != nil {
		s.log.WithError(err).Error("got error trying to fetch block")
		return false
	}

	return true
}
