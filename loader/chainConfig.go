package loader

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/CrocSwap/graphcache-go/types"
)

type ChainConfig struct {
	ChainID             int                 `json:"chain_id"`
	RPCs                map[string][]string `json:"rpcs"`
	Subgraph            string              `json:"subgraph"`
	DexContract         string              `json:"dex_contract"`
	QueryContract       string              `json:"query_contract"`
	QueryContractABI    string              `json:"query_contract_abi"`
	POAMiddleware       bool                `json:"poa_middleware"`
	BlockTime           float64             `json:"block_time"`
	Ignore              bool                `json:"ignore,omitempty"`
	EnableRPCCache      bool                `json:"enable_rpc_cache"`
	EnableSubgraphCache bool                `json:"enable_subgraph_cache"`
}

type NetworkConfig map[types.NetworkName]ChainConfig

func LoadChainConfigs(path string) NetworkConfig {
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var config NetworkConfig

	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func (c *NetworkConfig) networkForChainID(chainId types.ChainId) (types.NetworkName, bool) {
	for networkKey, configElem := range *c {
		if chainId == types.IntToChainId(configElem.ChainID) {
			return networkKey, true
		}
	}
	return "", false
}

func (c *NetworkConfig) chainIDForNetwork(network types.NetworkName) (types.ChainId, bool) {
	lookup, ok := (*c)[network]
	if ok {
		return types.IntToChainId(lookup.ChainID), true
	} else {
		return "", false
	}
}
