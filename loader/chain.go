package loader

import (
	"context"
	"fmt"
	"log"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type OnChainLoader struct {
	Cfg NetworkConfig
}

func (c *OnChainLoader) ethClientForChain(chainId types.ChainId) (*ethclient.Client, error) {
	cfg, okay := c.Cfg.ChainConfig(chainId)

	if !okay {
		log.Println("Warning no chain configuration for " + chainId)
		return nil, fmt.Errorf("Chain configuration missing")
	}

	rpcUrl := cfg.RPCEndpoint()
	client, err := ethclient.DialContext(context.Background(), rpcUrl)

	if err != nil {
		log.Println("Warning RPC connection error: " + err.Error())
	}

	return client, err
}

func callContractKey(methodName string, token types.EthAddress, client *ethclient.Client, abi abi.ABI) (interface{}, error) {
	callData, err := abi.Pack(methodName)
	if err != nil {
		log.Fatalf("Failed to parse method %s on ABI", methodName)
	}
	result, err := callContractFn(callData, methodName, token, client, abi)
	return result[0], err
}

func callContractFn(callData []byte, methodName string,
	contract types.EthAddress, client *ethclient.Client, abi abi.ABI) ([]interface{}, error) {

	result, err := contractDataCall(client, contract, callData)
	if err != nil {
		log.Printf("Warning calling %s() on contract "+err.Error(), methodName)
		return nil, err
	}

	unparsed, err := abi.Unpack(methodName, result)
	if err != nil || len(unparsed) == 0 {
		log.Fatalf("Failed to parse %s result on ABI", methodName)
	}

	return unparsed, nil
}

func contractDataCall(client *ethclient.Client, contract types.EthAddress, data []byte) ([]byte, error) {
	addr := common.HexToAddress(string(contract))

	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
}
