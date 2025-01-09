package tables

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

type LiqChangeTable struct{}

func (tbl LiqChangeTable) GetID(r LiqChange) string {
	return r.ID
}

func (tbl LiqChangeTable) GetTime(r LiqChange) int {
	return r.Time
}

func (tbl LiqChangeTable) GetBlock(r LiqChange) int {
	return r.Block
}

type PosType int8

const (
	PosTypeUnknown PosType = iota
	PosTypeAmbient
	PosTypeConcentrated
	PosTypeKnockout
	PosTypeSwap
)

var posTypeMap = map[string]PosType{
	"unknown":      PosTypeUnknown,
	"ambient":      PosTypeAmbient,
	"concentrated": PosTypeConcentrated,
	"knockout":     PosTypeKnockout,
	"swap":         PosTypeSwap,
}

var posTypeStringMap = map[PosType]string{
	PosTypeUnknown:      "unknown",
	PosTypeAmbient:      "ambient",
	PosTypeConcentrated: "concentrated",
	PosTypeKnockout:     "knockout",
	PosTypeSwap:         "swap",
}

func (p *PosType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*p = posTypeMap[s]
	return nil
}

func (p PosType) MarshalJSON() ([]byte, error) {
	return json.Marshal(posTypeStringMap[p])
}

type ChangeType int8

const (
	ChangeTypeUnknown ChangeType = iota
	ChangeTypeMint
	ChangeTypeBurn
	ChangeTypeCross
	ChangeTypeRecover
	ChangeTypeClaim
	ChangeTypeHarvest
	ChangeTypeSwap
)

var changeTypeMap = map[string]ChangeType{
	"unknown": ChangeTypeUnknown,
	"mint":    ChangeTypeMint,
	"burn":    ChangeTypeBurn,
	"cross":   ChangeTypeCross,
	"recover": ChangeTypeRecover,
	"claim":   ChangeTypeClaim,
	"harvest": ChangeTypeHarvest,
	"swap":    ChangeTypeSwap,
}

var changeTypeStringMap = map[ChangeType]string{
	ChangeTypeUnknown: "unknown",
	ChangeTypeMint:    "mint",
	ChangeTypeBurn:    "burn",
	ChangeTypeCross:   "cross",
	ChangeTypeRecover: "recover",
	ChangeTypeClaim:   "claim",
	ChangeTypeHarvest: "harvest",
	ChangeTypeSwap:    "swap",
}

func (c *ChangeType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*c = changeTypeMap[s]
	return nil
}

func (c ChangeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(changeTypeStringMap[c])
}

type EntityType int8

const (
	EntityTypeUnknown EntityType = iota
	EntityTypeSwap
	EntityTypeLimit
	EntityTypeLiqChange
)

var entityTypeMap = map[string]EntityType{
	"unknown":    EntityTypeUnknown,
	"swap":       EntityTypeSwap,
	"limitOrder": EntityTypeLimit,
	"liqchange":  EntityTypeLiqChange,
}

var entityTypeStringMap = map[EntityType]string{
	EntityTypeUnknown:   "unknown",
	EntityTypeSwap:      "swap",
	EntityTypeLimit:     "limitOrder",
	EntityTypeLiqChange: "liqchange",
}

func (e *EntityType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*e = entityTypeMap[s]
	return nil
}

func (e EntityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityTypeStringMap[e])
}

type LiqChange struct {
	ID           string     `json:"id" db:"id"`
	CallIndex    int        `json:"callIndex" db:"callIndex"`
	Network      string     `json:"network" db:"network"`
	TX           string     `json:"tx" db:"tx"`
	Base         string     `json:"base" db:"base"`
	Quote        string     `json:"quote" db:"quote"`
	PoolIdx      int        `json:"poolIdx" db:"poolIdx"`
	PoolHash     string     `json:"poolHash" db:"poolHash"`
	User         string     `json:"user" db:"user"`
	Block        int        `json:"block" db:"block"`
	Time         int        `json:"time" db:"time"`
	PositionType PosType    `json:"positionType" db:"positionType"`
	ChangeType   ChangeType `json:"changeType" db:"changeType"`
	BidTick      int        `json:"bidTick" db:"bidTick"`
	AskTick      int        `json:"askTick" db:"askTick"`
	IsBid        int        `json:"isBid" db:"isBid"`
	Liq          *float64   `json:"liq" db:"liq"`
	BaseFlow     *float64   `json:"baseFlow" db:"baseFlow"`
	QuoteFlow    *float64   `json:"quoteFlow" db:"quoteFlow"`
	Source       string     `json:"source" db:"source"`
	PivotTime    *int       `json:"pivotTime" db:"pivotTime"`
}

type LiqChangeSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	CallIndex       int    `json:"callIndex"`
	User            string `json:"user"`
	Pool            struct {
		Base    string `json:"base"`
		Quote   string `json:"quote"`
		PoolIdx string `json:"poolIdx"`
	} `json:"pool"`
	Block        string     `json:"block"`
	Time         string     `json:"time"`
	PositionType PosType    `json:"positionType"`
	ChangeType   ChangeType `json:"changeType"`
	BidTick      int        `json:"bidTick"`
	AskTick      int        `json:"askTick"`
	IsBid        bool       `json:"isBid"`
	Liq          string     `json:"liq"`
	BaseFlow     string     `json:"baseFlow"`
	QuoteFlow    string     `json:"quoteFlow"`
	PivotTime    string     `json:"pivotTime"`
}
type LiqChangeSubGraphData struct {
	LiqChanges []LiqChangeSubGraph `json:"liquidityChanges"`
}

type LiqChangeSubGraphResp struct {
	Data LiqChangeSubGraphData `json:"data"`
}

func (tbl LiqChangeTable) ConvertSubGraphRow(r LiqChangeSubGraph, network string) LiqChange {
	base, quote := r.Pool.Base, r.Pool.Quote
	baseFlow := parseNullableFloat64(r.BaseFlow)
	quoteFlow := parseNullableFloat64(r.QuoteFlow)

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
		baseFlow, quoteFlow = quoteFlow, baseFlow
	}

	return LiqChange{
		ID:           network + r.ID,
		CallIndex:    r.CallIndex,
		Network:      network,
		TX:           r.TransactionHash,
		Base:         base,
		Quote:        quote,
		PoolIdx:      parseInt(r.Pool.PoolIdx),
		PoolHash:     hashPool(base, quote, parseInt(r.Pool.PoolIdx)),
		User:         translateUser(r.User),
		Block:        parseInt(r.Block),
		Time:         parseInt(r.Time),
		PositionType: r.PositionType,
		ChangeType:   r.ChangeType,
		BidTick:      r.BidTick,
		AskTick:      r.AskTick,
		IsBid:        boolToInt(r.IsBid),
		Liq:          parseNullableFloat64(r.Liq),
		BaseFlow:     baseFlow,
		QuoteFlow:    quoteFlow,
		Source:       "graph",
		PivotTime:    parseNullableInt(r.PivotTime),
	}
}

func (tbl LiqChangeTable) SqlTableName() string { return "liqchanges" }

func (tbl LiqChangeTable) ReadSqlRow(rows *sql.Rows) LiqChange {
	var liqChange LiqChange
	err := rows.Scan(
		&liqChange.ID,
		&liqChange.CallIndex,
		&liqChange.Network,
		&liqChange.TX,
		&liqChange.Base,
		&liqChange.Quote,
		&liqChange.PoolIdx,
		&liqChange.PoolHash,
		&liqChange.User,
		&liqChange.Block,
		&liqChange.Time,
		&liqChange.PositionType,
		&liqChange.ChangeType,
		&liqChange.BidTick,
		&liqChange.AskTick,
		&liqChange.IsBid,
		&liqChange.Liq,
		&liqChange.BaseFlow,
		&liqChange.QuoteFlow,
		&liqChange.Source,
		&liqChange.PivotTime,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return liqChange
}

func (tbl LiqChangeTable) ParseSubGraphResp(body []byte) ([]LiqChangeSubGraph, error) {
	var parsed LiqChangeSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]LiqChangeSubGraph, 0)
	for _, entry := range parsed.Data.LiqChanges {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl LiqChangeTable) ParseSubGraphRespUnwrapped(body []byte) ([]LiqChangeSubGraph, error) {
	var parsed LiqChangeSubGraphData

	err := json.Unmarshal(body, &parsed.LiqChanges)
	if err != nil {
		return nil, err
	}

	ret := make([]LiqChangeSubGraph, 0)
	for _, entry := range parsed.LiqChanges {
		ret = append(ret, entry)
	}
	return ret, nil
}
