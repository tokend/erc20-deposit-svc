package watchlist

import (
	"context"
	"encoding/json"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	regources "gitlab.com/tokend/regources/generated"
	"time"
)

//GetToAdd returns channel assets to watch deposits for will be sent into
func (s *Service) GetToAdd() <-chan Details {
	return s.toAdd
}
//GetToRemove returns channel assets to stop watching deposits for will be sent into
func (s *Service) GetToRemove() <-chan string {
	return s.toRemove
}

//Run runs watchlist service
func (s *Service) Run(ctx context.Context) {
	defer close(s.toAdd)
	defer close(s.toRemove)

	running.WithBackOff(
		ctx,
		s.log,
		"asset-watcher",
		s.processAllAssetsOnce,
		10*time.Second,
		20*time.Second,
		5*time.Minute,
	)
}

func (s *Service) processAllAssetsOnce(ctx context.Context) error {
	active := make(map[string]bool)
	assetsToWatch, err := s.getWatchList()
	if err != nil {
		return errors.Wrap(err, "failed to get asset watch list")
	}
	for _, asset := range assetsToWatch {
		s.toAdd <- asset
		active[asset.ID] = true
	}

	for asset := range s.watchlist {
		if _, ok := active[asset]; !ok {
			s.toRemove <- asset
		}
	}

	s.watchlist = active
	return nil
}

func (s *Service) getWatchList() ([]Details, error) {
	s.streamer.SetFilters(query.AssetFilters{Owner: &s.owner})

	assetsResponse, err := s.streamer.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get asset list for owner", logan.F{
			"owner_address": s.owner,
		})
	}

	watchList, err := s.filter(assetsResponse.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter asset list")
	}

	links := assetsResponse.Links
	for len(assetsResponse.Data) > 0 {
		assetsResponse, err = s.streamer.Next()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get next page of assetsResponse", logan.F{
				"links": links,
			})
		}

		links = assetsResponse.Links
		filtered, err := s.filter(assetsResponse.Data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to filter asset list")
		}
		watchList = append(watchList, filtered...)
	}

	return watchList, nil
}

func (s *Service) filter(assets []regources.Asset) ([]Details, error) {
	result := make([]Details, 0, len(assets))
	for _, asset := range assets {
		details := asset.Attributes.Details
		assetDetails := AssetDetails{}
		_ = json.Unmarshal([]byte(details), &assetDetails)


		if !assetDetails.ERC20.Deposit {
			continue
		}

		if err := assetDetails.Validate(); err != nil {
			s.log.WithError(err).Warn("incorrect asset details")
			continue
		}

		result = append(result, Details{
			Asset:        asset,
			AssetDetails: assetDetails,
		})
	}

	return result, nil
}
