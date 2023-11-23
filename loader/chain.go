package loader

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Call3InputType struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

type Call3OutputType struct {
	Success    bool
	ReturnData []uint8
}

type CallJob struct {
	Contract types.EthAddress
	CallData []byte
	Result   chan []byte
}

type OnChainLoader struct {
	Cfg           NetworkConfig
	JobChans      map[int]chan CallJob
	BatchInterval time.Duration
	multicallAbi  abi.ABI
}

func NewOnChainLoader(cfg NetworkConfig) *OnChainLoader {
	chans := make(map[int]chan CallJob)
	for _, chain := range cfg {
		chans[chain.ChainID] = make(chan CallJob)
	}

	c := &OnChainLoader{
		Cfg:           cfg,
		JobChans:      chans,
		BatchInterval: time.Duration(500) * time.Millisecond,
		multicallAbi:  multicallAbi(),
	}
	for key, chain := range cfg {
		go c.multicallWorker(chain.ChainID, key)
	}
	return c
}

func multicallAbi() abi.ABI {
	filePath := "./artifacts/abis/Multicall.json"
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

func (c *OnChainLoader) ethClientForChain(chainId types.ChainId) (*ethclient.Client, error) {
	cfg, okay := c.Cfg.ChainConfig(chainId)

	if !okay {
		log.Println("Warning no chain configuration for " + chainId)
		return nil, fmt.Errorf("Chain configuration missing")
	}

	rpcUrl := cfg.RPCEndpoint
	client, err := ethclient.DialContext(context.Background(), rpcUrl)

	if err != nil {
		log.Println("Warning RPC connection error: " + err.Error())
	}

	return client, err
}

func (c *OnChainLoader) callContractKey(methodName string, token types.EthAddress, client *ethclient.Client, chainId types.ChainId, abi abi.ABI) (interface{}, error) {
	callData, err := abi.Pack(methodName)
	if err != nil {
		log.Fatalf("Failed to parse method %s on ABI", methodName)
	}
	result, err := c.callContractFn(callData, methodName, token, client, chainId, abi)
	return result[0], err
}

func (c *OnChainLoader) callContractFn(callData []byte, methodName string,
	contract types.EthAddress, client *ethclient.Client, chainId types.ChainId, abi abi.ABI) ([]interface{}, error) {

	result, err := c.contractDataCall(client, chainId, contract, callData)
	if err != nil {
		log.Printf("Warning calling %s() on contract "+err.Error(), methodName)
		return nil, err
	}

	unparsed, err := abi.Unpack(methodName, result)
	if err != nil || len(unparsed) == 0 {
		log.Fatalf("Failed to parse %s result on ABI", methodName)
	}

	return unparsed, nil
}

func (c *OnChainLoader) contractDataCall(client *ethclient.Client, chainId types.ChainId, contract types.EthAddress, data []byte) ([]byte, error) {
	job := CallJob{
		Contract: contract,
		CallData: data,
		Result:   make(chan []byte, 1),
	}
	chainIdInt, _ := strconv.ParseInt(string(chainId)[2:], 16, 32)
	c.JobChans[int(chainIdInt)] <- job

	// Wait for the multicall result and fall back to a direct call if it times out
	select {
	case result := <-job.Result:
		if len(result) == 0 {
			return []byte{}, fmt.Errorf("empty result from multicall, error")
		}
		return result, nil
	case <-time.After(2000 * time.Millisecond):
		log.Println("Multicall timed out, sending manually")
		addr := common.HexToAddress(string(contract))

		msg := ethereum.CallMsg{
			To:   &addr,
			Data: data,
		}

		result, err := client.CallContract(context.Background(), msg, nil)
		if err != nil {
			return []byte{}, err
		}
		return result, nil
	}
}

// Goroutine that aggregates calls and sends them to the multicall contract after a timeout
func (c *OnChainLoader) multicallWorker(chainId int, networkName types.NetworkName) {
	jobChan := c.JobChans[chainId]
	jobs := make([]CallJob, 0)
	batchTimer := time.NewTimer(1<<63 - 1)
	maxBatchSize := c.Cfg[networkName].MulticallMaxBatch
	if maxBatchSize == 0 {
		maxBatchSize = 10
	}

	log.Println("multicallWorker started", chainId, networkName, "maxBatchSize", maxBatchSize, "batchInterval", c.BatchInterval)

	for {
		// Timer starts as soon as the first job is received.
		// If the timer finishes or the batch is full, the batch is sent.
		select {
		case job := <-jobChan:
			if len(jobs) == 0 {
				batchTimer.Reset(c.BatchInterval)
			}
			jobs = append(jobs, job)
			if len(jobs) < maxBatchSize {
				continue
			}
		case <-batchTimer.C:
		}
		batchTimer.Reset(1<<63 - 1)

		err := c.multicall(jobs, chainId, networkName)
		// Cancel all jobs if the multicall fails
		if err != nil {
			for _, job := range jobs {
				job.Result <- []byte{}
			}
		}
		jobs = jobs[:0]
	}
}

// Sends a batch of calls to the multicall contract
func (c *OnChainLoader) multicall(jobs []CallJob, chainId int, networkName types.NetworkName) (err error) {
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Sprintln("multicall panic", err)
		}
	}()

	if len(jobs) == 0 {
		return nil
	}

	inputs := make([]Call3InputType, len(jobs))
	for i, job := range jobs {
		input := Call3InputType{
			Target:       common.HexToAddress(string(job.Contract)),
			AllowFailure: true,
			CallData:     job.CallData,
		}
		inputs[i] = input
	}
	packed, err := c.multicallAbi.Pack("aggregate3", inputs)
	if err != nil {
		log.Fatal("failed to pack aggregate3", err)
	}

	addr := common.HexToAddress(c.Cfg[networkName].MulticallContract)
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: packed,
	}

	client, err := c.ethClientForChain(types.IntToChainId(chainId))
	if err != nil {
		return err
	}
	multicallResult, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return err
	}

	var results []Call3OutputType
	err = c.multicallAbi.UnpackIntoInterface(&results, "aggregate3", multicallResult)
	if err != nil {
		log.Fatal("failed to unpack aggregate3", err)
	}

	for i, job := range jobs {
		result := results[i]
		if result.Success {
			job.Result <- result.ReturnData
		} else {
			job.Result <- []byte{}
		}
	}
	return nil
}
