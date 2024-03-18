package loader

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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

	// Retry subgraph query forever, because 99% of
	for err == nil {
		log.Println("Subgraph queried failed. Retrying in", SUBGRAPH_RETRY_SECS, "seconds. Error: ", err)

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

	req, err := http.NewRequest(http.MethodPost, cfg.Subgraph, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "crocswap-indexer/1.0")

	if err != nil {
		log.Println("Subgraph New Request Error: " + err.Error())
		return nil, err
	}

	client := &http.Client{}
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
	return body, nil
}

func readQueryPath(filename string) SubgraphQuery {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to read file: " + filename)
	}
	return SubgraphQuery(content)
}
