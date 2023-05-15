package types

import "math/big"

type UserBalance struct {
	token   EthAddress `json:"token"`
	balance big.Int    `json:"balance"`
}
