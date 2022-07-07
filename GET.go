package main

import (
	"strconv"
	"time"
	"fmt"

	eapCSV "github.com/TavernierAlicia/eap-CSV"
	eapFact "github.com/TavernierAlicia/eap-FACT"
	eapMail "github.com/TavernierAlicia/eap-MAIL"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func getEtabs(c *gin.Context) {

	mail := c.Query("mail")

	if mail != "" {
		etabs, err := dbGetEtabs(mail)

		if err != nil {
			ret404(c)
		} else {
			// etabs to json
			c.JSON(200, etabs)
		}

	} else {
		ret422(c)
	}
}


func getMenuCli(c *gin.Context) {
	etabToken := c.Param("etab")
	token := c.Request.Header.Get("Authorization")

	if token == "" {
		ret422(c)
	} else {
		etabid, err := dbCheckCliToken(etabToken)
		if err != nil {
			ret404(c)
		} else {
			err := dbCheckCliSess(token)

			if err != nil {
				ret401(c)
			} else {

				menu, err := dbGetEtabMenu(etabid)
				if err != nil {
					ret404(c)
				} else {
					c.JSON(200, menu)
				}
			}
		} 
	}
}


func getMenu(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")

	if token == "" {
		ret422(c)
	} else {
		etabid, err := dbCheckCliToken(token)

		if err != nil {
			etabid, err = dbCheckToken(token)
			if err != nil {
				ret401(c)
			
			} else {

					menu, err := dbGetEtabMenu(etabid)

					if err != nil {
						ret404(c)
					} else {
						c.JSON(200, menu)
					}
			}
		}
	}
}

func getPlanning(c *gin.Context) {
	etabToken := c.Param("etab")
	token := c.Request.Header.Get("Authorization")

	if token == "" {
		ret422(c)
	} else {
		etabid, err := dbCheckCliToken(etabToken)
		if err != nil {
			ret404(c)
		} else {
			err := dbCheckCliSess(token)

			if err != nil {
				ret401(c)
			} else {
				planning, err := dbGetPlanning(etabid)
				if err != nil {
					ret404(c)
				} else {
					c.JSON(200, planning)
				}
			}
		} 
	}
}


func getOrders(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")

	if token != "" {
		etabid, err := dbCheckToken(token)

		if err != nil {
			ret401(c)
		} else {
			orders, err := dbGetOrders(etabid)
			if err != nil {
				ret404(c)
			} else {
				c.JSON(200, orders)
			}
		}
	} else {
		ret422(c)
	}
}

func getOrder(c *gin.Context) {

	cli_uuid := c.Request.Header.Get("Authorization")
	token := c.Param("etab")

	orderid, err := strconv.ParseInt(c.Param("order_id"), 10, 64)

	if token != "" && orderid != 0 && err == nil {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			ret401(c)
		} else {
			// check cli_uuid
			err := dbCheckOrderCliSess(cli_uuid, orderid)

			if err != nil {
				ret404(c)
			} else {
				order, err := dbGetOrder(orderid)
				if err != nil {
					ret404(c)
				} else {
					c.JSON(200, order)
				}
			}
		}
	} else {
		ret422(c)
	}
}

func sendFact(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	cli_uuid := c.Query("cli_uuid")
	orderid, err := strconv.ParseInt(c.Query("order_id"), 10, 64)
	// mail := c.Query("mail")

	if token != "" && orderid != 0 && err == nil {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			ret404(c)
		} else {
			err := dbCheckOrderCliSess(cli_uuid, orderid)

			if err != nil {
				ret404(c)
			} else {
				// get fact link
				link, err := dbGetOrderFact(orderid)
				if err != nil {
					ret404(c)
				} else {
					// let's send this fact
					// err := eapMail.SendCliFact(link, mail)
					// if err != nil {
					// 	ret503(c)
					// } else {
					// 	c.JSON(200, "mail send")
					// }
					c.JSON(200, link)
				}
			}
		}
	} else {
		ret422(c)
	}
}

func getFactLink(c *gin.Context) {
	orderid, err := strconv.ParseInt(c.Query("order_id"), 10, 64)

	if orderid != 0 && err == nil {

		// get fact link
		link, err := dbGetOrderFact(orderid)
		if err != nil {
			ret404(c)
		} else {

			c.JSON(200, link)
		}
	} else {
		ret422(c)
	}
}

func getAllTickets(c *gin.Context) {
	etabid, err := checkAuth(c)
	datemin := c.Query("date_min")
	if datemin == "" {
		datemin = "1997-05-01 15:40:00"
	}
	datemax := c.Query("date_max")
	if datemax == "" {
		datemax = "3000-02-01 00:00:00"
	}

	if etabid != 0 && err == nil {

		// get fact link
		tickets, err := dbGetAllTickets(etabid, datemin, datemax)
		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, tickets)
		}
	} else {
		ret401(c)
	}
}

func getBossFact(c *gin.Context) {
	etabid, err := strconv.ParseInt(c.Query("etab_id"), 10, 64)

	if err == nil {
		// get etab infos
		etab, err := dbGetFactEtab(etabid)

		if err != nil {
			ret404(c)
		} else {

			etab.Fact_infos.Uuid = uuid.New().String()
			etab.Fact_infos.IsFirst = true
			etab.Fact_infos.Date = time.Now().Format("02-01-2006")
			etab.Fact_infos.Link = viper.GetString("links.cdn_fact") + etab.Fact_infos.Uuid + "_" + etab.Fact_infos.Date + ".pdf"
			err, etab.Fact_infos.Id = dbCreateBossFirstFact(etabid, etab.Fact_infos.Uuid, etab.Fact_infos.Link)

			if err != nil {
				ret503(c)
			} else {
				// create fact
				err = eapFact.CreateFact(etab)
				if err != nil {
					ret503(c)
				}
				// send fact
				err = eapMail.SendBossFact(etab)
				if err != nil {
					ret503(c)
				} else {
					c.JSON(200, etab)
				}
			}
		}
	} else {
		ret422(c)
	}
}

func getEtabParams(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		params, err := dbGetEtabParams(etabid)

		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, params)
		}
	}
}

func getQRs(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		qr0, qr1, err := dbGetQRs(etabid)
		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, gin.H{"qr0": qr0, "qr1": qr1})
		}
	}
}

func getProfile(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		profile, err := dbGetProfile(etabid)

		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, profile)
		}
	}

}

func getPaymentMethod(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		pay, err := dbGetPaymentMethods(etabid)

		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, pay)
		}
	}

}

func getEtabOffer(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		offer, err := dbGetOffer(etabid)

		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, offer)
		}
	}
}

func getCSV(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	}

	start := c.Query("start")
	end := c.Query("end")

	if err != nil || start == "" || end == "" {
		ret422(c)
	} else {
		content, err := eapCSV.DbGetCSVFacts(start, end, etabid)

		if err != nil {
			fmt.Println(err)
			ret404(c)
		} else {

			filepath, err := eapCSV.FactstoCSV(content, etabid, start, end)
			if err != nil {
				fmt.Println(err)
				ret404(c)
			} else {
				c.JSON(200, filepath)
			}
		}
	}
}

func getCategories(c *gin.Context) {

	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {

		categories, err := dbGetCategories(etabid)

		if err != nil {
			ret404(c)
		} else {
			c.JSON(200, categories)
		}
	}
}
