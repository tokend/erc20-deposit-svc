[![Build Status](https://travis-ci.org/tokend/erc20-deposit-svc.svg?branch=master)](https://travis-ci.org/tokend/erc20-deposit-svc)

# ERC20 Deposit service
ERC20 deposit service is a bridge between TokenD and Ethereum blockchain which allows 
to deposit tokens into TokenD from Ethereum blockchain. It listens for token transfers
specific addresses.

## Usage

Environmental variable `KV_VIPER_FILE` must be set and contain path to desired config file.

```bash
erc20-deposit-svc run withdraw
```

## Watchlist

In order for service to start watching deposits in specific asset, asset details in TokenD must have entry of the following form: 
```json5
{
//...
"erc20": {
   "deposit": true, 
   "address": "0x0000000000000000000",  //contract address
   },
//...
}
```

## Config

```yaml
rpc:
  endpoint: "ws://ETH_NODE_ADDRESS"

ethereum:
  checkpoint: #block to start listening for deposits from
  confirmations: 20 #number of confirmations to wait for

horizon:
  endpoint: "SOME_VALID_ADDRESS"
  signer: "S_ASSET_OWNER_SECRET_KEY" # only for requests

deposit:
  admin_signer: "S_ASSET_OWNER_SIGNER_SECRET_KEY" # used for signing transactions

log:
  level: debug
  disable_sentry: true

```

Just add public key of `deposit: admin_signer` as signer to corporate account for issuance

## Ethereum node

Node must be configured to accept connections through websockets. 
Origin must be explicitly or implicitly whitelisted:
either `--wsorigins "some_origin"`, or `--wsorigins *` to accept all connections.
