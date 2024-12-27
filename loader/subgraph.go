package loader

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"
	// "github.com/goccy/go-json"
)

type GraphRequest[V GraphReqVars | CombinedGraphReqVars] struct {
	Query     SubgraphQuery `json:"query"`
	Variables V             `json:"variables"`
}

type GraphReqVars struct {
	Order   string `json:"orderDir"`
	MinTime int    `json:"minTime"`
	MaxTime int    `json:"maxTime"`
}

type CombinedGraphReqVars struct {
	Order             string `json:"orderDir"`
	SwapMinBlock      int    `json:"swapMinBlock"`
	SwapMaxBlock      int    `json:"swapMaxBlock"`
	LiquidityMinBlock int    `json:"liqMinBlock"`
	LiquidityMaxBlock int    `json:"liqMaxBlock"`
	AggMinBlock       int    `json:"aggMinBlock"`
	AggMaxBlock       int    `json:"aggMaxBlock"`
	BalMinBlock       int    `json:"balMinBlock"`
	BalMaxBlock       int    `json:"balMaxBlock"`
	FeeMinBlock       int    `json:"feeMinBlock"`
	FeeMaxBlock       int    `json:"feeMaxBlock"`
	KoMinBlock        int    `json:"koMinBlock"`
	KoMaxBlock        int    `json:"koMaxBlock"`
}

type SubgraphQuery string

type SubgraphError struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data json.RawMessage `json:"data"`
}

func makeSubgraphVars(isAsc bool, startTime, endTime int) GraphReqVars {
	order := "desc"
	if isAsc {
		order = "asc"
	}
	return GraphReqVars{
		Order:   order,
		MinTime: startTime,
		MaxTime: endTime,
	}
}

func makeCombinedSubgraphVars(isAsc bool, swapMinBlock, swapMaxBlock, liqMinBlock, liqMaxBlock, aggMinBlock, aggMaxBlock, balMinBlock, balMaxBlock, feeMinBlock, feeMaxBlock, koMinBlock, koMaxBlock int) CombinedGraphReqVars {
	order := "desc"
	if isAsc {
		order = "asc"
	}
	return CombinedGraphReqVars{
		Order:             order,
		SwapMinBlock:      swapMinBlock,
		SwapMaxBlock:      swapMaxBlock,
		LiquidityMinBlock: liqMinBlock,
		LiquidityMaxBlock: liqMaxBlock,
		AggMinBlock:       aggMinBlock,
		AggMaxBlock:       aggMaxBlock,
		BalMinBlock:       balMinBlock,
		BalMaxBlock:       balMaxBlock,
		FeeMinBlock:       feeMinBlock,
		FeeMaxBlock:       feeMaxBlock,
		KoMinBlock:        koMinBlock,
		KoMaxBlock:        koMaxBlock,
	}
}

const SUBGRAPH_RETRY_SECS = 5

/* Retry subgraph query forever, because 99% of the time it's an issue with the subgraph endpoint
 * or subgraph rate limiting. Crashign and restarting won't fix the process and will result in loss
 * of state. But be aware that pods may still look health even if subgraph isn't working. */
func queryFromSubgraph(cfg ChainConfig, query SubgraphQuery, startTime int, endTime int, isAsc bool) ([]byte, error) {
	req := GraphRequest[GraphReqVars]{
		Query:     query,
		Variables: makeSubgraphVars(isAsc, startTime, endTime),
	}
	result, err := queryFromSubgraphTry(cfg, req)

	for err != nil {
		log.Printf("Subgraph %s query failed. Retrying in %d seconds. Error: %s", cfg.HexChainID(), SUBGRAPH_RETRY_SECS, err.Error())

		time.Sleep(time.Duration(SUBGRAPH_RETRY_SECS) * time.Second)
		result, err = queryFromSubgraphTry(cfg, req)
	}

	return result, err
}

func queryFromSubgraphCombined(cfg ChainConfig, query SubgraphQuery, isAsc bool, minBlocks SubgraphStartBlocks, maxBlock int) ([]byte, error) {
	req := GraphRequest[CombinedGraphReqVars]{
		Query:     query,
		Variables: makeCombinedSubgraphVars(isAsc, minBlocks.Swaps, maxBlock, minBlocks.Liq, maxBlock, minBlocks.Aggs, maxBlock, minBlocks.Bal, maxBlock, minBlocks.Fee, maxBlock, minBlocks.Ko, maxBlock),
	}
	result, err := queryFromSubgraphTry(cfg, req)

	for err != nil {
		log.Printf("Subgraph %s combined query failed. Retrying in %d seconds. Error: %s", cfg.HexChainID(), SUBGRAPH_RETRY_SECS, err.Error())

		time.Sleep(time.Duration(SUBGRAPH_RETRY_SECS) * time.Second)
		result, err = queryFromSubgraphTry(cfg, req)
	}

	return result, err
}

func queryFromSubgraphTry[V GraphReqVars | CombinedGraphReqVars](cfg ChainConfig, request GraphRequest[V]) ([]byte, error) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		log.Println("Subgraph Query Request Error:" + err.Error())
		return nil, err
	}

	url := strings.Replace(cfg.Subgraph, "[api-key]", os.Getenv("GRAPH_API_KEY"), 1)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "crocswap-indexer/1.0")

	if err != nil {
		log.Println("Subgraph New Request Error: " + err.Error())
		return nil, err
	}

	client := &http.Client{}
	client.Timeout = 20 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Subgraph Query Connection Error: " + err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Subgraph Query Read Error: " + err.Error())
		return nil, err
	}

	var subgraphErr SubgraphError
	err = json.Unmarshal(body, &subgraphErr)
	if err == nil && len(subgraphErr.Errors) > 0 {
		log.Println("Subgraph Query Error(s): ", subgraphErr.Errors)
		return nil, fmt.Errorf(subgraphErr.Errors[0].Message)
	}
	if subgraphErr.Data == nil {
		log.Println("Subgraph Query Error: Subgraph response has no data field:", string(body))
		return nil, fmt.Errorf("subgraph response has no data field")
	}
	return body, nil
}

func readQueryPath(filename string) SubgraphQuery {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to read file: " + filename)
	}
	return SubgraphQuery(content)
}
