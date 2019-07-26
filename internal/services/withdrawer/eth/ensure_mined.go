package eth

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
)

type ByHasher interface {
	TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error)
}

func EnsureHashMined(ctx context.Context, log *logan.Entry, getter ByHasher, hash common.Hash) {
	logger := log.WithField("tx_hash", hash.String())

	running.UntilSuccess(ctx, logger, "ensure-hash-mined", func(ctx context.Context) (bool, error) {
		tx, isPending, err := getter.TransactionByHash(ctx, hash)
		if err != nil {
			return false, errors.Wrap(err, "failed to get tx")
		}
		if tx == nil {
			return false, errors.New("tx not found")
		}
		if isPending {
			return false, nil
		}

		return true, nil
	}, 5*time.Second, 30*time.Second)
}
