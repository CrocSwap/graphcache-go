package loader

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"runtime/debug"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type ICrocQuery interface {
	QueryAmbientLiq(pos types.PositionLocation) (*big.Int, error)
	QueryRangeLiquidity(pos types.PositionLocation) (*big.Int, error)
	QueryRangeRewardsLiq(pos types.PositionLocation) (*big.Int, error)
	QueryKnockoutLiq(pos types.KOClaimLocation) (KnockoutLiqResp, error)
	QueryKnockoutPivot(pos types.PositionLocation) (uint32, error)
}

type NonCrocQuery struct{}

func (q *NonCrocQuery) QueryAmbientLiq(pos types.PositionLocation) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (q *NonCrocQuery) QueryRangeLiquidity(pos types.PositionLocation) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (q *NonCrocQuery) QueryRangeRewardsLiq(pos types.PositionLocation) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (q *NonCrocQuery) QueryKnockoutLiq(pos types.KOClaimLocation) (KnockoutLiqResp, error) {
	return KnockoutLiqResp{Liq: big.NewInt(0), KnockedOut: false}, nil
}

func (q *NonCrocQuery) QueryKnockoutPivot(pos types.PositionLocation) (uint32, error) {
	return 0, nil
}

type CrocQuery struct {
	queryAbi abi.ABI
	addrs    map[types.ChainId]types.EthAddress
	chain    *OnChainLoader
}

func NewCrocQuery(chain *OnChainLoader) *CrocQuery {
	return &CrocQuery{
		queryAbi: crocQueryAbi(),
		addrs:    crocQueryAddrs(chain.Cfg),
		chain:    chain,
	}
}

func (q *CrocQuery) QueryAmbientLiq(pos types.PositionLocation) (*big.Int, error) {
	callData, err := q.queryAbi.Pack("queryAmbientTokens",
		common.HexToAddress(string(pos.User)),
		common.HexToAddress(string(pos.Base)),
		common.HexToAddress(string(pos.Quote)), big.NewInt(int64(pos.PoolIdx)))
	if err != nil {
		log.Fatalf("Failed to parse queryPrice on ABI: %s", err.Error())
	}

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryAmbientTokens", nil)
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

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryRangeTokens", nil)
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

	return q.callQueryFirstReturn(pos.ChainId, callData, "queryConcRewards", nil)
}

type KnockoutLiqResp struct {
	Liq        *big.Int
	KnockedOut bool
}

func (q *CrocQuery) QueryKnockoutLiq(pos types.KOClaimLocation) (resp KnockoutLiqResp, err error) {
	resp = KnockoutLiqResp{Liq: big.NewInt(0), KnockedOut: false}
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

	result, err := q.callQueryResults(pos.ChainId, callData, "queryKnockoutTokens", nil)

	if err != nil {
		return
	}

	resp.Liq = result[0].(*big.Int)
	resp.KnockedOut = result[3].(bool)
	return
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

	result, err := q.callQueryResults(pos.ChainId, callData, "queryKnockoutPivot", nil)

	if err != nil {
		return 0, err
	}

	return result[1].(uint32), nil
}

func (q *CrocQuery) QueryPoolPrice(pool types.PoolLocation, blockNumber *big.Int) (price float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in QueryPoolPrice: %v", r)
			log.Println("Stack trace: ", string(debug.Stack()))
			price = 0.0
			err = fmt.Errorf("panic in QueryPoolPrice")
		}
	}()
	callData, err := q.queryAbi.Pack("queryPrice",
		common.HexToAddress(string(pool.Base)), common.HexToAddress(string(pool.Quote)), big.NewInt(int64(pool.PoolIdx)))
	if err != nil {
		log.Fatalf("Failed to parse queryPrice on ABI: %s", err.Error())
	}

	crocPrice, err := q.callQueryFirstReturn(pool.ChainId, callData, "queryPrice", blockNumber)
	if err != nil {
		return 0.0, err
	}
	crocPriceF := new(big.Float).SetInt(crocPrice)
	denom := new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 64))
	crocPriceF.Quo(crocPriceF, denom)
	crocPriceF.Mul(crocPriceF, crocPriceF)
	price, _ = crocPriceF.Float64()
	return
}

func (q *CrocQuery) callQueryResults(chainId types.ChainId,
	callData []byte, methodName string, blockNumber *big.Int) ([]interface{}, error) {

	client, err := q.chain.ethClientForChain(chainId)

	if err != nil {
		return make([]interface{}, 0), err
	}

	contractAddr, err := q.getQueryContractAddr(chainId)
	if err != nil {
		return make([]interface{}, 0), err
	}

	return q.chain.callContractFn(callData, methodName, contractAddr, client, chainId, q.queryAbi, blockNumber)
}

func (q *CrocQuery) callQueryFirstReturn(chainId types.ChainId,
	callData []byte, methodName string, blockNumber *big.Int) (*big.Int, error) {
	result, err := q.callQueryResults(chainId, callData, methodName, blockNumber)

	if err != nil {
		return big.NewInt(0), err
	}
	return result[0].(*big.Int), nil
}

func (q *CrocQuery) getQueryContractAddr(chain types.ChainId) (types.EthAddress, error) {
	addr, ok := q.addrs[chain]

	if !ok {
		log.Printf("No CrocQuery contract foudn for %s", chain)
		err := fmt.Errorf("no CrocQuery contract foudn for %s", chain)
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

func crocQueryAddrs(cfg NetworkConfig) map[types.ChainId]types.EthAddress {
	addrs := make(map[types.ChainId]types.EthAddress)

	for _, chainCfg := range cfg {
		chainId := types.IntToChainId(chainCfg.ChainID)
		addrs[chainId] = types.EthAddress(chainCfg.QueryContract)
	}

	return addrs
}
