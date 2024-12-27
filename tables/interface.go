package tables

import (
	"encoding/hex"
	"strconv"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type ITable[Row any, SubGraphRow any] interface {
	GetID(r Row) string
	GetTime(r Row) int
	GetBlock(r Row) int
	ConvertSubGraphRow(SubGraphRow, string) Row
	SqlTableName() string
	ParseSubGraphResp(body []byte) ([]SubGraphRow, error)
	ParseSubGraphRespUnwrapped(body []byte) ([]SubGraphRow, error)
}

func parseNullableInt(s string) *int {
	if s == "" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		// Handle error
		return nil
	}
	return &val
}

func parseNullableFloat64(s string) *float64 {
	null := 0.0
	if s == "" {
		return &null
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Handle error
		return nil
	}
	return &val
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Handle error
		return 0
	}
	return val
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		// Handle error
		return 0
	}
	return val
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func hashPool(base string, quote string, poolIdx int) string {
	if base > quote {
		base, quote = quote, base
	}

	hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "uint256"},
		[]interface{}{base, quote, poolIdx})

	return "0x" + hex.EncodeToString(hash)
}
