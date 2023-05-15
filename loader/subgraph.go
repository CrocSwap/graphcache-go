package loader

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GraphRequest struct {
	Query     SubgraphQuery `json:"query"`
	Variables GraphReqVars  `json:"variables"`
}

type GraphReqVars struct {
	MinTime int `json:"minTime"`
	MaxTime int `json:"maxTime"`
}

type SubgraphQuery string

func queryFromSubgraph(cfg ChainConfig, query SubgraphQuery, minTime int) ([]byte, error) {
	request := GraphRequest{
		Query: query,
		Variables: GraphReqVars{
			MinTime: minTime,
			MaxTime: int(time.Now().Unix()),
		},
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Subgraph Query Read Error: " + err.Error())
		return nil, err
	}
	return body, nil
}

func readQueryPath(filename string) SubgraphQuery {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to read file: " + filename)
	}
	return SubgraphQuery(content)
}
