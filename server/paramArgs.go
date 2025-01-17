package server

import (
	"os"
	"strconv"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/gin-gonic/gin"
)

func parseAddrParam(c *gin.Context, paramName string) types.EthAddress {
	arg := c.Query(paramName)
	if arg == "" {
		wrapMissingParam(c, paramName)
		return ""
	}

	parsed := types.ValidateEthAddr(arg)
	if parsed == "" {
		wrapErrMsgFmt(c, "Invalid Ethereum address arg=%s", arg)
		return ""
	}

	return parsed
}

func parseChainParam(c *gin.Context, paramName string) types.ChainId {
	arg := c.Query(paramName)
	if arg == "" {
		wrapMissingParam(c, paramName)
		return ""
	}

	parsed := types.ValidateChainId(arg)
	if parsed == "" {
		wrapErrMsgFmt(c, "Invalid ChainID arg=%s", paramName)
	}
	return parsed
}

func parseIntParam(c *gin.Context, paramName string) int {
	arg := c.Query(paramName)
	if arg == "" {
		wrapMissingParam(c, paramName)
		return -1
	}

	parsed, err := strconv.Atoi(arg)
	if err != nil {
		wrapErrMsgFmt(c, "Invalid int arg=%s", arg)
		return -1
	}

	return parsed
}

func parseIntOptional(c *gin.Context, paramName string, dflt int) int {
	arg := c.Query(paramName)
	if arg == "" {
		return dflt
	}

	parsed, err := strconv.Atoi(arg)
	if err != nil {
		wrapErrMsgFmt(c, "Invalid int arg=%s", arg)
		return -1
	}

	return parsed
}

func parseIntMaxParam(c *gin.Context, paramName string, maxSize int) int {
	parsed := parseIntParam(c, paramName)
	// If the unrestricted password is set and the unr param is set to it, then we allow the max size to be exceeded
	unrestricted := (os.Getenv("UNRESTRICT_PASSWORD") != "" && c.Query("unr") == os.Getenv("UNRESTRICT_PASSWORD"))

	if parsed > maxSize && !unrestricted {
		wrapErrMsgFmt(c, "n Exceeds max size of %d", maxSize)
	}

	return parsed
}

func parseBoolParam(c *gin.Context, paramName string) bool {
	arg := c.Query(paramName)
	if arg == "" {
		wrapMissingParam(c, paramName)
		return false
	}

	result := false
	if arg == "true" {
		result = true
	}
	return result
}

func parseBoolOptional(c *gin.Context, paramName string, dflt bool) bool {
	arg := c.Query(paramName)
	if arg == "" {
		return dflt
	} else if arg == "true" {
		return true
	} else {
		return false
	}
}
