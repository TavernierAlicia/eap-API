package main

import (
	"github.com/gin-gonic/gin"
)

func getEtabs(c *gin.Context) {
	mail := c.Request.Header.Get("mail")

	if mail != "" {
		err, etabs := dbGetEtabs(mail)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "etabs not found",
			})
		} else {
			// etabs to json
			c.JSON(200, etabs)
		}

	} else {
		// send error code
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}
