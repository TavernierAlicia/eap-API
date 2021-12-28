package main

import (
	"github.com/gin-gonic/gin"
)

func deleteAllconn(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		err = ResetAllConn(etabid)

		if err != nil {
			c.JSON(503, gin.H{
				"message": "delete all conns failed",
			})
		} else {
			c.JSON(200, gin.H{
				"message": "connections deleted",
			})
		}
	}

}

func disconnect(c *gin.Context) {
	_, err := checkAuth(c)

	auth := c.Request.Header.Get("Authorization")

	if auth != "" && err == nil {
		err = dbDisconnect(auth)

		if err != nil {
			c.JSON(503, gin.H{
				"message": "disconnect failed",
			})
		} else {
			c.JSON(200, gin.H{
				"message": "disconnected",
			})
		}
	} else {
		c.JSON(401, gin.H{
			"message": "auth required",
		})
	}
}
