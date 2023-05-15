package server

import (
	"net/http"

	"github.com/CrocSwap/graphcache-go/views"
	"github.com/gin-gonic/gin"
)

type APIWebServer struct {
	Views views.IViews
	host  string
	port  int
}

func (s *APIWebServer) Serve() {
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/user_balances", s.queryUserBalance)
	r.Run()
}

func (s *APIWebServer) queryUserBalance(c *gin.Context) {
	chainId, _ := parseChainParam(c, "chainId")
	user, _ := parseAddrParam(c, "user")

	if chainId == "" || user == "" {
		wrapMissingParams(c)
		return
	}

	resp, err := s.Views.QueryUserBalances(chainId, user)
	wrapDataErrResp(c, resp, err)
}
