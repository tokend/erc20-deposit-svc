package verifier

import (
	"context"
	"encoding/json"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdr"
	"gitlab.com/tokend/go/xdrbuild"
	regources "gitlab.com/tokend/regources/generated"
	"strconv"
)

func (s *Service) approveRequest(
	ctx context.Context,
	request regources.ReviewableRequest,
	toAdd, toRemove uint32,
	extDetails map[string]interface{}) error {
	id, err := strconv.ParseUint(request.ID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse request id")
	}
	bb, err := json.Marshal(extDetails)
	if err != nil {
		return errors.Wrap(err, "failed to marshal external bb map")
	}
	envelope, err := s.builder.Transaction(s.depositCfg.AssetOwner).Op(xdrbuild.ReviewRequest{
		ID:     id,
		Hash:   &request.Attributes.Hash,
		Action: xdr.ReviewRequestOpActionApprove,
		ReviewDetails: xdrbuild.ReviewDetails{
			TasksToAdd:      toAdd,
			TasksToRemove:   toRemove,
			ExternalDetails: string(bb),
		},
		Details: xdrbuild.IssuanceDetails{},
	}).Sign(s.depositCfg.AssetIssuer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to prepareVerify transaction envelope")
	}
	_, err = s.txSubmitter.Submit(ctx, envelope, true)
	if err != nil {
		return errors.Wrap(err, "failed to approve withdraw request")
	}

	return nil
}

func (s *Service) permanentReject(
	ctx context.Context,
	request regources.ReviewableRequest, reason string) error {
	id, err := strconv.ParseUint(request.ID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse request id")
	}
	envelope, err := s.builder.Transaction(s.depositCfg.AssetOwner).Op(xdrbuild.ReviewRequest{
		ID:     id,
		Hash:   &request.Attributes.Hash,
		Action: xdr.ReviewRequestOpActionPermanentReject,
		ReviewDetails: xdrbuild.ReviewDetails{
			ExternalDetails: "{}",
		},
		Details: xdrbuild.IssuanceDetails{},
	}).Sign(s.depositCfg.AssetIssuer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to prepare transaction envelope")
	}
	_, err = s.txSubmitter.Submit(ctx, envelope, true)
	if err != nil {
		return errors.Wrap(err, "failed to approve permanently reject request")
	}

	return nil
}
