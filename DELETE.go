package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func deleteAllconn(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		err = dbResetAllConn(etabid)

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

func deleteItem(c *gin.Context) {
	_, err := checkAuth(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		itemid, err := strconv.ParseInt(c.Request.Header.Get("item_id"), 10, 64)

		if err != nil {
			c.JSON(422, gin.H{
				"message": "invalid entries",
			})
		} else {
			err = dbDeleteItem(itemid)
			if err != nil {
				c.JSON(503, gin.H{
					"message": "cannot delete item",
				})
			} else {
				c.JSON(200, "Deleted")
			}
		}

	}
}
