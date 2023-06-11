package server

import (
	"net/http"

	"github.com/CrocSwap/graphcache-go/views"
	"github.com/gin-gonic/gin"
)

type APIWebServer struct {
	Views views.IViews
}

func (s *APIWebServer) Serve() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("gcgo/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("gcgo/user_balance_tokens", s.queryUserTokens)
	r.GET("gcgo/user_positions", s.queryUserPositions)
	r.GET("gcgo/pool_positions", s.queryPoolPositions)
	r.GET("gcgo/pool_position_apy_leaders", s.queryPoolPositions)
	r.GET("gcgo/user_pool_positions", s.queryUserPoolPositions)
	r.GET("gcgo/position_stats", s.querySinglePosition)
	r.GET("gcgo/user_limit_orders", s.queryUserLimits)
	r.GET("gcgo/pool_limit_orders", s.queryPoolLimits)
	r.GET("gcgo/user_pool_limit_orders", s.queryUserPoolLimits)
	r.GET("gcgo/limit_stats", s.querySingleLimit)
	r.GET("gcgo/user_txs", s.queryUserTxHist)
	r.GET("gcgo/pool_txs", s.queryPoolTxHist)
	r.GET("gcgo/pool_liq_curve", s.queryPoolLiqCurve)
	r.GET("gcgo/pool_stats", s.queryPoolStats)
	r.GET("gcgo/chain_stats", s.queryChainStats)
	r.Run()
}

func (s *APIWebServer) queryUserTokens(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserTokens(chainId, user)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserPositions(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserPositions(chainId, user)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserLimits(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserLimits(chainId, user)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserTxHist(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserTxHist(chainId, user, n)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolPositions(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolPositions(chainId, base, quote, poolIdx, n, true)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolPositionsApyLeaders(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolApyLeaders(chainId, base, quote, poolIdx, n, true)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolTxHist(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)
	time := parseIntOptional(c, "time", 0)
	period := parseIntOptional(c, "period", 3600)

	if len(c.Errors) > 0 {
		return
	}

	if time > 0 {
		resp := s.Views.QueryPoolTxHistFrom(chainId, base, quote, poolIdx, n, time, period)
		wrapDataErrResp(c, resp, nil)
	} else {
		resp := s.Views.QueryPoolTxHist(chainId, base, quote, poolIdx, n)
		wrapDataErrResp(c, resp, nil)
	}
}

func (s *APIWebServer) queryPoolLimits(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolLimits(chainId, base, quote, poolIdx, n)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserPoolPositions(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserPoolPositions(chainId, user, base, quote, poolIdx)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolLiqCurve(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolLiquidityCurve(chainId, base, quote, poolIdx)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolStats(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	histTime := parseIntOptional(c, "histTime", 0)

	if len(c.Errors) > 0 {
		return
	}

	if histTime > 0 {
		resp := s.Views.QueryPoolStatsFrom(chainId, base, quote, poolIdx, histTime)
		wrapDataErrResp(c, resp, nil)
	} else {
		resp := s.Views.QueryPoolStats(chainId, base, quote, poolIdx)
		wrapDataErrResp(c, resp, nil)
	}
}

func (s *APIWebServer) queryChainStats(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryChainStats(chainId, n)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserPoolLimits(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserPoolLimits(chainId, user, base, quote, poolIdx)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) querySinglePosition(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	bidTick := parseIntParam(c, "bidTick")
	askTick := parseIntParam(c, "askTick")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QuerySinglePosition(chainId, user, base, quote, poolIdx, bidTick, askTick)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) querySingleLimit(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	bidTick := parseIntParam(c, "bidTick")
	askTick := parseIntParam(c, "askTick")
	isBid := parseBoolParam(c, "isBid")
	pivotTime := parseIntParam(c, "pivotTime")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QuerySingleLimit(chainId, user, base, quote, poolIdx,
		bidTick, askTick, isBid, pivotTime)
	wrapDataErrResp(c, resp, nil)
}
