package main

import (
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

		var randomData JSONTODATA
		c.BindJSON(&randomData)

		itemid := randomData.ItemID

		err = dbDeleteItem(itemid)
		if err != nil {
			ret503(c)
		} else {
			c.JSON(200, "Deleted")
		}

	}
}

func deleteCategory(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		var empty JSONTODATA
		var category JSONTODATA
		c.BindJSON(&category)

		if category.Category_id != 0 || category != empty {
			err = dbDeleteCategory(etabid, category.Category_id)
			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, gin.H{
					"message": "category deleted",
				})
			}
		} else {
			ret422(c)
		}
	}
}
