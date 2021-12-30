package main

import (
	"github.com/gin-gonic/gin"
)

func updateOrderStatus(c *gin.Context) {
	var details OrderDetails
	var checkDetails OrderDetails

	c.BindJSON(&details)

	if details != checkDetails && details.Token != "" {

		// check if client
		// it's a boss or server
		_, err := checkToken(details.Token)
		if err != nil {
			c.JSON(401, gin.H{
				"message": "no user for this token",
			})
		} else {
			err := dbUpdateOrderStatus(details)
			if err != nil {
				c.JSON(404, gin.H{
					"message": "order update failed",
				})
			} else {
				c.JSON(200, gin.H{
					"message": "updated",
				})
			}
		}

	} else {
		// send error code
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}
