package config

import (
	"math/big"
	"reflect"

	"github.com/tokend/erc20-deposit-svc/internal/services/withdrawer/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/keypair/figurekeypair"
)

type FunnelConfig struct {
	GasPrice  *big.Int       `fig:"gas_price,required"`
	Threshold *big.Int       `fig:"threshold,required"`
	HotWallet common.Address `fig:"hot_wallet,required"`
	KeyPair   *eth.Keypair   `fig:"private_key,required"`
}

func (c *config) FunnelConfig() FunnelConfig {
	return c.funnelConfig.Do(func() interface{} {
		var result FunnelConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks, figurekeypair.Hooks, eth.KeypairHook, hooks).
			From(kv.MustGetStringMap(c.getter, "funnel")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out funnel"))
		}

		return result
	}).(FunnelConfig)
}

var hooks = figure.Hooks{
	"[]common.Address": func(raw interface{}) (reflect.Value, error) {
		addressesStrings, err := cast.ToStringSliceE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "Failed to cast provider to map[string]interface{}")
		}

		var addresses []common.Address
		for i, addrStr := range addressesStrings {
			if !common.IsHexAddress(addrStr) {
				// provide value does not look like valid address
				return reflect.Value{}, errors.From(errors.New("invalid address"), logan.F{
					"address_string": addrStr,
					"address_i":      i,
				})
			}

			addresses = append(addresses, common.HexToAddress(addrStr))
		}

		return reflect.ValueOf(addresses), nil
	},
}
