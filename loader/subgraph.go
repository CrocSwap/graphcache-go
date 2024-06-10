package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type GraphRequest struct {
	Query     SubgraphQuery `json:"query"`
	Variables GraphReqVars  `json:"variables"`
}

type GraphReqVars struct {
	Order   string `json:"orderDir"`
	MinTime int    `json:"minTime"`
	MaxTime int    `json:"maxTime"`
}

type SubgraphQuery string

type SubgraphError struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data json.RawMessage `json:"data"`
}

func makeSubgraphVars(isAsc bool, startTime, endTime int) GraphReqVars {
	if isAsc {
		return GraphReqVars{
			Order:   "asc",
			MinTime: startTime,
			MaxTime: endTime,
		}
	} else {
		return GraphReqVars{
			Order:   "desc",
			MinTime: startTime,
			MaxTime: endTime,
		}
	}
}

const SUBGRAPH_RETRY_SECS = 5

/* Retry subgraph query forever, because 99% of the time it's an issue with the subgraph endpoint
 * or subgraph rate limiting. Crashign and restarting won't fix the process and will result in loss
 * of state. But be aware that pods may still look health even if subgraph isn't working. */
func queryFromSubgraph(cfg ChainConfig, query SubgraphQuery, startTime int, endTime int, isAsc bool) ([]byte, error) {
	result, err := queryFromSubgraphTry(cfg, query, startTime, endTime, isAsc)

	for err != nil {
		log.Println("Subgraph query failed. Retrying in", SUBGRAPH_RETRY_SECS, "seconds. Error: ", err)

		time.Sleep(time.Duration(SUBGRAPH_RETRY_SECS) * time.Second)
		result, err = queryFromSubgraphTry(cfg, query, startTime, endTime, isAsc)
	}

	return result, err
}

func queryFromSubgraphTry(cfg ChainConfig, query SubgraphQuery, startTime int, endTime int, isAsc bool) ([]byte, error) {
	request := GraphRequest{
		Query:     query,
		Variables: makeSubgraphVars(isAsc, startTime, endTime),
	}

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
