package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"math/big"
	"time"

	regources "gitlab.com/tokend/regources/generated"

	"github.com/spf13/cast"
	"gitlab.com/tokend/go/xdrbuild"

	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"

	"github.com/tokend/erc20-deposit-svc/internal/data"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/tokend/erc20-deposit-svc/internal/data/eth"
)

const externalSystemTypeEthereumKey = "external_system_type:ethereum"

func (s *Service) Run(ctx context.Context) error {
	fields := logan.F{}
	systemType, err := s.getSystemType(externalSystemTypeEthereumKey)
	if err != nil {
		return errors.Wrap(err, "failed to get external system type")
	}
	if systemType == nil {
		return errors.New("no key value for external system type")
	}
	for i := 0; i < s.config.DeployerConfig().ContractCount; i++ {
		contract, err := s.deployContract()
		if err != nil {
			return errors.Wrap(err, "failed to deploy contract")
		}
		fields["contract"] = contract.Hex()
		s.log.WithFields(fields).Info("contract deployed")

		// critical section. contract has been deployed, we need to create entity at any cost
		running.UntilSuccess(context.Background(), s.log, "create-pool-entity", func(i context.Context) (bool, error) {
			if err := s.createPoolEntities(contract.Hex(), *systemType); err != nil {
				return false, err
			}
			return true, nil
		}, 1*time.Second, 2*time.Second)
	}
	return nil
}

func (s *Service) deployContract() (*common.Address, error) {
	_, tx, _, err := data.DeployContract(&bind.TransactOpts{
		From:  s.config.DeployerConfig().KeyPair.Address(),
		Nonce: nil,
		Signer: func(signer types.Signer, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return s.config.DeployerConfig().KeyPair.SignTX(tx)
		},
		Value:    big.NewInt(0),
		GasPrice: eth.FromGwei(s.config.DeployerConfig().GasPrice),
		GasLimit: s.config.DeployerConfig().GasLimit.Uint64(),
		Context:  context.TODO(),
	}, s.eth, s.config.DeployerConfig().ContractOwner)

	if err != nil {
		return nil, errors.Wrap(err, "failed to submit contract tx")
	}

	eth.EnsureHashMined(context.Background(), s.log.WithField("tx_hash", tx.Hash().Hex()), s.eth, tx.Hash())

	receipt, err := s.eth.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tx receipt")
	}

	// TODO check transaction state/status to see if contract actually was deployed
	// TODO panic if we are not sure if contract is valid

	return &receipt.ContractAddress, nil
}

func (s *Service) createPoolEntities(address string, systemType uint32) error {
	deployerID := Hash64(s.config.DeployerConfig().KeyPair.Address().Bytes())
	data := EthereumAddress{
		Type: "address",
		Data: EthereumAddressData{
			Address: address,
		},
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data")
	}

	envelope, err := s.builder.Transaction(s.config.DeployerConfig().Signer).
		Op(xdrbuild.CreateExternalPoolEntry(cast.ToInt32(systemType), cast.ToString(dataBytes), deployerID)).
		Sign(s.config.DeployerConfig().Signer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal tx")
	}

	result, err := s.submitter.Submit(context.TODO(), envelope, true)
	if err != nil {
		body, _ := json.Marshal(result)
		s.log.Error(string(body))
		return errors.Wrap(err, "failed to submit tx")
	}

	return nil
}

func (s *Service) getSystemType(key string) (*uint32, error) {
	body, err := s.config.Horizon().Get(fmt.Sprintf("/v3/key_values/%s", key))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key value")
	}
	var response regources.KeyValueEntryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal response")
	}
	return response.Data.Attributes.Value.U32, nil
}

func Hash64(msg []byte) uint64 {
	table := crc64.MakeTable(crc64.ISO)
	return crc64.Checksum(msg, table)
}

type EthereumAddress struct {
	Type string              `json:"type"`
	Data EthereumAddressData `json:"data"`
}

type EthereumAddressData struct {
	Address string `json:"address"`
}
