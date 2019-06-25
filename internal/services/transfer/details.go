package transfer

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

//Details contain transfer details to send to issuer 
type Details struct {
	TransactionHash string
	Destination     common.Address
	BlockTime time.Time
	Amount          *big.Int
	Decimals        int64
}
