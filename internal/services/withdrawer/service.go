package withdrawer

import (
	"context"
	"github.com/tokend/erc20-deposit-svc/internal/data"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tokend/erc20-deposit-svc/internal/data/eth"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func (s *Service) Run(ctx context.Context) error {
	token, err := data.NewERC20(s.details.ERC20.Address, s.eth)
	if err != nil {
		return errors.Wrap(err, "failed to init token contract", logan.F{
			"token_address": s.details.ERC20.Address,
		})
	}

	fields := logan.F{}
	entities := s.addressProvider.BindedExternalSystemEntities(ctx, int32(s.systemType))
	for _, entity := range entities {
		address := common.HexToAddress(entity.Data.Address)
		fields["contract_address"] = address.Hex()
		contract, err := s.getContract(address)
		if err != nil {
			return errors.Wrap(err, "failed to get contract")
		}
		ok, err := s.isOwner(contract)
		if err != nil {
			return errors.Wrap(err, "failed to check owner")
		}
		if !ok {
			s.log.WithFields(fields).Warn("not an owner")
			continue
		}

		fields["tokend_addr"] = s.details.ERC20.Address.Hex()
		balance, err := token.BalanceOf(nil, address)
		if err != nil {
			return errors.Wrap(err, "failed to get balance of")
		}
		fields["balance"] = balance.String()
		if balance.Cmp(s.threshold) == -1 {
			s.log.WithFields(fields).Info("lower than threshold")
			continue
		}
		tx, err := contract.WithdrawAllTokens(&bind.TransactOpts{
			From:     s.keyPair.Address(),
			GasPrice: eth.FromGwei(s.gasPrice),
			GasLimit: 200000,
			Signer: func(signer types.Signer, addresses common.Address, transaction *types.Transaction) (*types.Transaction, error) {
				return s.keyPair.SignTX(transaction)
			},
		}, s.hotWallet, s.details.ERC20.Address)
		if err != nil {
			return errors.Wrap(err, "failed to withdraw token")
		}
		fields["tx_hash"] = tx.Hash()

		eth.EnsureHashMined(ctx, s.log.WithFields(fields), s.eth, tx.Hash())

	}
	return nil
}

// return deposit contract instance by address, doing some checks.
// not safe for concurrent use
func (s *Service) getContract(address common.Address) (*data.Contract, error) {
	if s.contracts == nil {
		s.contracts = map[string]data.Contract{}
	}

	if contract, ok := s.contracts[address.Hex()]; ok {
		return &contract, nil
	}

	contract, err := data.NewContract(address, s.eth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init contract")
	}

	s.contracts[address.Hex()] = *contract

	return contract, nil
}

func (s *Service) isOwner(contract *data.Contract) (bool, error) {
	owner, err := contract.Owner(nil)
	if err != nil {
		return false, errors.Wrap(err, "failed to get contract owner")
	}
	return owner.Hex() == s.keyPair.Address().Hex(), nil
}
