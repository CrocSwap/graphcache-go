package types

import "math/big"

type UserBalance struct {
	Token   EthAddress `json:"token"`
	Balance big.Int    `json:"balance"`
}

type TokenMetadata struct {
	Decimals int    `json:"decimals"`
	Symbol   string `json:"symbol"`
}

type TokenPairMetadata struct {
	BaseDecimals  int    `json:"baseDecimals"`
	BaseSymbol    string `json:"baseSymbol"`
	QuoteDecimals int    `json:"quoteDecimals"`
	QuoteSymbol   string `json:"quoteSymbol"`
}

func PairTokenMetadata(base TokenMetadata, quote TokenMetadata) TokenPairMetadata {
	return TokenPairMetadata{
		BaseSymbol:    base.Symbol,
		BaseDecimals:  base.Decimals,
		QuoteSymbol:   quote.Symbol,
		QuoteDecimals: quote.Decimals,
	}
}
