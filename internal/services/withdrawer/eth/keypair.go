package eth

import (
	"context"

	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// Keypair helper struct to simplify Wallet usage where only one signer is needed
type Keypair struct {
	wallet *Wallet
}

func NewKeypair(hex string) (*Keypair, error) {
	hex = strings.TrimPrefix(hex, "0x")

	wallet := NewWallet()
	_, err := wallet.ImportHEX(hex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to import private key")
	}

	return &Keypair{
		wallet: wallet,
	}, nil
}

func (kp *Keypair) Address() common.Address {
	return kp.wallet.Addresses(context.Background())[0]
}

func (kp *Keypair) SignTX(tx *types.Transaction) (*types.Transaction, error) {
	return kp.wallet.SignTX(kp.Address(), tx)
}
