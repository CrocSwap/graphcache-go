package server

import (
	"errors"
	"strconv"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/gin-gonic/gin"
)

func parseAddrParam(c *gin.Context, paramName string) (types.EthAddress, error) {
	arg := c.Query(paramName)
	if arg == "" {
		return "", nil
	}

	parsed := types.ValidateEthAddr(arg)
	if parsed == "" {
		return "", errors.New("Invalid Ethereum address arg")
	}
	return parsed, nil
}

func parseChainParam(c *gin.Context, paramName string) (types.ChainId, error) {
	arg := c.Query(paramName)
	if arg == "" {
		return "", nil
	}

	parsed := types.ValidateChainId(arg)
	if parsed == "" {
		return "", errors.New("Invalid chainId arg")
	}
	return parsed, nil
}

func parseIntParam(c *gin.Context, paramName string) (*int, error) {
	arg := c.Query(paramName)
	if arg == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(arg)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseBoolParam(c *gin.Context, paramName string) (*bool, error) {
	arg := c.Query(paramName)
	if arg == "" {
		return nil, nil
	}

	result := false
	if arg == "true" {
		result = true
	}
	return &result, nil
}

func readBlockQueryArg(c *gin.Context) int {
	block, _ := parseBlockOptional(c.Query("block"))
	return block
}

func parseBlockOptional(blockArg string) (int, error) {
	if blockArg == "" {
		return 0, nil
	}

	block, err := strconv.Atoi(blockArg)
	if err != nil {
		return 0, err
	}
	return block, nil
}
