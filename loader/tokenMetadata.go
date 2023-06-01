package loader

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (c *OnChainLoader) FetchTokenMetadata(chain types.ChainId, token types.EthAddress) (types.TokenMetadata, error) {
	// Special case handling for Native ETH
	if token == "0x0000000000000000000000000000000000000000" {
		return types.TokenMetadata{
			Symbol:   "ETH",
			Decimals: 18,
		}, nil
	}

	cfg, okay := c.Cfg.ChainConfig(chain)
	var metadata types.TokenMetadata

	if !okay {
		log.Println("Warning no chain configuration for " + chain)
		return metadata, fmt.Errorf("Chain configuration missing")
	}

	rpcUrl := cfg.RPCEndpoint()
	client, err := ethclient.DialContext(context.Background(), rpcUrl)

	if err != nil {
		log.Println("Warning RPC connection error: " + err.Error())
		return metadata, err
	}

	metadata.Symbol, err = tokenSymbolQuery(client, token)
	if err != nil {
		return metadata, err
	}

	metadata.Decimals, err = tokenDecimalQuery(client, token)
	return metadata, err
}

func tokenSymbolQuery(client *ethclient.Client, token types.EthAddress) (string, error) {
	result, err := callContractKey("symbol", token, client, tokenABI())
	if err != nil {
		return "", nil
	}
	return result.(string), nil
}

func tokenDecimalQuery(client *ethclient.Client, token types.EthAddress) (int, error) {
	result, err := callContractKey("decimals", token, client, tokenABI())
	if err != nil {
		return 0, nil
	}
	return int(result.(uint64)), nil
}

const erc20ABI = `[
	{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},
	{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}
]`

func tokenABI() abi.ABI {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}
	return parsedABI
}
