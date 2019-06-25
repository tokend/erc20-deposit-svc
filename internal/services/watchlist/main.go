package watchlist

import (
	"github.com/tokend/erc20-deposit-svc/internal/horizon/getters"
	"gitlab.com/distributed_lab/logan/v3"
)

//Service is struct representing watchlist service
type Service struct {
	streamer  getters.AssetHandler
	log       *logan.Entry
	owner     string
	watchlist map[string]bool
	toAdd     chan Details
	toRemove  chan string
}

//Opts contain parameters required to build service
type Opts struct {
	Streamer   getters.AssetHandler
	Log        *logan.Entry
	AssetOwner string
}

//New creates new watchlist service
func New(opts Opts) *Service {
	toAdd := make(chan Details)
	toRemove := make(chan string)
	return &Service{
		streamer:  opts.Streamer,
		owner:     opts.AssetOwner,
		log:       opts.Log.WithField("service", "watchlist"),
		watchlist: make(map[string]bool),
		toRemove:  toRemove,
		toAdd:     toAdd,
	}
}
