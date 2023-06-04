package types

import (
	"log"
	"strconv"
	"strings"
)

type ChainId string
type EthAddress string
type EthTxHash string
type EthStorageHash string
type NetworkName string

func ValidateEthAddr(arg string) EthAddress {
	if strings.HasPrefix(arg, "0x") && len(arg) == 42 {
		return EthAddress(strings.ToLower(arg))
	} else if len(arg) == 40 {
		return EthAddress("0x" + strings.ToLower((arg)))
	}
	return ""
}

func ValidateEthHash(arg string) EthTxHash {
	if strings.HasPrefix(arg, "0x") && len(arg) == 66 {
		return EthTxHash(strings.ToLower(arg))
	} else if len(arg) == 64 {
		return EthTxHash("0x" + strings.ToLower((arg)))
	}
	return ""
}

func RequireEthAddr(arg string) EthAddress {
	result := ValidateEthAddr(arg)
	if result == "" {
		log.Fatal(result)
	}
	return result
}

// Arbitrary, but a chain ID hex address should never be more than 8
// digits (plus leading 0x).
const MAX_CHAIN_LENGTH = 12

func ValidateChainId(arg string) ChainId {
	if strings.HasPrefix(arg, "0x") && len(arg) < MAX_CHAIN_LENGTH {
		return ChainId(strings.ToLower(arg))
	}
	return ""
}

func intToHex(num int) string {
	return "0x" + strings.ToLower(strconv.FormatInt(int64(num), 16))
}

func IntToChainId(num int) ChainId {
	return ChainId(intToHex(num))
}
