package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/keypair/figurekeypair"
)

type EthereumConfig struct {
	Checkpoint uint64 `fig:"checkpoint"`
	Confirmations int64 `fig:"confirmations"`
}

func (c *config) EthereumConfig() EthereumConfig {

	c.once.Do(func() interface{} {
		var result EthereumConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks, figurekeypair.Hooks).
			From(kv.MustGetStringMap(c.getter, "ethereum")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out deposit"))
		}

		c.ethereumConfig = result
		return nil
	})

	return c.ethereumConfig
}
