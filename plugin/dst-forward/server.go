package dstforward

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerServer() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hi")
	})
}
