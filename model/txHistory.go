package model

import (
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type HistoryWriter struct {
	netCfg        loader.NetworkConfig
	commitEventFn func(types.PoolTxEvent)
}

func NewHistoryWriter(netCfg loader.NetworkConfig, commitFn func(types.PoolTxEvent)) *HistoryWriter {
	return &HistoryWriter{
		netCfg:        netCfg,
		commitEventFn: commitFn,
	}
}

func (h *HistoryWriter) CommitSwap(s tables.Swap) {
	h.commitEventFn(types.PoolTxEvent{
		EthTxHeader: types.EthTxHeader{
			BlockNum: s.Block,
			TxHash:   types.ValidateEthHash(s.TX),
			TxTime:   s.Time,
			User:     types.ValidateEthAddr(s.User),
		},

		PoolLocation: types.PoolLocation{
			ChainId: h.netCfg.RequireChainID(types.NetworkName(s.Network)),
			Base:    types.ValidateEthAddr(s.Base),
			Quote:   types.ValidateEthAddr(s.Quote),
			PoolIdx: s.PoolIdx,
		},

		PoolEventFlow: types.PoolEventFlow{
			BaseFlow:  s.BaseFlow,
			QuoteFlow: s.QuoteFlow,
		},

		PoolEventDescriptor: types.PoolEventDescriptor{
			EntityType:   tables.EntityTypeSwap,
			ChangeType:   tables.ChangeTypeSwap,
			PositionType: tables.PosTypeSwap,
		},

		PoolRangeFields: types.PoolRangeFields{
			IsBuy:     s.IsBuy > 0,
			InBaseQty: s.InBaseQty > 0,
		},
	})
}

func (h *HistoryWriter) CommitLiqChange(s tables.LiqChange) {
	baseFlow := float64(0)
	quoteFlow := float64(0)

	if s.BaseFlow != nil {
		baseFlow = *s.BaseFlow
	}
	if s.QuoteFlow != nil {
		quoteFlow = *s.QuoteFlow
	}

	entityType := tables.EntityTypeLiqChange
	if s.PositionType == tables.PosTypeKnockout {
		entityType = tables.EntityTypeLimit
	}

	h.commitEventFn(types.PoolTxEvent{
		EthTxHeader: types.EthTxHeader{
			BlockNum: s.Block,
			TxHash:   types.ValidateEthHash(s.TX),
			TxTime:   s.Time,
			User:     types.ValidateEthAddr(s.User),
		},

		PoolLocation: types.PoolLocation{
			ChainId: h.netCfg.RequireChainID(types.NetworkName(s.Network)),
			Base:    types.ValidateEthAddr(s.Base),
			Quote:   types.ValidateEthAddr(s.Quote),
			PoolIdx: s.PoolIdx,
		},

		PoolEventFlow: types.PoolEventFlow{
			BaseFlow:  baseFlow,
			QuoteFlow: quoteFlow,
		},

		PoolEventDescriptor: types.PoolEventDescriptor{
			EntityType:   entityType,
			ChangeType:   s.ChangeType,
			PositionType: s.PositionType,
		},

		PoolRangeFields: types.PoolRangeFields{
			IsBuy:     s.IsBid > 0,
			InBaseQty: s.IsBid > 0,
			BidTick:   s.BidTick,
			AskTick:   s.AskTick,
		},
	})
}
