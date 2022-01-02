package main

import (
	"strconv"

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

func EditEtabParams(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		var params EtabParams

		c.BindJSON(&params)

		if err != nil || params.Etab_name == "" || !regSiret(params.Siret) || !regPhone(params.Phone) {
			c.JSON(422, gin.H{
				"message": "invalid entries",
			})
		} else {
			err = dbUpdateEtabParams(params, etabid)
			if err != nil {
				c.JSON(500, gin.H{
					"message": "cannot update params",
				})
			} else {
				getEtabParams(c)
			}
		}
	}
}

func EditProfile(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		var profile Profile
		var checkProfile Profile

		c.BindJSON(&profile)

		if profile != checkProfile && regMail(profile.Mail) {
			etabs, err := dbGetEtabs(profile.Mail)
			var ifExists bool

			if err == nil {
				ifExists = true
				for _, etab := range etabs {
					if int64(etab.Id) == etabid {
						ifExists = false
						break
					}
				}
			} else {
				ifExists = false
			}

			if ifExists == false {
				err = dbUpdateProfile(profile, etabid)
				if err != nil {
					c.JSON(500, gin.H{
						"message": "cannot update profile",
					})
				} else {
					getProfile(c)
				}
			} else {
				c.JSON(401, gin.H{
					"message": "mail already taken",
				})
			}
		} else {
			c.JSON(422, gin.H{
				"message": "invalid entries",
			})
		}
	}

}

func EditPaymentMethod(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "not connected",
		})
	} else {
		var pay Payment
		var checkPay Payment

		c.BindJSON(&pay)

		if pay != checkPay && regIban(pay.Iban) && regCP(strconv.Itoa(pay.Fact_cp)) {
			err = dbUpdatePaymentMethod(pay, etabid)

			if err != nil {
				c.JSON(500, gin.H{
					"message": "cannot update payment method",
				})
			} else {
				getPaymentMethod(c)
			}

		} else {
			c.JSON(422, gin.H{
				"message": "invalid entries",
			})
		}
	}
}
