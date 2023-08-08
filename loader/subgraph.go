package loader

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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

func QueryFromSubgraph(cfg ChainConfig, query SubgraphQuery, startTime int, endTime int, isAsc bool) ([]byte, error) {
	request := GraphRequest{
		Query:     query,
		Variables: makeSubgraphVars(isAsc, startTime, endTime),
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		log.Println("Subgraph Query Request Error: " + err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.Subgraph, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")

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

func ReadQueryPath(filename string) SubgraphQuery {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to read file: " + filename)
	}
	return SubgraphQuery(content)
}
