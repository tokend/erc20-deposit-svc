package config

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/tokend/erc20-deposit-svc/internal/data/eth"

	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type FunnelConfig struct {
	GasPrice   *big.Int       `fig:"gas_price,required"`
	Threshold  *big.Int       `fig:"threshold,required"`
	HotWallet  common.Address `fig:"hot_wallet,required"`
	PrivateKey string         `fig:"private_key,required"`
}

func (c *config) FunnelConfig() FunnelConfig {
	return c.funnelConfig.Do(func() interface{} {
		var result FunnelConfig

		err := figure.
			Out(&result).
			With(figure.BaseHooks, eth.KeypairHook, hooks).
			From(kv.MustGetStringMap(c.getter, "funnel")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out funnel"))
		}

		return result
	}).(FunnelConfig)
}

var hooks = figure.Hooks{
	"common.Address": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case string:
			if !common.IsHexAddress(v) {
				// provide value does not look like valid address
				return reflect.Value{}, errors.New("invalid address")
			}
			return reflect.ValueOf(common.HexToAddress(v)), nil
		default:
			return reflect.Value{}, fmt.Errorf("unsupported conversion from %T", value)
		}
	},
	"*eth.Keypair": func(raw interface{}) (reflect.Value, error) {
		switch value := raw.(type) {
		case string:
			kp, err := eth.NewKeypair(value)
			if err != nil {
				return reflect.Value{}, errors.Wrap(err, "failed to init keypair")
			}
			return reflect.ValueOf(kp), nil
		default:
			return reflect.Value{}, fmt.Errorf("cant init keypair from type: %T", value)
		}
	},
}
