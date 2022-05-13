package main

import (
	"fmt"
	"strconv"

	eapMail "github.com/TavernierAlicia/eap-MAIL"
	"github.com/gin-gonic/gin"
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
			ret422(c)
		} else {
			// send ok code
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
	mail := c.Request.Header.Get("mail")
	etabId, err := strconv.ParseInt(c.Request.Header.Get("etabid"), 10, 64)

	if mail != "" && etabId != 0 && err == nil {
		// get owner infos for the mail
		ownerInfos, err := dbGetOwnerInfos(mail, etabId)
		if err != nil {
			// send error code
			ret401(c)
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
		fmt.Println(token)

		if err != nil {
			ret503(c)
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
	var PLOrder Order

	c.BindJSON(&PLOrder)

	if PLOrder.Cli_uuid != "" && PLOrder.Token != "" && len(PLOrder.Order_items) != 0 {
		// check token && get etabid
		etabid, err := dbCheckCliToken(PLOrder.Token)

		if err != nil {
			ret401(c)
		} else {
			// check client_uuid
			err := dbInsertCliSess(PLOrder.Cli_uuid)

			if err != nil {
				ret404(c)
			} else {
				// Now insert order
				orderid, err := dbPlaceOrder(PLOrder, etabid)

				if err != nil {
					ret404(c)
				} else {
					c.JSON(200, orderid)
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

		if item.Name != "" && item.Description != "" {
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
