package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type config struct {
	depositConfig  DepositConfig
	ethereumConfig EthereumConfig

	comfig.Logger
	getter kv.Getter
	once   comfig.Once
	Horizoner
	Ether
}

type Config interface {
	DepositConfig() DepositConfig
	EthereumConfig() EthereumConfig

	comfig.Logger
	Horizoner
	Ether
}

func NewConfig(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Horizoner: NewHorizoner(getter),
		Logger:    comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Ether:     NewEther(getter),
	}
}
