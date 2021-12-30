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

func getMenu(c *gin.Context) {
	token := c.Request.Header.Get("token")
	clientUuid := c.Request.Header.Get("client-uuid")

	if token == "" || clientUuid == "" {
		// send error code
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	} else {
		etabid, err := checkCliToken(token)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "no QR for this token",
			})
		} else {
			err := insertCliSess(clientUuid)

			if err != nil {
				c.JSON(404, gin.H{
					"message": "cli insertion failed",
				})
			} else {
				menu, err := getEtabMenu(etabid)

				if err != nil {
					c.JSON(404, gin.H{
						"message": "menu not found",
					})
				} else {
					c.JSON(200, menu)
				}
			}
		}
	}
}

func getPlanning(c *gin.Context) {
	token := c.Request.Header.Get("token")

	if token != "" {
		// check token && get etabid
		etabid, err := checkCliToken(token)
		if err != nil {
			c.JSON(404, gin.H{
				"message": "no QR for this token",
			})
		} else {
			planning, err := dbGetPlanning(etabid)
			if err != nil {
				c.JSON(404, gin.H{
					"message": "planning not found",
				})
			} else {
				c.JSON(200, planning)
			}
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}
