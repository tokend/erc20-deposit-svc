package cli

import (
	"context"
	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/services/depositer"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)
//Run runs service
func Run(args []string) bool {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	app := kingpin.New("erc20-deposit-svc", "")
	runCmd := app.Command("run", "run command")
	deposit := runCmd.Command("deposit", "run deposit service")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
	}

	cfg := config.NewConfig(kv.MustFromEnv())
	log = cfg.Log()


	switch cmd {
	case deposit.FullCommand():
		svc := depositer.New(cfg)
		svc.Run(context.Background())
	default:
		log.WithField("command", cmd).Error("Unknown command")
		return false
	}

	return true
}
