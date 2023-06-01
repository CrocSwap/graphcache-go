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
	r.Run()
}

func (s *APIWebServer) queryUserTokens(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" || user == "" {
		wrapMissingParams(c)
		return
	}

	resp, err := s.Views.QueryUserTokens(chainId, user)
	wrapDataErrResp(c, resp, err)
}

func (s *APIWebServer) queryUserPositions(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" || user == "" {
		wrapMissingParams(c)
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

	if chainId == "" || base == "" || quote == "" || poolIdx < 0 || n < 0 {
		wrapMissingParams(c)
		return
	}

	resp, err := s.Views.QueryPoolPositions(chainId, base, quote, poolIdx, n)
	wrapDataErrResp(c, resp, err)
}
