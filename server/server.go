package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/views"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type APIWebServer struct {
	Views views.IViews
}

func (s *APIWebServer) Serve(basePrefix string, listenAddr string, extendedApi bool) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	for _, prefix := range []string{basePrefix, basePrefix + "-canary"} {
		r.GET(prefix+"/", func(c *gin.Context) { c.Status(http.StatusOK) })
		r.GET(prefix+"/user_balance_tokens", s.queryUserTokens)
		r.GET(prefix+"/user_positions", s.queryUserPositions)
		r.GET(prefix+"/pool_positions", s.queryPoolPositions)
		r.GET(prefix+"/pool_position_apy_leaders", s.queryPoolPositionsApyLeaders)
		r.GET(prefix+"/user_pool_positions", s.queryUserPoolPositions)
		r.GET(prefix+"/position_stats", s.querySinglePosition)
		r.GET(prefix+"/user_limit_orders", s.queryUserLimits)
		r.GET(prefix+"/pool_limit_orders", s.queryPoolLimits)
		r.GET(prefix+"/user_pool_limit_orders", s.queryUserPoolLimits)
		r.GET(prefix+"/user_pool_txs", s.queryUserPoolTxHist)
		r.GET(prefix+"/limit_stats", s.querySingleLimit)
		r.GET(prefix+"/user_txs", s.queryUserTxHist)
		r.GET(prefix+"/pool_txs", s.queryPoolTxHist)
		r.GET(prefix+"/pool_liq_curve", s.queryPoolLiqCurve)
		r.GET(prefix+"/pool_stats", s.queryPoolStats)
		r.GET(prefix+"/all_pool_stats", s.queryAllPoolStats)
		r.GET(prefix+"/pool_candles", s.queryPoolCandles)
		r.GET(prefix+"/pool_list", s.queryPoolList)
		r.GET(prefix+"/chain_stats", s.queryChainStats)
		r.GET(prefix+"/plume_task", s.queryPlumeTask)
		if extendedApi {
			r.GET(prefix+"/historic_positions", s.queryHistoricPositions)
		}
	}

	log.Println("API Serving at", listenAddr+"/"+basePrefix)
	r.Run(listenAddr)
}

func (s *APIWebServer) queryUserTokens(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserTokens(chainId, user)
	c.Header("Cache-Control", "public, max-age=60")
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
	afterTime, beforeTime := getTimeParameters(c)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserTxHist(chainId, user, n, afterTime, beforeTime)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolPositions(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)
	omitEmpty := parseBoolOptional(c, "omitEmpty", false)
	afterTime, beforeTime := getTimeParameters(c)
	if len(c.Errors) > 0 {
		return
	}

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolPositions(chainId, base, quote, poolIdx, n, omitEmpty, afterTime, beforeTime)
	c.Header("Cache-Control", "public, max-age=5")
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
	c.Header("Cache-Control", "public, max-age=60")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolTxHist(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)
	afterTime, beforeTime := getTimeParameters(c)
	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolTxHist(chainId, base, quote, poolIdx, n, afterTime, beforeTime)
	c.Header("Cache-Control", "public, max-age=5")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryHistoricPositions(c *gin.Context) {
	// time of the liquidity snapshot
	time := parseIntParam(c, "time")
	// all pool filters are optional
	chainId := types.ValidateChainId(c.Query("chainId"))
	base := types.ValidateEthAddr(c.Query("base"))
	quote := types.ValidateEthAddr(c.Query("quote"))
	poolIdx, _ := strconv.Atoi(c.Query("poolIdx"))
	user := types.ValidateEthAddr(c.Query("user"))
	omitEmpty := parseBoolOptional(c, "omitEmpty", true)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryHistoricPositions(chainId, base, quote, poolIdx, time, user, omitEmpty)
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolLimits(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)
	afterTime, beforeTime := getTimeParameters(c)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolLimits(chainId, base, quote, poolIdx, n, afterTime, beforeTime)
	c.Header("Cache-Control", "public, max-age=10")
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
	c.Header("Cache-Control", "public, max-age=60")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolStats(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	histTime := parseIntOptional(c, "histTime", 0)
	with24hPrices := parseBoolOptional(c, "with24hPrices", false)

	if histTime > 0 && with24hPrices {
		wrapErrMsg(c, "Cannot specify both histTime and with24hPrices")
	}

	if len(c.Errors) > 0 {
		return
	}

	if histTime > 0 {
		resp := s.Views.QueryPoolStatsFrom(chainId, base, quote, poolIdx, histTime)
		if int64(histTime) < time.Now().Unix()-60 {
			c.Header("Cache-Control", "public, max-age=60")
		} else {
			c.Header("Cache-Control", "public, max-age=5")
		}
		wrapDataErrResp(c, resp, nil)
	} else {
		resp := s.Views.QueryPoolStats(chainId, base, quote, poolIdx, with24hPrices)
		c.Header("Cache-Control", "public, max-age=5")
		wrapDataErrResp(c, resp, nil)
	}
}

func (s *APIWebServer) queryPoolCandles(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	period := parseIntParam(c, "period")
	timeParam := parseIntOptional(c, "time", 0)
	n := parseIntMaxParam(c, "n", 3000)

	if period > 3600 {
		if period%3600 != 0 {
			wrapErrMsg(c, "Period over 3600 must be a multiple of 3600")
		}
	}
	if period > 604800 {
		wrapErrMsg(c, "Period must be less than 604800")
	}

	if len(c.Errors) > 0 {
		return
	}

	timeRange := views.CandleRangeArgs{
		N:         n,
		Period:    period,
		StartTime: nil,
	}
	if timeParam > 0 {
		timeRange.StartTime = &timeParam
	}

	resp := s.Views.QueryPoolCandles(chainId, base, quote, poolIdx, timeRange)
	c.Header("Cache-Control", "public, max-age=10")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryPoolList(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryPoolSet(chainId)
	c.Header("Cache-Control", "public, max-age=60")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryAllPoolStats(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	histTime := parseIntOptional(c, "histTime", 0)
	with24hPrices := parseBoolOptional(c, "with24hPrices", false)

	if histTime > 0 && with24hPrices {
		wrapErrMsg(c, "Cannot specify both histTime and with24hPrices")
	}

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryAllPoolStats(chainId, histTime, with24hPrices)
	c.Header("Cache-Control", "public, max-age=60")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryChainStats(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	n := parseIntMaxParam(c, "n", 200)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryChainStats(chainId, n)
	c.Header("Cache-Control", "public, max-age=60")
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
	c.Header("Cache-Control", "public, max-age=60")
	wrapDataErrResp(c, resp, nil)
}

func (s *APIWebServer) queryUserPoolTxHist(c *gin.Context) {
	chainId := parseChainParam(c, "chainId")
	user := parseAddrParam(c, "user")
	base := parseAddrParam(c, "base")
	quote := parseAddrParam(c, "quote")
	poolIdx := parseIntParam(c, "poolIdx")
	n := parseIntMaxParam(c, "n", 200)
	var err error
	nStr := c.Query("n")
	if nStr != "" {
		n, err = strconv.Atoi(nStr)
		if err != nil {
			wrapErrMsg(c, "Invalid int arg="+nStr)
		}
	}
	afterTime, beforeTime := getTimeParameters(c)

	if len(c.Errors) > 0 {
		return
	}

	resp := s.Views.QueryUserPoolTxHist(chainId, user, base, quote, poolIdx, n, afterTime, beforeTime)
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

func getTimeParameters(c *gin.Context) (afterTime int, beforeTime int) {
	afterTime = parseIntOptional(c, "time", 0)
	beforeTime = parseIntOptional(c, "timeBefore", 0)
	period := parseIntOptional(c, "period", 0)

	if afterTime > 0 && beforeTime == 0 && period > 0 {
		beforeTime = afterTime + period
	}
	if beforeTime > 0 && afterTime == 0 && period > 0 {
		afterTime = beforeTime - period
	}

	if afterTime > 0 && beforeTime > 0 && afterTime > beforeTime {
		wrapErrMsg(c, "afterTime must be less than beforeTime")
		return
	}
	if afterTime > 0 && beforeTime == 0 {
		beforeTime = 1999999999
	}
	return
}

func (s *APIWebServer) queryPlumeTask(c *gin.Context) {
	task := c.Query("task")
	user := types.ValidateEthAddr(c.Query("address"))
	if user == "" {
		resp := views.PlumeTaskStatus{
			Error: "Invalid Ethereum address",
			Code:  1,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	resp := s.Views.QueryPlumeUserTask(user, task)
	c.JSON(http.StatusOK, resp)
}
