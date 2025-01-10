package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type responseProvenance struct {
	Hostname  string `json:"hostname"`
	ServeTime int    `json:"serveTime"`
}

type fullResponse struct {
	Data     any                `json:"data"`
	Metadata responseProvenance `json:"provenance"`
}

func wrapDataResp(c *gin.Context, result any) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "getHostnameError"
	}

	prov := responseProvenance{
		Hostname:  hostname,
		ServeTime: int(time.Now().UnixMilli()),
	}

	c.JSON(http.StatusOK, fullResponse{Data: result, Metadata: prov})
}

func wrapDataErrResp(c *gin.Context, result any, err error) {
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		wrapDataResp(c, result)
	}
}

func wrapErrResp(c *gin.Context, err error) {
	c.String(http.StatusInternalServerError, err.Error())
	c.Error(err)
}

func wrapErrMsg(c *gin.Context, err string) {
	wrapErrResp(c, fmt.Errorf(err))
}

func wrapErrMsgFmt(c *gin.Context, err string, a ...any) {
	wrapErrResp(c, fmt.Errorf(err+". ", a...))
}

func wrapMissingParam(c *gin.Context, paramName string) {
	wrapErrMsgFmt(c, "Missing parameter=%s", paramName)
}
