package eth

import (
	"encoding/hex"

	"crypto/ecdsa"

	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	ErrNoKey = errors.New("wallet doesn't have requested key")
)

type Wallet struct {
	hd     bool
	master *ETHDeriver
	keys   map[common.Address]ecdsa.PrivateKey
}

func NewHDWallet(hdprivate string, n uint64) (*Wallet, error) {
	master, err := NewETHDeriver(hdprivate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init key deriver")
	}

	wallet := &Wallet{
		hd:     true,
		master: master,
		keys:   make(map[common.Address]ecdsa.PrivateKey),
	}

	// TODO check horizon for account sequence and extended keys as needed
	if err := wallet.extend(n); err != nil {
		return nil, errors.Wrap(err, "failed to extend master")
	}

	return wallet, nil
}

func NewWallet() *Wallet {
	return &Wallet{
		keys: make(map[common.Address]ecdsa.PrivateKey),
	}
}

func (wallet *Wallet) ImportHEX(data string) (common.Address, error) {
	raw, err := hex.DecodeString(data)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "failed to decode string")
	}
	return wallet.Import(raw)
}

func (wallet *Wallet) Import(raw []byte) (common.Address, error) {
	pk, err := crypto.ToECDSA(raw)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "failed to convert pk")
	}
	address := crypto.PubkeyToAddress(pk.PublicKey)
	wallet.keys[address] = *pk
	return address, nil
}

func (wallet *Wallet) extend(i uint64) error {
	for uint64(len(wallet.keys)) < i {
		child, err := wallet.master.ChildPrivate(uint32(len(wallet.keys)))
		if err != nil {
			return errors.Wrap(err, "failed to extend child")
		}

		raw, err := hex.DecodeString(child)
		if err != nil {
			return errors.Wrap(err, "failed to decode private key")
		}

		if _, err := wallet.Import(raw); err != nil {
			return errors.Wrap(err, "failed to import key")
		}
	}

	return nil
}

func (wallet *Wallet) Addresses(ctx context.Context) (result []common.Address) {
	for addr := range wallet.keys {
		result = append(result, addr)
	}
	return result
}

func (wallet *Wallet) SignTX(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	key, ok := wallet.keys[address]
	if !ok {
		return nil, ErrNoKey
	}
	return SignTXWithPrivate(&key, tx)
}

func SignTXWithPrivate(key *ecdsa.PrivateKey, tx *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(tx, types.HomesteadSigner{}, key)
}

func (wallet *Wallet) HasAddress(address common.Address) bool {
	_, ok := wallet.keys[address]
	return ok
}
