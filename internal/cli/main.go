package cli

import (
	"context"
	"fmt"

	"github.com/tokend/erc20-deposit-svc/internal/services/funnel"

	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/services/depositer"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gopkg.in/alecthomas/kingpin.v2"
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
	funnelService := runCmd.Command("funnel", "run funnel service")
	versionCmd := app.Command("version", "service revision")

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
	case funnelService.FullCommand():
		svc := funnel.New(cfg)
		err := svc.Run(context.Background())
		if err != nil {
			log.WithError(err).Error("failed to run funnel")
			return false
		}
	case versionCmd.FullCommand():
		fmt.Println(config.ERC20DepositVersion)
	default:
		log.WithField("command", cmd).Error("Unknown command")
		return false
	}

	return true
}
