package main

import (
	"fmt"
	"strconv"
	"time"

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
		etabid, err := dbCheckCliToken(token)

		if err != nil {
			c.JSON(401, gin.H{
				"message": "no QR for this token",
			})
		} else {
			err := dbInsertCliSess(clientUuid)

			if err != nil {
				c.JSON(404, gin.H{
					"message": "cli insertion failed",
				})
			} else {
				menu, err := dbGetEtabMenu(etabid)

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
		etabid, err := dbCheckCliToken(token)
		if err != nil {
			// try same for boss
			etabid, err := dbCheckToken(token)

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

func getOrders(c *gin.Context) {
	token := c.Request.Header.Get("token")

	if token != "" {
		etabid, err := dbCheckToken(token)

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

func getOrder(c *gin.Context) {
	token := c.Request.Header.Get("token")
	orderid, err := strconv.ParseInt(c.Request.Header.Get("order_id"), 10, 64)
	cli_uuid := c.Request.Header.Get("cli_uuid")

	if token != "" && orderid != 0 && err == nil {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			c.JSON(401, gin.H{
				"message": "no user for this token",
			})
		} else {
			// check cli_uuid
			err := dbCheckCliSess(cli_uuid, orderid)

			if err != nil {
				c.JSON(404, gin.H{
					"message": "no client with this id",
				})
			} else {
				order, err := dbGetOrder(orderid)
				if err != nil {
					c.JSON(404, gin.H{
						"message": "order not found",
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

func sendFact(c *gin.Context) {
	token := c.Request.Header.Get("token")
	orderid, err := strconv.ParseInt(c.Request.Header.Get("order_id"), 10, 64)
	cli_uuid := c.Request.Header.Get("cli_uuid")
	mail := c.Request.Header.Get("mail")

	if token != "" && orderid != 0 && err == nil && mail != "" {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "no QR with this token",
			})
		} else {
			err := dbCheckCliSess(cli_uuid, orderid)

			if err != nil {
				c.JSON(404, gin.H{
					"message": "no client with this id",
				})
			} else {
				// get fact link
				link, err := dbGetOrderFact(orderid)
				if err != nil {
					c.JSON(404, gin.H{
						"message": "no fact found",
					})
				} else {
					// let's send this fact
					fmt.Println("ready to send " + link)
					err := sendCliFact(link, mail)
					if err != nil {
						c.JSON(500, gin.H{
							"message": "cannot send mail",
						})
					} else {
						c.JSON(200, "mail send")
					}
				}
			}
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func getFactLink(c *gin.Context) {
	orderid, err := strconv.ParseInt(c.Request.Header.Get("order_id"), 10, 64)

	if orderid != 0 && err == nil {

		// get fact link
		link, err := dbGetOrderFact(orderid)
		if err != nil {
			c.JSON(404, gin.H{
				"message": "no fact found",
			})
		} else {

			c.JSON(200, link)
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func getBossFact(c *gin.Context) {
	etabid, err := strconv.ParseInt(c.Request.Header.Get("etab_id"), 10, 64)

	if err == nil {
		// get etab infos
		etab, err := dbGetFactEtab(etabid)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "offer not found",
			})
		} else {

			// TODO: generate fact
			etab.Fact_infos.Link = "../tests/zpl.pdf"
			etab.Fact_infos.Date = time.Now().Format("02-01-2006")

			// send fact
			sendBossFact(etab)
		}

	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func getEtabParams(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		params, err := dbGetEtabParams(etabid)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "params not found",
			})
		} else {
			c.JSON(200, params)
		}
	}
}

func getProfile(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		profile, err := dbGetProfile(etabid)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "profile not found",
			})
		} else {
			c.JSON(200, profile)
		}
	}

}

func getPaymentMethod(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		pay, err := dbGetPaymentMethods(etabid)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "payment method not found",
			})
		} else {
			c.JSON(200, pay)
		}
	}

}

func getEtabOffer(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		offer, err := dbGetOffer(etabid)

		if err != nil {
			c.JSON(404, gin.H{
				"message": "offer not found",
			})
		} else {
			c.JSON(200, offer)
		}
	}
}
