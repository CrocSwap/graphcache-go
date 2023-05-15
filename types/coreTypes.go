package types

import (
	"strconv"
	"strings"
)

type ChainId string
type EthAddress string
type NetworkName string

func ValidateEthAddr(arg string) EthAddress {
	if strings.HasPrefix(arg, "0x") && len(arg) == 42 {
		return EthAddress(strings.ToLower(arg))
	}
	return ""
}

func ValidateChainId(arg string) ChainId {
	if strings.HasPrefix(arg, "0x") {
		return ChainId(strings.ToLower(arg))
	}
	return ""
}

func intToHex(num int) string {
	return strings.ToLower(strconv.FormatInt(int64(num), 16))
}

func IntToChainId(num int) ChainId {
	return ChainId(intToHex(num))
}
