package main

import (
	"strconv"
	"time"

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

func getMenu(c *gin.Context) {
	token := c.Request.Header.Get("token")
	clientUuid := c.Query("cli_uuid")

	if token == "" || clientUuid == "" {
		// send error code
		ret422(c)
	} else {
		etabid, err := dbCheckCliToken(token)

		if err != nil {
			ret401(c)
		} else {
			err := dbInsertCliSess(clientUuid)

			if err != nil {
				ret404(c)
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
	token := c.Request.Header.Get("token")

	if token != "" {
		// check token && get etabid
		etabid, err := dbCheckCliToken(token)
		if err != nil {
			// try same for boss
			etabid, err := dbCheckToken(token)

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
		} else {
			planning, err := dbGetPlanning(etabid)
			if err != nil {
				ret404(c)
			} else {
				c.JSON(200, planning)
			}
		}
	} else {
		ret422(c)
	}
}

func getOrders(c *gin.Context) {
	token := c.Request.Header.Get("token")

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
	token := c.Request.Header.Get("token")
	cli_uuid := c.Query("cli_uuid")
	orderid, err := strconv.ParseInt(c.Query("order_id"), 10, 64)

	if token != "" && orderid != 0 && err == nil {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			ret401(c)
		} else {
			// check cli_uuid
			err := dbCheckCliSess(cli_uuid, orderid)

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
	token := c.Request.Header.Get("token")
	cli_uuid := c.Query("cli_uuid")
	orderid, err := strconv.ParseInt(c.Query("order_id"), 10, 64)
	mail := c.Query("mail")

	if token != "" && orderid != 0 && err == nil && mail != "" {
		// check cli token
		_, err := dbCheckCliToken(token)

		if err != nil {
			ret404(c)
		} else {
			err := dbCheckCliSess(cli_uuid, orderid)

			if err != nil {
				ret404(c)
			} else {
				// get fact link
				link, err := dbGetOrderFact(orderid)
				if err != nil {
					ret404(c)
				} else {
					// let's send this fact
					err := eapMail.SendCliFact(link, mail)
					if err != nil {
						ret503(c)
					} else {
						c.JSON(200, "mail send")
					}
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
			ret404(c)
		} else {

			filepath, err := eapCSV.FactstoCSV(content, etabid, start, end)
			if err != nil {
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
