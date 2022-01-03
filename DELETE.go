package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func deleteAllconn(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		err = dbResetAllConn(etabid)

		if err != nil {
			ret503(c)

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
			ret503(c)

		} else {
			c.JSON(200, gin.H{
				"message": "disconnected",
			})
		}
	} else {
		ret401(c)
	}
}

func deleteItem(c *gin.Context) {
	_, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		itemid, err := strconv.ParseInt(c.Request.Header.Get("item_id"), 10, 64)

		if err != nil {
			ret422(c)
		} else {
			err = dbDeleteItem(itemid)
			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, "Deleted")
			}
		}

	}
}
