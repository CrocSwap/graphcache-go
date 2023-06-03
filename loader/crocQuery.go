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
	callData, err := q.queryAbi.Pack("queryAmbientTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), big.NewInt(int64(pos.PoolIdx)))
	if err != nil {
		log.Fatalf("Failed to parse queryPrice on ABI: %s", err.Error())
	}

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryAmbientTokens")
}

func (q *CrocQuery) QueryRangeLiquidity(pos types.PositionLocation) (*big.Int, error) {
	callData, err := q.queryAbi.Pack("queryRangeTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)),
		big.NewInt(int64(pos.PoolIdx)),
		big.NewInt(int64(pos.BidTick)),
		big.NewInt(int64(pos.AskTick)))

	if err != nil {
		log.Fatalf("Failed to parse queryRangeTokens on ABI: %s", err.Error())
	}

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryRangeTokens")
}

func (q *CrocQuery) QueryRangeRewardsLiq(pos types.PositionLocation) (*big.Int, error) {
	callData, err := q.queryAbi.Pack("queryConcRewards",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)),
		big.NewInt(int64(pos.PoolIdx)),
		big.NewInt(int64(pos.BidTick)), big.NewInt(int64(pos.AskTick)))

	if err != nil {
		log.Fatalf("Failed to parse queryConcRewards on ABI: %s", err.Error())
	}

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryConcRewards")
}

func (q *CrocQuery) QueryKnockoutLiq(pos types.KOClaimLocation) (*big.Int, error) {
	callData, err := q.queryAbi.Pack("queryKnockoutTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)),
		big.NewInt(int64(pos.PoolIdx)),
		uint32(pos.PivotTime), pos.IsBid,
		big.NewInt(int64(pos.BidTick)),
		big.NewInt(int64(pos.AskTick)))

	if err != nil {
		log.Fatalf("Failed to parse queryKnockoutTokens on ABI: %s", err.Error())
	}

	result, err := q.callQueryResults(pos.ChainId, callData, "queryKnockoutTokens")

	if err != nil {
		return big.NewInt(0), err
	}

	return result[0].(*big.Int), nil
}

func (q *CrocQuery) QueryKnockoutPivot(pos types.PositionLocation) (uint32, error) {
	tick := pos.LiquidityLocation.PivotTick()
	callData, err := q.queryAbi.Pack("queryKnockoutPivot",
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)),
		big.NewInt(int64(pos.PoolIdx)),
		pos.IsBid, big.NewInt(int64(tick)))

	if err != nil {
		log.Fatalf("Failed to parse queryKnockoutPivot on ABI: %s", err.Error())
	}

	result, err := q.callQueryResults(pos.ChainId, callData, "queryKnockoutPivot")

	if err != nil {
		return 0, err
	}

	return result[1].(uint32), nil
}

func (q *CrocQuery) callQueryResults(chainId types.ChainId,
	callData []byte, methodName string) ([]interface{}, error) {

	client, err := q.chain.ethClientForChain(chainId)

	if err != nil {
		return make([]interface{}, 0), err
	}

	contractAddr, err := q.getQueryContractAddr(chainId)
	if err != nil {
		return make([]interface{}, 0), err
	}

	return callContractFn(callData, methodName, contractAddr, client, q.queryAbi)
}

func (q *CrocQuery) callQueryFirstReturn(chainId types.ChainId,
	callData []byte, methodName string) (*big.Int, error) {
	result, err := q.callQueryResults(chainId, callData, methodName)

	if err != nil {
		return big.NewInt(0), err
	}
	return result[0].(*big.Int), nil

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
