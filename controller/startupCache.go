package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

func LoadStartupCache(startupCacheSource string, syncer SubgraphSyncer) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Warning: panic during startup cache loading:", r)
		}
	}()
	log.Printf("Loading startup cache from %s", startupCacheSource)
	cache := NewStartupCacheProvider(startupCacheSource, syncer.ChainId())

	lastBlocks := loader.SubgraphStartBlocks{}
	lastBlocksMap := map[string]*int{
		"swaps":            &lastBlocks.Swaps,
		"aggEvents":        &lastBlocks.Aggs,
		"liquidityChanges": &lastBlocks.Liq,
		"knockoutCrosses":  &lastBlocks.Ko,
		"feeChanges":       &lastBlocks.Fee,
		"userBalances":     &lastBlocks.Bal,
	}

	maxBlock := 0 // just for visual output
	allChunks := make([]string, 0)
	for tableName := range lastBlocksMap {
		chunkNames, err := cache.GetChunks(tableName)
		if err != nil {
			log.Printf("Warning: failed to get startup cache chunks for %s: %s", tableName, err)
			return
		}
		allChunks = append(allChunks, chunkNames...)
		for _, chunkName := range chunkNames {
			blockStr := strings.Split(strings.Split(chunkName, "_")[1], "-")[1]
			block, _ := strconv.Atoi(blockStr[:len(blockStr)-5])
			if block > maxBlock {
				maxBlock = block
			}
		}
	}

	// To prevent slowdown due to sorting when inserting into TX history, we interleave chunks from each table
	slices.SortFunc(allChunks, func(i, j string) int {
		iBlockStr := strings.Split(strings.Split(i, "_")[1], "-")[0]
		iBlock, _ := strconv.Atoi(iBlockStr)
		jBlockStr := strings.Split(strings.Split(j, "_")[1], "-")[0]
		jBlock, _ := strconv.Atoi(jBlockStr)
		if iBlock < jBlock {
			return -1
		} else if iBlock > jBlock {
			return 1
		}
		return 0
	})

	for _, chunkName := range allChunks {
		tableName := strings.Split(chunkName, "_")[0]
		log.Printf("Loading startup chunk %s", chunkName)
		data, err := cache.GetTableChunk(tableName, chunkName)
		if err != nil {
			log.Printf("Warning: failed to get chunk %s", chunkName)
			return
		}

		if data == nil {
			log.Printf("Warning: no data for chunk %s", chunkName)
			return
		}

		lastObs, _, err := syncer.IngestEntries(tableName, data, *lastBlocksMap[tableName], maxBlock)
		if err != nil {
			log.Printf("Warning: failed to ingest startup cache chunk %s for table %s: %s", chunkName, tableName, err)
			return
		}
		*lastBlocksMap[tableName] = lastObs

		syncer.SetStartBlocks(lastBlocks)
	}

	syncer.SetStartBlocks(lastBlocks)
	log.Printf("Startup cache loaded. Swap: %d Agg: %d Bal: %d Liq: %d Ko: %d Fee: %d", lastBlocks.Swaps, lastBlocks.Aggs, lastBlocks.Bal, lastBlocks.Liq, lastBlocks.Ko, lastBlocks.Fee)
}

type startupCacheProvider struct {
	cacheSource string
	chainId     types.ChainId
	isHttp      bool
}

func NewStartupCacheProvider(cacheSource string, chainId types.ChainId) startupCacheProvider {
	isHttp := false
	if strings.HasPrefix(cacheSource, "http") {
		isHttp = true
	}
	return startupCacheProvider{
		cacheSource: cacheSource,
		chainId:     chainId,
		isHttp:      isHttp,
	}
}

// If startup loader encounters an error the indexer will have to continue syncing from the
// last loaded block, which might be slow. Better to retry a few times before giving up.
const HTTP_MAX_ATTEMPTS = 5

func (p startupCacheProvider) GetChunks(tableName string) (chunks []string, err error) {
	if p.isHttp {
		attempts := 0
		var resp *http.Response
		for {
			attempts++
			if attempts > HTTP_MAX_ATTEMPTS {
				return nil, err
			}
			h := http.Client{Timeout: 10 * time.Second}
			resp, err = h.Get(fmt.Sprintf("%s/%s/%s/chunks.json", p.cacheSource, p.chainId, tableName))
			if err != nil {
				log.Printf("Warning: failed to fetch startup cache days: %s. Attempt %d/%d", err, attempts, HTTP_MAX_ATTEMPTS)
				time.Sleep(5 * time.Second)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Printf("Warning: failed to fetch startup cache days, status code: %d. Attempt %d/%d", resp.StatusCode, attempts, HTTP_MAX_ATTEMPTS)
				time.Sleep(5 * time.Second)
				continue
			}

			var data []byte
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Warning: failed to read startup cache days response: %s. Attempt %d/%d", err, attempts, HTTP_MAX_ATTEMPTS)
				time.Sleep(5 * time.Second)
				continue
			}

			err = json.Unmarshal(data, &chunks)
			return chunks, err
		}
	} else {
		tablePath := filepath.Join(p.cacheSource, string(p.chainId), tableName)
		files, err := os.ReadDir(tablePath)
		if err != nil {
			log.Println("Warning: failed to read startup cache table directory:", err)
			return nil, err
		}

		chunkNames := []string{}
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), tableName+"_") {
				chunkNames = append(chunkNames, file.Name())
			}
		}
		slices.SortFunc(chunkNames, func(i, j string) int {
			iBlockStr := strings.Split(strings.Split(i, "_")[1], "-")[0]
			iBlock, _ := strconv.Atoi(iBlockStr)
			jBlockStr := strings.Split(strings.Split(j, "_")[1], "-")[0]
			jBlock, _ := strconv.Atoi(jBlockStr)
			if iBlock < jBlock {
				return -1
			} else if iBlock > jBlock {
				return 1
			}
			return 0
		})
		return chunkNames, nil
	}
}

func (p startupCacheProvider) GetTableChunk(tableName string, chunkName string) (data []byte, err error) {
	if p.isHttp {
		var resp *http.Response
		attempts := 0
		for {
			attempts++
			if attempts > HTTP_MAX_ATTEMPTS {
				return nil, err
			}
			// call the http endpoint with /{chainId}/{day}/{table} to get the table data
			h := http.Client{Timeout: 10 * time.Second}
			resp, err = h.Get(fmt.Sprintf("%s/%s/%s/%s", p.cacheSource, p.chainId, tableName, chunkName))
			if err != nil {
				log.Printf("Warning: failed to fetch startup cache chunk %s for table %s: %s. Attempt %d/%d", chunkName, tableName, err, attempts, HTTP_MAX_ATTEMPTS)
				time.Sleep(5 * time.Second)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				return nil, nil
			}

			if resp.StatusCode != 200 {
				log.Printf("Warning: failed to fetch startup cache chunk %s for table %s, status code: %d. Attempt %d/%d", chunkName, tableName, resp.StatusCode, attempts, HTTP_MAX_ATTEMPTS)
				time.Sleep(5 * time.Second)
				continue
			}

			data, err = io.ReadAll(resp.Body)
			return data, err
		}
	} else {
		tablePath := filepath.Join(p.cacheSource, string(p.chainId), tableName, chunkName)
		data, err := os.ReadFile(tablePath)
		if err != nil {
			log.Println("Warning: failed to read startup cache chunk:", err)
			return nil, err
		}
		return data, nil
	}
}
