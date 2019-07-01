package verifier

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/getters"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/submit"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
)

type Opts struct {
	Client ethclient.Client

	Submitter submit.Interface
	Builder   xdrbuild.Builder
	Log       *logan.Entry
	Streamer  getters.CreateIssuanceRequestHandler
	Config    config.Config
	Asset     watchlist.Details
}

type Service struct {
	depositCfg config.DepositConfig
	ethCfg     config.EthereumConfig
	asset      watchlist.Details

	builder     xdrbuild.Builder
	issuances   getters.CreateIssuanceRequestHandler
	txSubmitter submit.Interface
	log         *logan.Entry

	client *ethclient.Client
}

func New(opts Opts) *Service {

	return &Service{
		client:      &opts.Client,
		log:         opts.Log,
		depositCfg:  opts.Config.DepositConfig(),
		ethCfg:      opts.Config.EthereumConfig(),
		txSubmitter: opts.Submitter,
		builder:     opts.Builder,
		asset:       opts.Asset,

		issuances: opts.Streamer,
	}
}
