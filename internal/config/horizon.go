package config

import (
	"github.com/tokend/erc20-deposit-svc/internal/horizon/client"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/keypair"
	"gitlab.com/tokend/keypair/figurekeypair"
	"net/http"
	"net/url"
)

type Horizoner interface {
	Horizon() *client.Client
}

type horizoner struct {
	getter kv.Getter
	once   comfig.Once
	value  *client.Client
}

func NewHorizoner(getter kv.Getter) Horizoner {
	return &horizoner{getter: getter}
}

func (h *horizoner) Horizon() *client.Client {
	h.once.Do(func() interface{} {
		var config struct {
			Endpoint *url.URL     `fig:"endpoint,required"`
			Signer   keypair.Full `fig:"signer,required"`
		}

		err := figure.
			Out(&config).
			With(figure.BaseHooks, figurekeypair.Hooks).
			From(kv.MustGetStringMap(h.getter, "horizon")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out horizon"))
		}

		hrz := client.New(http.DefaultClient, config.Endpoint)
		if config.Signer != nil {
			hrz = hrz.WithSigner(config.Signer)
		}

		h.value = hrz
		return nil
	})

	return h.value
}
