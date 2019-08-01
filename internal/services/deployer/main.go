package deployer

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/horizon"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/client"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/submit"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
	regources "gitlab.com/tokend/regources/generated"
)

type txSubmitter interface {
	Submit(ctx context.Context, envelope string, waitIngest bool) (*regources.TransactionResponse, error)
}

type Service struct {
	log       *logan.Entry
	eth       *ethclient.Client
	config    config.Config
	builder   *xdrbuild.Builder
	horizon   *client.Client
	submitter txSubmitter
}

func New(cfg config.Config) *Service {
	builder, err := horizon.NewConnector(cfg.Horizon()).Builder()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to make builder")
	}
	return &Service{
		log:       cfg.Log(),
		eth:       cfg.EthClient(),
		config:    cfg,
		builder:   builder,
		horizon:   cfg.Horizon(),
		submitter: submit.New(cfg.Horizon()),
	}
}
