package config

import (
	"math/big"

	"github.com/tokend/erc20-deposit-svc/internal/data/eth"

	"gitlab.com/tokend/keypair/figurekeypair"

	"gitlab.com/tokend/keypair"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type DeployerConfig struct {
	KeyPair       *eth.Keypair   `fig:"private_key,required"`
	GasPrice      *big.Int       `fig:"gas_price,required"`
	GasLimit      *big.Int       `fig:"gas_limit,required"`
	ContractCount int            `fig:"contract_count,required"`
	ContractOwner common.Address `fig:"contract_owner,required"`
	Signer        keypair.Full   `json:"signer,required"`
}

func (c *config) DeployerConfig() DeployerConfig {
	return c.deployerConfig.Do(func() interface{} {
		var result DeployerConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks, figurekeypair.Hooks, hooks).
			From(kv.MustGetStringMap(c.getter, "deployer")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out deployer"))
		}

		return result
	}).(DeployerConfig)
}
