package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Subscribe(c *gin.Context) {

	// recept data
	var subForm Subscription
	var checkForm Subscription
	c.BindJSON(&subForm)

	// check data if not nil
	var match bool
	ok := true
	if subForm != checkForm {
		// reg iban
		match, _ = regexp.MatchString("^[a-zA-Z]{2}[0-9]{2}\\s?[a-zA-Z0-9]{4}\\s?[0-9]{4}\\s?[0-9]{3}([a-zA-Z0-9]\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,3})?$", subForm.Iban)
		if !match {
			ok = false
		}

		// reg siret
		match, _ = regexp.MatchString("^[0-9]{14}$", subForm.Siret)
		if !match {
			ok = false
		}

		// reg mail
		match, _ = regexp.MatchString("^((\\w[^\\W]+)[\\.\\-]?){1,}\\@(([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})|(([a-zA-Z\\-0-9]+\\.)+[a-zA-Z]{2,}))$", subForm.Mail)
		if !match {
			ok = false
		}

		// reg phone
		match, _ = regexp.MatchString("^(0[1-9]{1}[0-9]{8}|\\+?33[1-9][0-9]{8})$", subForm.Phone)
		if !match {
			ok = false
		}

		// reg cp
		match, _ = regexp.MatchString("^[0-9]{5}$", strconv.Itoa(subForm.Cp))

		if !match {
			ok = false
		}

		// reg fact_cp
		match, _ = regexp.MatchString("^[0-9]{5}$", strconv.Itoa(subForm.Fact_cp))

		if !match {
			ok = false
		}
	} else {
		ok = false
	}

	if ok {
		temptoken, err := PostDBSub(subForm)
		if err != nil {
			// send error code
			c.JSON(503, gin.H{
				"message": "subscribtion failed",
			})
		} else {
			err = AddPWD(subForm, temptoken)
			// send error code
			if err != nil {
				c.JSON(503, gin.H{
					"message": "subscribtion failed",
				})
			} else {
				// send ok code
				c.JSON(201, gin.H{
					"message": "account created",
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

func createPWD(c *gin.Context) {
	// recept data
	var pwdForm PWD
	var checkForm PWD
	c.BindJSON(&pwdForm)

	if pwdForm != checkForm && pwdForm.Token != "" && pwdForm.Password != "" && pwdForm.Confirm_password != "" {
		if pwdForm.Password == pwdForm.Confirm_password {
			// check security token and insert new PWD
			err := insertNewPWD(pwdForm)
			if err != nil {
				c.JSON(503, gin.H{
					"message": "add pwd failed",
				})
			} else {
				// send ok code
				c.JSON(201, gin.H{
					"message": "password created",
				})
			}
		} else {
			// send error code
			c.JSON(422, gin.H{
				"message": "passwords mismatch",
			})
		}
	} else {
		// send error code
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func Connect(c *gin.Context) {
	// recept data
	var connForm ClientConn
	var checkForm ClientConn
	c.BindJSON(&connForm)

	if checkForm != connForm && connForm.Mail != "" && connForm.Password != "" {
		// check password
		token, err := CliConnect(connForm)
		if err != nil {
			c.JSON(422, gin.H{
				"message": "password mail mismatch",
			})
		} else {
			// send ok code
			c.JSON(200, gin.H{
				"message": "connected",
				"token":   token,
			})
		}
	} else {
		// send error code
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}

func SM4resetPWD(c *gin.Context) {
	mail := c.Request.Header.Get("mail")
	etabId, err := strconv.ParseInt(c.Request.Header.Get("etabid"), 10, 64)

	if mail != "" && etabId != 0 && err == nil {
		// get owner infos for the mail
		ownerInfos, err := getOwnerInfos(mail, etabId)
		if err != nil {
			// send error code
			c.JSON(404, gin.H{
				"message": "owner infos not found",
			})
		} else {
			// Add security token
			temptoken, err := AddSecuToken(etabId)
			if err != nil {
				c.JSON(503, gin.H{
					"message": "add temptoken failed",
				})
			} else {
				// disconnect everyone
				err = ResetAllConn(etabId)

				if err != nil {
					c.JSON(503, gin.H{
						"message": "reset connections failed",
					})
				} else {
					err = NewPWD(ownerInfos, temptoken)
					if err != nil {
						c.JSON(503, gin.H{
							"message": "send mail failed",
						})
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
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}

}

func QRConnect(c *gin.Context) {
	var authToken ServQRToken
	var checkForm ServQRToken

	c.BindJSON(&authToken)

	if authToken != checkForm {
		token, err := checkNcreateSession(authToken.Token)
		fmt.Println(token)

		if err != nil {
			c.JSON(503, gin.H{
				"message": "create connection failed",
			})
		} else {
			c.JSON(200, gin.H{
				"message": "connected",
				"token":   token,
			})
		}
	} else {
		c.JSON(422, gin.H{
			"message": "invalid entries",
		})
	}
}
