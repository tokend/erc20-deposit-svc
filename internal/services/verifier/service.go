package verifier

import (
	"context"
	"fmt"
	"time"

	"github.com/tokend/erc20-deposit-svc/internal/horizon/page"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	regources "gitlab.com/tokend/regources/generated"
)

func (s *Service) Run(ctx context.Context) {
	s.prepare()
	issuancePage := &regources.ReviewableRequestListResponse{}
	var err error
	running.WithBackOff(ctx, s.log, "verifier", func(ctx context.Context) error {
		if len(issuancePage.Data) < requestPageSizeLimit {
			issuancePage, err = s.issuances.List()
		} else {
			issuancePage, err = s.issuances.Next()
		}
		if err != nil {
			return errors.Wrap(err, "error occurred while issuance request page fetching")
		}
		for _, data := range issuancePage.Data {
			details := issuancePage.Included.MustCreateIssuanceRequest(data.Relationships.RequestDetails.Data.GetKey())
			err := s.confirmIssuanceSuccessful(ctx, data, details)
			if err != nil {
				s.log.
					WithError(err).
					WithField("details", details).
					Warn("failed to process issuance request")
				continue
			}
		}
		return nil
	}, 15*time.Second, 15*time.Second, time.Hour)
}

func (s *Service) prepare() {
	state := reviewableRequestStatePending
	reviewer := s.adminID.Address()
	pendingTasks := fmt.Sprintf("%d", taskCheckTxConfirmed)
	filters := query.CreateIssuanceRequestFilters{
		Asset: &s.asset.ID,
		ReviewableRequestFilters: query.ReviewableRequestFilters{
			State:        &state,
			Reviewer:     &reviewer,
			PendingTasks: &pendingTasks,
		},
	}

	s.issuances.SetFilters(filters)
	s.issuances.SetIncludes(query.CreateIssuanceRequestIncludes{
		ReviewableRequestIncludes: query.ReviewableRequestIncludes{
			RequestDetails: true,
		},
	})
	limit := fmt.Sprintf("%d", requestPageSizeLimit)
	s.issuances.SetPageParams(page.Params{Limit: &limit})
}
