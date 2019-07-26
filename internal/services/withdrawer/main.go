package withdrawer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
	"github.com/tokend/erc20-deposit-svc/internal/services/withdrawer/eth"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/addrstate"
)

type Opts struct {
	Client          *ethclient.Client
	Config          config.Config
	Asset           watchlist.Details
	SystemType      uint32
	AddressProvider *addrstate.Watcher
}

type Service struct {
	details         watchlist.Details
	contract        *Contract
	log             *logan.Entry
	addressProvider *addrstate.Watcher
	eth             *ethclient.Client
	keyPair         *eth.Keypair
	hotWallet       common.Address
	gasPrice        *big.Int
	threshold       *big.Int
	systemType      uint32
}

func NewWithdrawer(opts Opts) *Service {
	return &Service{
		details:         opts.Asset,
		log:             opts.Config.Log(),
		addressProvider: opts.AddressProvider,
		eth:             opts.Client,
		keyPair:         opts.Config.FunnelConfig().KeyPair,
		hotWallet:       opts.Config.FunnelConfig().HotWallet,
		gasPrice:        opts.Config.FunnelConfig().GasPrice,
		threshold:       opts.Config.FunnelConfig().Threshold,
		systemType:      opts.SystemType,
	}
}
