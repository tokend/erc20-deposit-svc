package depositer

import (
	"context"
	"sync"

	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/submit"
	"github.com/tokend/erc20-deposit-svc/internal/services/issuer"
	"github.com/tokend/erc20-deposit-svc/internal/services/verifier"
	"github.com/tokend/erc20-deposit-svc/internal/transaction"
	"gitlab.com/tokend/keypair"

	"github.com/tokend/erc20-deposit-svc/internal/horizon"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/getters"
	"github.com/tokend/erc20-deposit-svc/internal/services/transfer"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
)

type Service struct {
	assetWatcher   *watchlist.Service
	log            *logan.Entry
	config         config.Config
	builder        xdrbuild.Builder
	adminID        keypair.Address
	spawned        sync.Map
	assetsToAdd    <-chan watchlist.Details
	assetsToRemove <-chan string
	sync.WaitGroup
}

//New creates new depositer service that gathers asset watcher, issuer and transfer listener
func New(cfg config.Config) *Service {
	assetWatcher := watchlist.New(watchlist.Opts{
		Streamer: getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:      cfg.Log(),
	})

	connector := horizon.NewConnector(cfg.Horizon())
	builder, err := connector.Builder()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to make builder")
	}

	horizonInfo, err := connector.State()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to get state from connector")
	}

	return &Service{
		log:     cfg.Log(),
		config:  cfg,
		builder: *builder,
		adminID: keypair.MustParseAddress(horizonInfo.Data.Attributes.MasterAccountId),

		assetWatcher:   assetWatcher,
		assetsToAdd:    assetWatcher.GetToAdd(),
		assetsToRemove: assetWatcher.GetToRemove(),
		spawned:        sync.Map{},
		WaitGroup:      sync.WaitGroup{},
	}
}

//Run starts depositer service
func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	s.Add(2)
	go s.spawner(ctx)
	go s.cancellor(ctx)
	s.Wait()
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
	fields := logan.F{"asset": details.ID}

	transferStreamer := transfer.New(transfer.Opts{
		Client:       *s.config.EthClient(),
		Log:          s.log,
		AssetDetails: details,
		Config:       s.config.EthereumConfig(),
	})
	if transferStreamer == nil {
		s.log.WithFields(fields).Warn("transfer streamer is nil, skipping this asset")
		return
	}

	issueSubmitter := issuer.New(issuer.Opts{
		AssetDetails: details,
		Log:          s.log,
		Streamer: transaction.NewStreamer(
			getters.NewDefaultTransactionHandler(s.config.Horizon()),
			s.log,
		),
		Builder:     s.builder,
		Signer:      s.config.DepositConfig().AdminSigner,
		TxSubmitter: submit.New(s.config.Horizon()),
		Chan:        transferStreamer.GetCh(),
	})
	verifierService := verifier.New(verifier.Opts{
		Builder:   s.builder,
		Log:       s.log,
		Config:    s.config,
		Submitter: submit.New(s.config.Horizon()),
		Client:    *s.config.EthClient(),
		Asset:     details,
		AdminID:   s.adminID,
		Streamer:  getters.NewDefaultCreateIssuanceRequestHandler(s.config.Horizon()),
	})

	localCtx, cancelFunc := context.WithCancel(ctx)
	s.spawned.Store(details.Asset.ID, cancelFunc)

	go transferStreamer.Run(localCtx)
	go issueSubmitter.Run(localCtx)
	go verifierService.Run(localCtx)

	s.log.WithFields(fields).Info("Started listening for deposits")
}
