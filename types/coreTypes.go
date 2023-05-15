package types

import "strings"

type ChainId string
type EthAddress string

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
