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
	r.GET("/pool_positions", s.QueryPoolPositions)
	r.GET("/position_stats", s.QuerySinglePosition)
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

func (s *APIWebServer) QueryPoolPositions(c *gin.Context) {
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

	resp, err := s.Views.QueryPoolPositions(chainId, base, quote, *poolIdx, *n)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) QuerySinglePosition(c *gin.Context) {
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
