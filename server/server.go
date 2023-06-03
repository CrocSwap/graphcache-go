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
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/user_balance_tokens", s.queryUserTokens)
	r.GET("/user_positions", s.queryUserPositions)
	r.GET("/pool_positions", s.queryPoolPositions)
	r.GET("/user_pool_positions", s.queryUserPoolPositions)
	r.GET("/position_stats", s.querySinglePosition)
	r.GET("/user_limit_orders", s.queryUserLimits)
	r.GET("/pool_limit_orders", s.queryPoolLimits)
	r.GET("/user_pool_limit_orders", s.queryUserPoolLimits)
	r.GET("/limit_stats", s.querySingleLimit)
	r.Run()
}

func (s *APIWebServer) queryUserTokens(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if user == "" {
		wrapMissingParams(c, "user")
		return
	}

	resp, err := s.Views.QueryUserTokens(chainId, user)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryUserPositions(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if user == "" {
		wrapMissingParams(c, "user")
		return
	}

	resp, err := s.Views.QueryUserPositions(chainId, user)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryUserLimits(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if user == "" {
		wrapMissingParams(c, "user")
		return
	}

	resp, err := s.Views.QueryUserLimits(chainId, user)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryPoolPositions(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")
	n, _ := parseIntParam(c, "n")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}
	if n == nil {
		wrapMissingParams(c, "n")
		return
	}

	resp, err := s.Views.QueryPoolPositions(chainId, base, quote, *poolIdx, *n, true)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryPoolLimits(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")
	n, _ := parseIntParam(c, "n")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}
	if n == nil {
		wrapMissingParams(c, "n")
		return
	}

	resp, err := s.Views.QueryPoolLimits(chainId, base, quote, *poolIdx, *n)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryUserPoolPositions(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if user == "" {
		wrapMissingParams(c, "user")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}

	resp, err := s.Views.QueryUserPoolPositions(chainId, user, base, quote, *poolIdx)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryUserPoolLimits(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if user == "" {
		wrapMissingParams(c, "user")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}

	resp, err := s.Views.QueryUserPoolLimits(chainId, user, base, quote, *poolIdx)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) querySinglePosition(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")
	bidTick, _ := parseIntParam(c, "bidTick")
	askTick, _ := parseIntParam(c, "askTick")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}
	if bidTick == nil {
		wrapMissingParams(c, "bidTick")
		return
	}
	if askTick == nil {
		wrapMissingParams(c, "askTick")
		return
	}

	resp, err := s.Views.QuerySinglePosition(chainId, user, base, quote, *poolIdx, *bidTick, *askTick)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) querySingleLimit(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")
	base, _ := parseAddrParam(c, "base")
	quote, _ := parseAddrParam(c, "quote")
	poolIdx, _ := parseIntParam(c, "poolIdx")
	bidTick, _ := parseIntParam(c, "bidTick")
	askTick, _ := parseIntParam(c, "askTick")
	isBid, _ := parseBoolParam(c, "isBid")
	pivotTime, _ := parseIntParam(c, "pivotTime")

	if chainId == "" {
		wrapMissingParams(c, "chainId")
		return
	}
	if base == "" {
		wrapMissingParams(c, "base")
		return
	}
	if quote == "" {
		wrapMissingParams(c, "quote")
		return
	}
	if poolIdx == nil {
		wrapMissingParams(c, "poolIdx")
		return
	}
	if bidTick == nil {
		wrapMissingParams(c, "bidTick")
		return
	}
	if askTick == nil {
		wrapMissingParams(c, "askTick")
		return
	}
	if isBid == nil {
		wrapMissingParams(c, "isBid")
		return
	}
	if pivotTime == nil {
		wrapMissingParams(c, "pivotTime")
		return
	}

	resp, err := s.Views.QuerySingleLimit(chainId, user, base, quote, *poolIdx,
		*bidTick, *askTick, *isBid, *pivotTime)
	wrapDataErrResp(c, resp, err)
}
