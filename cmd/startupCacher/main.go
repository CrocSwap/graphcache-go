package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"
	// "github.com/goccy/go-json"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type CacheMeta struct {
	ChainID  types.ChainId `json:"chain_id"`
	Subgraph string        `json:"subgraph"`
	// Cache includes everything until MaxBlock for each table (inclusive)
	MaxBlocks loader.SubgraphStartBlocks `json:"max_blocks"`
}

type TableState struct {
	QueryMinBlock *int
	ObsIds        map[string]struct{}
}

// Attempt to refresh the cache every 4 hours (though it will still be updated only if there's more than a day's worth of data).
const CACHE_UPDATE_INTERVAL = 4 * time.Hour

const MAX_BLOCK = 999999999

// Since we need to both partially parse and re-serialize rows from different tables, we need
// to dynamically unmarshal them and then parse `id` and `time` fields.
// Serializing them back is needed because we can't just dump raw query results, we need to
// group them in chunks.
type Entry map[string]json.RawMessage

// Reads data from the subgraph and caches it in a directory structure like:
// `{startupCacheDir}/{chainId}/{table}/{table}_{firstBlockOfChunk}-{lastBlockOfChunk}.json`
// Will only retrieve day-sized chunks of data that are at least 24 hours old to
// prevent caching bad data.
func startupCacher(cfg loader.ChainConfig, startupCacheDir string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		if r := recover(); r != nil {
			log.Printf("Cacher for %d recovered from panic: %s", cfg.ChainID, r)
		}
	}()

	chainId := types.IntToChainId(cfg.ChainID)
	syncCfg := loader.SyncChannelConfig{
		Chain:   cfg,
		Network: types.NetworkName(cfg.NetworkName),
	}

	chainPath := filepath.Join(startupCacheDir, string(chainId))
	meta, tableStates, err := loadMeta(startupCacheDir, chainId)
	if err != nil {
		log.Printf("Error loading cache meta for chain %s: %s", chainId, err)
		return
	}
	meta.Subgraph = cfg.Subgraph

	now := time.Now().UTC()
	stopDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	log.Printf("Caching data for chain %s from %v", chainId, meta.MaxBlocks)
	meta.ChainID = types.IntToChainId(cfg.ChainID)
	meta.Subgraph = cfg.Subgraph

	minBlocks := &meta.MaxBlocks

	tableEntries := make(map[string][]Entry)
	for {
		log.Printf("Querying subgraph for chain %s, blocks: Swap: %d Agg: %d Bal: %d Liq: %d Ko: %d Fee: %d", chainId, minBlocks.Swaps, minBlocks.Aggs, minBlocks.Bal, minBlocks.Liq, minBlocks.Ko, minBlocks.Fee)
		_, combined, err := loader.CombinedQuery(syncCfg, *minBlocks, MAX_BLOCK)
		if err != nil {
			log.Printf("Error querying subgraph for chain %s, retrying: %s", chainId, err)
			time.Sleep(10 * time.Second)
			continue
		}

		if combined.Meta.Block.Number == 0 {
			log.Printf("Warning subgraph for chain %s latest block is 0, retrying", chainId)
			time.Sleep(10 * time.Second)
			continue
		}

		tables := map[string]*json.RawMessage{
			"swaps":            &combined.Swaps,
			"aggEvents":        &combined.Aggs,
			"userBalances":     &combined.Bals,
			"feeChanges":       &combined.Fees,
			"liquidityChanges": &combined.Liqs,
		}

		if cfg.ChainID != 7700 {
			tables["knockoutCrosses"] = &combined.Kos
		}

		anyTableHasMore := false
		for tableName, tableData := range tables {
			entries := make([]Entry, 0)
			err = json.Unmarshal(*tableData, &entries)
			if err != nil {
				log.Printf("Error unmarshalling %s data for chain %s: %s", tableName, chainId, err)
				log.Println(string(*tableData))
				anyTableHasMore = true
				time.Sleep(10 * time.Second)
				continue
			}
			if len(entries) == 0 {
				log.Printf("Warning subgraph for chain %s for table %s returned no entries while at least one was expected", chainId, tableName)
				anyTableHasMore = true
				time.Sleep(10 * time.Second)
				continue
			}
			for _, entry := range entries {
				entryId, entryTime, entryBlock := parseEntryFields(entry)

				if entryTime.Before(stopDate) {

					if _, seen := tableStates[tableName].ObsIds[entryId]; !seen {
						// log.Println(entryTime, day, entryBlock, tableName, entryId)
						tableEntries[tableName] = append(tableEntries[tableName], entry)
						tableStates[tableName].ObsIds[entryId] = struct{}{}
						anyTableHasMore = true
					}
					if entryBlock > *tableStates[tableName].QueryMinBlock {
						*tableStates[tableName].QueryMinBlock = entryBlock
					}
				}
			}
		}

		tableLengths := map[string]int{
			"swaps":            len(tableStates["swaps"].ObsIds),
			"aggEvents":        len(tableStates["aggEvents"].ObsIds),
			"userBalances":     len(tableStates["userBalances"].ObsIds),
			"feeChanges":       len(tableStates["feeChanges"].ObsIds),
			"knockoutCrosses":  len(tableStates["knockoutCrosses"].ObsIds),
			"liquidityChanges": len(tableStates["liquidityChanges"].ObsIds),
		}

		log.Println("Table lengths for", chainId, tableLengths)
		saveChunks(&tableEntries, string(chainId), chainPath, false, ENTRIES_PER_FILE)
		if !anyTableHasMore {
			log.Printf("Finished querying %s: Swap: %d Agg: %d Bal: %d Liq: %d Ko: %d Fee: %d", chainId, minBlocks.Swaps, minBlocks.Aggs, minBlocks.Bal, minBlocks.Liq, minBlocks.Ko, minBlocks.Fee)
			break
		}
	}

	sanityCheck(tableStates)
	saveChunks(&tableEntries, string(chainId), chainPath, true, ENTRIES_PER_FILE)

	meta.MaxBlocks = *minBlocks
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		log.Printf("Error marshalling cache meta data for chain %s: %s", chainId, err)
		return
	}
	saveToFile(metaData, chainPath, "meta.json")

	runtime.GC()
	log.Printf("Chain %s cache updated to %s", chainId, stopDate)
}

// Bigger files make more sense but they take longer to load because of insertion sorting.
const ENTRIES_PER_FILE = 1000

// Saves chunks to storage, either all of them if `saveAll` is `true` or only until
// less than `entriesPerFile` entries are left.
func saveChunks(tableEntries *map[string][]Entry, chainId string, chainPath string, saveAll bool, entriesPerFile int) {
	for tableName, entries := range *tableEntries {
		tablePath := filepath.Join(chainPath, tableName)
		if len(entries) == 0 || (len(entries) < entriesPerFile && !saveAll) {
			continue
		}

		err := os.MkdirAll(tablePath, 0755)
		if err != nil {
			log.Printf("Error creating directory for chain %s for table %s: %s", chainId, tableName, err)
			continue
		}

		// Write entries as chunks of entriesPerFile entries
		remainingEntries := entries
		for {
			entryCount := min(entriesPerFile, len(remainingEntries))
			savingChunk := remainingEntries[:entryCount]
			_, _, firstBlock := parseEntryFields(savingChunk[0])
			_, _, lastBlock := parseEntryFields(savingChunk[len(savingChunk)-1])

			tableData, err := json.MarshalIndent(savingChunk, "", " ")
			if err != nil {
				log.Printf("Error marshalling entries for chain %s, table %s %d-%d: %s", chainId, tableName, firstBlock, lastBlock, err)
				continue
			}
			chunkFilename := fmt.Sprintf("%s_%d-%d.json", tableName, firstBlock, lastBlock)
			log.Printf("Saving entries for chain %s, table %s %d-%d to %s", chainId, tableName, firstBlock, lastBlock, chunkFilename)
			saveToFile(tableData, tablePath, chunkFilename)
			remainingEntries = remainingEntries[entryCount:]
			if (len(remainingEntries) < entriesPerFile && !saveAll) || (len(remainingEntries) == 0) {
				break
			}
		}
		// To *really* deallocate saved entries
		remainingCopy := make([]Entry, len(remainingEntries))
		copy(remainingCopy, remainingEntries)
		clear((*tableEntries)[tableName])
		delete(*tableEntries, tableName)
		(*tableEntries)[tableName] = remainingCopy
		runtime.GC()
	}
}

func saveToFile(data []byte, directory string, filename string) (err error) {
	tablePathTemp := filepath.Join(directory, "."+filename)
	err = os.WriteFile(tablePathTemp, data, 0644)
	if err != nil {
		log.Printf("Error writing data to file %s: %s", tablePathTemp, err)
		return
	}
	err = os.Rename(tablePathTemp, filepath.Join(directory, filename))
	if err != nil {
		log.Printf("Error renaming file %s: %s", tablePathTemp, err)
		return
	}
	return
}

// Loads chain meta from storage. Can take minutes for large chains.
func loadMeta(startupCacheDir string, chainId types.ChainId) (metaPtr *CacheMeta, states map[string]TableState, err error) {
	log.Printf("Loading meta for chain %s", chainId)
	var meta CacheMeta
	metaPtr = &meta
	meta.ChainID = chainId
	states = map[string]TableState{
		"swaps":            {QueryMinBlock: &meta.MaxBlocks.Swaps, ObsIds: make(map[string]struct{}, 100000)},
		"aggEvents":        {QueryMinBlock: &meta.MaxBlocks.Aggs, ObsIds: make(map[string]struct{}, 100000)},
		"userBalances":     {QueryMinBlock: &meta.MaxBlocks.Bal, ObsIds: make(map[string]struct{}, 100000)},
		"feeChanges":       {QueryMinBlock: &meta.MaxBlocks.Fee, ObsIds: make(map[string]struct{})},
		"knockoutCrosses":  {QueryMinBlock: &meta.MaxBlocks.Ko, ObsIds: make(map[string]struct{})},
		"liquidityChanges": {QueryMinBlock: &meta.MaxBlocks.Liq, ObsIds: make(map[string]struct{}, 100000)},
	}

	loadableTableNames := []string{}
	files, err := os.ReadDir(filepath.Join(startupCacheDir, string(chainId)))
	if err != nil {
		if os.IsNotExist(err) {
			return metaPtr, states, nil
		}
		log.Println("Error reading chain directory:", err)
		return nil, nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			if _, ok := states[file.Name()]; ok {
				loadableTableNames = append(loadableTableNames, file.Name())
			}
		}
	}

	for _, tableName := range loadableTableNames {
		tablePath := filepath.Join(startupCacheDir, string(chainId), tableName)
		files, err := os.ReadDir(tablePath)
		if err != nil {
			log.Printf("Error reading table %s for chain %s: %s", tableName, chainId, err)
			continue
		}
		tableChunks := make([]string, 0)
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), tableName+"_") {
				tableChunks = append(tableChunks, file.Name())
			}
		}
		slices.Sort(tableChunks)
		table := states[tableName]

		for _, tableChunk := range tableChunks {
			chunkPath := filepath.Join(tablePath, tableChunk)
			chunkData, err := os.ReadFile(chunkPath)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Printf("Error reading table chunk %s for chain %s: %s", chunkPath, chainId, err)
				}
				continue
			}
			var entries []Entry
			err = json.Unmarshal(chunkData, &entries)
			if err != nil {
				log.Printf("Error unmarshalling table chunk %s for chain %s: %s", chunkPath, chainId, err)
				continue
			}
			for _, entry := range entries {
				entryId, _, entryBlock := parseEntryFields(entry)
				table.ObsIds[entryId] = struct{}{}

				if entryBlock > *table.QueryMinBlock {
					*table.QueryMinBlock = entryBlock
				}
			}
			states[tableName] = table
		}
	}

	log.Printf("Loaded meta for chain %s: %v", chainId, meta)
	return metaPtr, states, err
}

func parseEntryFields(entry Entry) (entryId string, entryTime time.Time, entryBlock int) {
	entryId = string(entry["id"])
	entryTimeStr := string(entry["time"])
	entryTimeInt, err := strconv.ParseInt(entryTimeStr[1:len(entryTimeStr)-1], 10, 64)
	if err != nil {
		log.Printf("Error parsing entry time for %v: %s", entry, err)
		panic(err)
	}
	// log.Println(entryTimeInt)
	entryTime = time.Unix(entryTimeInt, 0).UTC()
	entryBlockStr := string(entry["block"])
	entryBlock, err = strconv.Atoi(entryBlockStr[1 : len(entryBlockStr)-1])
	if err != nil {
		log.Printf("Error parsing block number for %v: %s", entry, err)
		panic(err)
	}
	return
}

func sanityCheck(states map[string]TableState) bool {
	aggPartsSum := 0
	for tableName, table := range states {
		if tableName == "feeChanges" || tableName == "liquidityChanges" || tableName == "swaps" {
			aggPartsSum += len(table.ObsIds)
		}
	}
	if aggPartsSum != len(states["aggEvents"].ObsIds) {
		log.Printf("Warning: length of liq+swap+fee tables %d is different than agg table %d, cache is likely corrupted!", aggPartsSum, len(states["aggEvents"].ObsIds))
		return false
	}
	return true
}

func main() {
	var startupCacheDir = flag.String("startupCacheDir", "", "Directory to load startup cache from")
	var listenAddr = flag.String("listenAddr", "", "Enable serving cache over HTTP on this address")
	flag.Parse()

	enabledNetCfgPaths := flag.Args()
	if *startupCacheDir == "" {
		log.Fatal("No startup cache directory provided. Use -startupCacheDir to specify the directory.")
	}
	if len(enabledNetCfgPaths) == 0 {
		log.Fatal("No network config paths provided. Supply paths to network config files as arguments.")
	}
	log.Println("Got network config paths:", enabledNetCfgPaths)

	if *listenAddr != "" {
		StartCacheServer(*listenAddr, *startupCacheDir)
	} else {
		log.Println("Not starting the HTTP server, running in local mode")
	}

	chainCfgs := make(map[types.ChainId]loader.ChainConfig)
	for _, netCfgPath := range enabledNetCfgPaths {
		netCfg := loader.LoadNetworkConfig(netCfgPath)
		for _, chainCfg := range netCfg {
			chainCfgs[types.IntToChainId(chainCfg.ChainID)] = chainCfg
		}
	}

	if len(chainCfgs) == 0 {
		log.Fatal("No chain configurations loaded. Check network config paths.")
	}

	wg := sync.WaitGroup{}
	for {
		for _, chainCfg := range chainCfgs {
			wg.Add(1)
			go startupCacher(chainCfg, *startupCacheDir, &wg)
		}
		wg.Wait()
		time.Sleep(CACHE_UPDATE_INTERVAL)
	}
}
