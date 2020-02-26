package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tokend/erc20-deposit-svc/internal/horizon/submit"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"hash/crc64"
	"math/big"
	"strings"

	regources "gitlab.com/tokend/regources/generated"

	"github.com/spf13/cast"
	"gitlab.com/tokend/go/xdrbuild"

	"github.com/tokend/erc20-deposit-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/tokend/erc20-deposit-svc/internal/data/eth"
)

const externalSystemTypeEthereumKey = "external_system_type:ethereum"

func (s *Service) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if rvr := recover(); rvr != nil {
			// we are spending actual ether here,
			// so in case of emergency abandon the operations completely
			cancel()
			err = errors.Wrap(errors.FromPanic(rvr), "service panicked")
		}
	}()

	systemType, err := s.getSystemType(externalSystemTypeEthereumKey)
	if err != nil {
		return errors.Wrap(err, "failed to get external system type")
	}
	if systemType == nil {
		return errors.New("no key value for external system type")
	}
	deployedEntries, err := s.getExternalSystemPoolEntityCount(*systemType)
	if err != nil {
		return errors.Wrap(err, "unable to get deployed entries count")
	}

	for i := int(deployedEntries); i < s.config.DeployerConfig().ContractCount; i++ {
		nonce, err := s.eth.PendingNonceAt(ctx, s.config.DeployerConfig().KeyPair.Address())
		if err != nil {
			return errors.Wrap(err, "failed to retrieve account nonce")
		}
		contractAddress := crypto.CreateAddress(s.config.DeployerConfig().KeyPair.Address(), nonce)

		if err := s.createPoolEntities(contractAddress.Hex(), *systemType); err != nil {
			return errors.Wrap(err, "failed to create pool entry")
		}

		contract, err := s.deployContract(big.NewInt(0).SetUint64(nonce))
		if err != nil {
			return errors.Wrap(err, "failed to deploy contract")
		}

		fields := logan.F{}
		fields["contract"] = contract.Hex()
		s.log.WithFields(fields).Info("contract deployed")

		if contract.Hex() != contractAddress.Hex() {
			fields["expected_contract"] = contractAddress
			return errors.From(errors.New("contract address mismatch"), fields)
		}
	}

	return nil
}

func (s *Service) deployContract(nonce *big.Int) (*common.Address, error) {
	_, tx, _, err := data.DeployContract(&bind.TransactOpts{
		From:  s.config.DeployerConfig().KeyPair.Address(),
		Nonce: nonce,
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
			Address: strings.ToLower(address),
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

	_, err = s.submitter.Submit(context.TODO(), envelope, true)
	if err != nil {
		fields := make(logan.F, 1)
		if fail, ok := err.(submit.TxFailure); ok {
			fields["tx"] = fail
		}
		return errors.Wrap(err, "failed to submit tx", fields)
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

// TODO: Use newer endpoint
func (s *Service) getExternalSystemPoolEntityCount(systemType uint32) (uint64, error) {
	rawStats, err := s.horizon.Get("/statistics")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get system stats")
	}
	var stats systemStatistics
	if err := json.Unmarshal(rawStats, &stats); err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal system stats")
	}

	count := stats.ExternalSystemPoolEntriesCount[fmt.Sprintf("%d", systemType)]
	return count, nil
}

type systemStatistics struct {
	// ExternalSystemPoolEntriesCount shows number of active entries per external system type
	ExternalSystemPoolEntriesCount map[string]uint64 `json:"external_system_pool_entries_count,omitempty"`
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
