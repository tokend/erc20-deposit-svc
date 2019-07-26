package funnel

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tokend/erc20-deposit-svc/internal/services/withdrawer/eth"

	"gitlab.com/distributed_lab/logan/v3/errors"

	"gitlab.com/distributed_lab/running"

	withdrawer2 "github.com/tokend/erc20-deposit-svc/internal/services/withdrawer"

	"github.com/tokend/erc20-deposit-svc/internal/transaction"
	regources "gitlab.com/tokend/regources/generated"

	"gitlab.com/tokend/addrstate"

	"github.com/ethereum/go-ethereum/ethclient"

	"gitlab.com/distributed_lab/logan/v3"

	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/getters"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
)

type transactionStreamer interface {
	StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
	) (<-chan regources.TransactionListResponse, <-chan error)
}

const externalSystemTypeEthereumKey = "external_system_type:ethereum"

type Service struct {
	config          config.Config
	assetWatcher    *watchlist.Service
	log             *logan.Entry
	addressProvider *addrstate.Watcher
	streamer        transactionStreamer

	eth     *ethclient.Client
	keyPair *eth.Keypair

	spawned        sync.Map
	assetsToAdd    <-chan watchlist.Details
	assetsToRemove <-chan string
	sync.WaitGroup

	externalSystemType uint32
}

func New(cfg config.Config) *Service {
	assetWatcher := watchlist.New(watchlist.Opts{
		Streamer: getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:      cfg.Log(),
	})

	return &Service{
		config:         cfg,
		assetWatcher:   assetWatcher,
		log:            cfg.Log(),
		streamer:       transaction.NewStreamer(getters.NewDefaultTransactionHandler(cfg.Horizon())),
		eth:            cfg.EthClient(),
		assetsToAdd:    assetWatcher.GetToAdd(),
		assetsToRemove: assetWatcher.GetToRemove(),
	}
}

func (s *Service) Run(ctx context.Context) error {
	mutators := []addrstate.StateMutator{
		addrstate.ExternalSystemBindingMutator{SystemType: int32(s.externalSystemType)},
	}

	keypair, err := eth.NewKeypair(s.config.FunnelConfig().PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to init keypair")
	}
	s.keyPair = keypair

	systemType, err := s.getSystemType(externalSystemTypeEthereumKey)
	if err != nil {
		return errors.Wrap(err, "failed to get external system type")
	}
	if systemType == nil {
		return errors.New("no key value for external system type")
	}

	s.externalSystemType = *systemType
	addrProvider := addrstate.New(
		ctx,
		s.log,
		mutators,
		s.streamer,
	)
	s.addressProvider = addrProvider

	go s.assetWatcher.Run(ctx)

	s.Add(2)
	go s.spawner(ctx)
	go s.cancellor(ctx)
	s.Wait()

	return nil
}

func (s *Service) spawner(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToAdd {
		if _, ok := s.spawned.Load(asset.ID); !ok {
			s.spawn(ctx, asset)
		}
	}
}

func (s *Service) cancellor(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToRemove {
		if raw, ok := s.spawned.Load(asset); ok {
			cancelFunc := raw.(context.CancelFunc)
			cancelFunc()
			s.spawned.Delete(asset)
		}
	}
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {
	localCtx, cancelFunc := context.WithCancel(ctx)
	s.spawned.Store(details.Asset.ID, cancelFunc)

	withdrawer := withdrawer2.NewWithdrawer(
		withdrawer2.Opts{
			Client:          s.eth,
			Config:          s.config,
			Asset:           details,
			SystemType:      s.externalSystemType,
			AddressProvider: s.addressProvider,
			KeyPair:         s.keyPair,
		})

	s.log.WithField("asset", details.ID).Info("Started listening for deposits")
	running.WithBackOff(localCtx, s.log, "withdrawer-service", withdrawer.Run, time.Second, 10*time.Second, 5*time.Minute)
}

func (s *Service) getSystemType(key string) (*uint32, error) {
	body, err := s.config.Horizon().Get(fmt.Sprintf("/v3/key_values/%s", key))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key value")
	}
	var response regources.KeyValueEntryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal response")
	}
	return response.Data.Attributes.Value.U32, nil
}
