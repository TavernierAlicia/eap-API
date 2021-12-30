package main

import (
	"strconv"

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
			c.JSON(401, gin.H{
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
			// try same for boss
			etabid, err := checkToken(token)

			if err != nil {
				c.JSON(401, gin.H{
					"message": "no user for this token",
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

func GetOrders(c *gin.Context) {
	token := c.Request.Header.Get("token")

	if token != "" {
		etabid, err := checkToken(token)

		if err != nil {
			c.JSON(401, gin.H{
				"message": "no user for this token",
			})
		} else {
			orders, err := dbGetOrders(etabid)
			if err != nil {
				c.JSON(404, gin.H{
					"message": "planning not found",
				})
			} else {
				c.JSON(200, orders)
			}
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func GetOrder(c *gin.Context) {
	token := c.Request.Header.Get("token")
	orderid, err := strconv.ParseInt(c.Request.Header.Get("order_id"), 10, 64)
	cli_uuid := c.Request.Header.Get("cli_uuid")

	if token != "" && orderid != 0 && err == nil {
		// check cli token
		_, err := checkCliToken(token)

		if err != nil {
			c.JSON(401, gin.H{
				"message": "no user for this token",
			})
		} else {
			// check cli_uuid
			err := checkCliSess(cli_uuid, orderid)

			if err != nil {
				c.JSON(404, gin.H{
					"message": "no client with this id",
				})
			} else {
				order, err := dbGetOrder(orderid)
				if err != nil {
					c.JSON(404, gin.H{
						"message": "planning not found",
					})
				} else {
					c.JSON(200, order)
				}
			}
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}
