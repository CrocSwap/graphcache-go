package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseProvenance struct {
	hostname  string `json:"hostname"`
	serveTime int    `json:"serveTime"`
}

func wrapDataResp(c *gin.Context, result any) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "getHostnameError"
	}

	prov := ResponseProvenance{
		hostname:  hostname,
		serveTime: int(time.Now().UnixMilli()),
	}

	c.JSON(http.StatusOK, gin.H{"data": result, "provenance": prov})
}

func wrapDataErrResp(c *gin.Context, result any, err error) {
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		wrapDataResp(c, result)
	}
}

func wrapMissingParams(c *gin.Context) {
	c.String(http.StatusUnprocessableEntity, "Missing parameter")
}
