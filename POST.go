package main

import (
	"strconv"
	"fmt"
	eapMail "github.com/TavernierAlicia/eap-MAIL"
	eapFact "github.com/TavernierAlicia/eap-FACT"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"net/http"


)

func Subscribe(c *gin.Context) {

	// recept data
	var subForm eapMail.Subscription
	var checkForm eapMail.Subscription
	c.BindJSON(&subForm)

	// check data if not nil

	if subForm != checkForm {

		if regIban(subForm.Iban) && regSiret(subForm.Siret) && regMail(subForm.Mail) && regPhone(subForm.Phone) && regCP(strconv.Itoa(subForm.Cp)) && regCP(strconv.Itoa(subForm.Fact_cp)) {
			temptoken, err := dbPostSub(subForm)
			if err != nil {
				// send error code
				ret503(c)
			} else {
				err = eapMail.AddPWD(subForm, temptoken)
				// send error code
				if err != nil {
					ret503(c)
				} else {
					// send ok code
					c.JSON(201, gin.H{
						"message": "account created",
					})
				}
			}
		} else {
			ret422(c)
		}

	} else {
		// send error code
		ret422(c)
	}

}

func createPWD(c *gin.Context) {
	// recept data
	var pwdForm PWD
	var checkForm PWD
	c.BindJSON(&pwdForm)

	if pwdForm != checkForm && pwdForm.Token != "" && pwdForm.Password != "" && pwdForm.Confirm_password != "" {
		if pwdForm.Password == pwdForm.Confirm_password {
			// check security token and insert new PWD
			err := dbInsertNewPWD(pwdForm)
			if err != nil {
				ret503(c)
			} else {
				// send ok code
				c.JSON(201, gin.H{
					"message": "password created",
				})
			}
		} else {
			// send error code
			ret422(c)
		}
	} else {
		// send error code
		ret422(c)
	}
}

func Connect(c *gin.Context) {
	// recept data
	var connForm ClientConn
	var checkForm ClientConn
	c.BindJSON(&connForm)

	if checkForm != connForm && connForm.Mail != "" && connForm.Password != "" {
		// check password
		token, err := dbCliConnect(connForm)
		if err != nil {
			if err.Error() == "suspended account" {
				ret403(c)
			} else if err != nil {
				ret401(c)
			}
		} else {
			c.JSON(200, gin.H{
				"message": "connected",
				"token":   token,
			})
		}
	} else {
		ret422(c)
	}
}

func SM4resetPWD(c *gin.Context) {
	var randomData JSONTODATA
	c.BindJSON(&randomData)

	mail := randomData.Mail
	etabId := randomData.EtabID

	if mail != "" && etabId != 0 {
		// get owner infos for the mail
		ownerInfos, err := dbGetOwnerInfos(mail, etabId)
		if err != nil {
			// send error code
			ret404(c)
		} else {
			// Add security token
			temptoken, err := dbAddSecuToken(etabId)
			if err != nil {
				ret503(c)
			} else {
				// disconnect everyone
				err = dbResetAllConn(etabId)

				if err != nil {
					ret503(c)
				} else {
					err = eapMail.NewPWD(ownerInfos, temptoken)
					if err != nil {
						ret503(c)
					} else {
						c.JSON(200, gin.H{
							"message": "ready for password reset",
						})
					}
				}
			}
		}
	} else {
		// send error code
		ret422(c)
	}

}

func QRConnect(c *gin.Context) {
	var authToken ServQRToken
	var checkForm ServQRToken

	c.BindJSON(&authToken)

	if authToken != checkForm {
		token, err := dbCheckNcreateSession(authToken.Token)
		if err != nil {
			if err.Error() == "suspended account" {
				ret403(c)
			} else if err != nil {
				ret401(c)
			}
		} else {
			c.JSON(200, gin.H{
				"message": "connected",
				"token":   token,
			})
		}

	} else {
		ret422(c)
	}
}

func placeOrder(c *gin.Context) {
	var PLOrder eapFact.Order

	c.BindJSON(&PLOrder)

	if PLOrder.Cli_uuid != "" && PLOrder.Token != "" && len(PLOrder.Order_items) != 0 {
		// check token && get etabid
		etabid, err := dbCheckCliToken(PLOrder.Token)

		if err != nil {
			ret404(c)
		} else {
			// check client_uuid
			err := dbCheckCliSess(PLOrder.Cli_uuid)

			if err != nil {
				ret401(c)
			} else {
				// Now insert order
				
				link := viper.GetString("links.cdn_tickets")
				orderid, uuidticket, err := dbPlaceOrder(PLOrder, etabid, link)
				dest := viper.GetString("links.cdn_tickets_dest")+uuidticket+".pdf"


				etab, err := dbGetEtabInfos(etabid)
				err = eapFact.CreateTicket(orderid, dest, PLOrder, etab)

				if err != nil {
					if err == invalidData {
						ret409(c)
					} else {
						fmt.Println(err)
						ret503(c)
					}

				} else {
					c.JSON(200, orderid)
					_, err := http.Post("http://ws.easy-as-pie.fr/new-order/"+strconv.Itoa(etabid), "", nil)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

	} else {
		ret422(c)
	}
}

func postItem(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		var item Item
		c.BindJSON(&item)

		if item.Name != "" && item.Price > 0 && item.Category != "" {
			err = dbInsertItem(item, etabid)

			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, "Inserted")
			}

		} else {
			ret422(c)
		}
	}
}

func postCategory(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		var category JSONTODATA
		c.BindJSON(&category)

		if category.Category_name != "" {
			err = dbInsertCategory(etabid, category.Category_name)
			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, gin.H{
					"message": "category inserted",
				})
			}
		} else {
			ret422(c)
		}
	}
}


func Cli(c *gin.Context) {

	token := c.Param("token")

	if token == "" {
		ret422(c)
	} else {
		_, err := dbCheckCliToken(token)

		if err != nil {
				ret401(c)
		} else {

			clientUuid := uuid.New().String()
			err := dbInsertCliSess(clientUuid)

			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, clientUuid)
			}
		}
	}
}


func Send(c *gin.Context) {

	var msg eapMail.Message
	var check eapMail.Message
	c.BindJSON(&msg)

	if msg != check && msg.Name != "" && regMail(msg.Mail) == true && msg.Msg != "" {
		err := eapMail.SendContact(msg)

		if err != nil {
			ret503(c)
		} else {
			c.JSON(200, "send")
		}
	} else {
		ret422(c)
	}
}