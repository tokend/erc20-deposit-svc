package issuer

import (
	"context"
	"github.com/tokend/erc20-deposit-svc/internal/services/transfer"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
	"gitlab.com/tokend/keypair"
	regources "gitlab.com/tokend/regources/generated"
	"time"
)

type txSubmitter interface {
	Submit(ctx context.Context, envelope string, waitIngest bool) (*regources.TransactionResponse, error)
}

type transactionStreamer interface {
	StreamTransactions(ctx context.Context, changeTypes, entryTypes []int,
	) (<-chan regources.TransactionListResponse, <-chan error)
}

// addressProvider must be implemented by WatchAddress storage to pass into Service constructor.
type addressProvider interface {
	ExternalAccountAt(ctx context.Context, ts time.Time, externalSystem int32, externalAddress string, payload *string) (address *string)
	Balance(ctx context.Context, address string, asset string) (balance *string)
}

type Service struct {
	streamer    transactionStreamer
	txSubmitter txSubmitter
	builder     xdrbuild.Builder
	asset       watchlist.Details
	log         *logan.Entry

	owner           keypair.Address
	issuer          keypair.Full
	addressProvider addressProvider
	ch              <-chan transfer.Details
}

type Opts struct {
	Streamer     transactionStreamer
	TxSubmitter  txSubmitter
	Builder      xdrbuild.Builder
	AssetDetails watchlist.Details
	Signer       keypair.Full
	Log          *logan.Entry
	Chan         <-chan transfer.Details
}

func New(opts Opts) *Service {

	return &Service{
		asset:       opts.AssetDetails,
		issuer:      opts.Signer,
		streamer:    opts.Streamer,
		builder:     opts.Builder,
		txSubmitter: opts.TxSubmitter,
		log: opts.Log.WithFields(logan.F{
			"asset_code": opts.AssetDetails.ID,
		}),
		owner: keypair.MustParseAddress(opts.AssetDetails.Relationships.Owner.Data.ID),
		ch: opts.Chan,
	}
}
