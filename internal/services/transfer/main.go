package transfer

import (
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tokend/erc20-deposit-svc/internal/config"
	"github.com/tokend/erc20-deposit-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
)

const erc20ABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]"

//Service respresents transfer listener service
type Service struct {
	log      *logan.Entry
	contract *bind.BoundContract
	client   *ethclient.Client
	cfg      config.EthereumConfig

	asset watchlist.Details

	ch chan Details

	sync.RWMutex
	decimals int64
}

func (s *Service) GetCh() <-chan Details {
	return s.ch
}

//Opts contain parameters to create service with
type Opts struct {
	AssetDetails watchlist.Details
	Log          *logan.Entry
	Client       ethclient.Client
	Config       config.EthereumConfig
}

//New creates new transfer listener service
func New(opts Opts) *Service {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		opts.Log.WithError(err).Fatal("failed to parse contract ABI")
	}

	contract := bind.NewBoundContract(
		opts.AssetDetails.ERC20.Address,
		parsed,
		&opts.Client,
		&opts.Client,
		&opts.Client,
	)

	decimals := new(uint8)
	err = contract.Call(&bind.CallOpts{}, decimals, "decimals")
	if err != nil {
		opts.Log.WithError(err).Error("failed to get decimals of erc20 contract")
		return nil
	}

	return &Service{
		log:      opts.Log,
		asset:    opts.AssetDetails,
		contract: contract,
		cfg:      opts.Config,
		ch:       make(chan Details, 10),
		client:   &opts.Client,
		RWMutex:  sync.RWMutex{},
		decimals: int64(*decimals),
	}
}
