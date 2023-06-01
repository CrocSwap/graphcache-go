package loader

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type CrocQuery struct {
	queryAbi abi.ABI
	addrs    map[types.ChainId]types.EthAddress
	chain    *OnChainLoader
}

func NewCrocQuery(chain *OnChainLoader) *CrocQuery {
	return &CrocQuery{
		queryAbi: crocQueryAbi(),
		addrs:    crocQueryAddrs(),
		chain:    chain,
	}
}

func (q *CrocQuery) QueryAmbientSeeds(pos types.PositionLocation) (*big.Int, error) {
	client, err := q.chain.ethClientForChain(pos.ChainId)

	if err != nil {
		return big.NewInt(0), err
	}

	contractAddr, err := q.getQueryContractAddr(pos.ChainId)
	if err != nil {
		return big.NewInt(0), err
	}

	poolInt := big.NewInt(int64(pos.PoolIdx))

	callData, err := q.queryAbi.Pack("queryAmbientTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), &poolInt)
	if err != nil {
		log.Fatalf("Failed to parse queryPrice on ABI: %s", err.Error())
	}

	result, err := callContractFn(callData, "queryAmbientTokens", contractAddr, client, q.queryAbi)

	if err != nil {
		return big.NewInt(0), nil
	}

	return result[0].(*big.Int), nil
}

func (q *CrocQuery) QueryRangeLiquidity(pos types.PositionLocation) (*big.Int, error) {
	client, err := q.chain.ethClientForChain(pos.ChainId)

	if err != nil {
		return big.NewInt(0), err
	}

	contractAddr, err := q.getQueryContractAddr(pos.ChainId)
	if err != nil {
		return big.NewInt(0), err
	}

	poolInt := big.NewInt(int64(pos.PoolIdx))
	bidTick := big.NewInt(int64(pos.BidTick))
	askTick := big.NewInt(int64(pos.AskTick))
	callData, err := q.queryAbi.Pack("queryRangeTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), &poolInt,
		&bidTick, &askTick)

	if err != nil {
		log.Fatalf("Failed to parse queryRangeTokens on ABI: %s", err.Error())
	}

	result, err := callContractFn(callData, "queryRangeTokens", contractAddr, client, q.queryAbi)

	if err != nil {
		return big.NewInt(0), nil
	}

	return result[0].(*big.Int), nil
}

func (q *CrocQuery) QueryRangeRewardsLiq(pos types.PositionLocation) (*big.Int, error) {
	client, err := q.chain.ethClientForChain(pos.ChainId)

	if err != nil {
		return big.NewInt(0), err
	}

	contractAddr, err := q.getQueryContractAddr(pos.ChainId)
	if err != nil {
		return big.NewInt(0), err
	}

	poolInt := big.NewInt(int64(pos.PoolIdx))
	bidTick := big.NewInt(int64(pos.BidTick))
	askTick := big.NewInt(int64(pos.AskTick))
	callData, err := q.queryAbi.Pack("queryConcRewards",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), &poolInt,
		&bidTick, &askTick)

	if err != nil {
		log.Fatalf("Failed to parse queryConcRewards on ABI: %s", err.Error())
	}

	result, err := callContractFn(callData, "queryConcRewards", contractAddr, client, q.queryAbi)

	if err != nil {
		return big.NewInt(0), nil
	}

	return result[0].(*big.Int), nil
}

func (q *CrocQuery) QueryKnockoutLiq(pos types.PositionLocation) (*big.Int, bool, error) {
	client, err := q.chain.ethClientForChain(pos.ChainId)

	if err != nil {
		return big.NewInt(0), false, err
	}

	contractAddr, err := q.getQueryContractAddr(pos.ChainId)
	if err != nil {
		return big.NewInt(0), false, err
	}

	poolInt := big.NewInt(int64(pos.PoolIdx))
	bidTick := big.NewInt(int64(pos.BidTick))
	askTick := big.NewInt(int64(pos.AskTick))
	callData, err := q.queryAbi.Pack("queryKnockoutTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), &poolInt,
		uint32(pos.PivotTime), pos.IsBid, &bidTick, &askTick)

	if err != nil {
		log.Fatalf("Failed to parse queryKnockoutTokens on ABI: %s", err.Error())
	}

	result, err := callContractFn(callData, "queryKnockoutTokens", contractAddr, client, q.queryAbi)

	if err != nil {
		return big.NewInt(0), false, nil
	}

	return result[0].(*big.Int), result[3].(bool), nil
}

func (q *CrocQuery) getQueryContractAddr(chain types.ChainId) (types.EthAddress, error) {
	addr, ok := q.addrs[chain]

	if !ok {
		log.Printf("No CrocQuery contract foudn for %s", chain)
		err := fmt.Errorf("No CrocQuery contract foudn for %s", chain)
		return "", err
	}

	return addr, nil
}

func crocQueryAbi() abi.ABI {
	filePath := "./artifacts/abis/CrocQuery.json"
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("Failed to read ABI contract at " + filePath)
	}

	parsedABI, err := abi.JSON(file)
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	return parsedABI
}

func crocQueryAddrs() map[types.ChainId]types.EthAddress {
	addrs := make(map[types.ChainId]types.EthAddress)
	addrs["0x5"] = "0xc9900777baa5EE94Cd2C6509fb09278A1A46b7e8"
	addrs["0x1"] = "0xc2e1f740E11294C64adE66f69a1271C5B32004c8"
	return addrs
}
