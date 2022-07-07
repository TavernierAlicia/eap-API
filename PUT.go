package main

import (
	"strconv"
	"bytes"
	"path/filepath"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	eapMail "github.com/TavernierAlicia/eap-MAIL"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func updateOrderStatus(c *gin.Context) {
	var details OrderDetails
	var checkDetails OrderDetails

	token := c.Request.Header.Get("Authorization")

	c.BindJSON(&details)

	if details != checkDetails && token != "" && details.OrderId != 0 {

		// check if client
		// it's a boss or server
		_, err := dbCheckToken(token)
		if err != nil {
			ret401(c)
		} else {
			err := dbUpdateOrderStatus(details)
			if err != nil {
				ret404(c)
			} else {
				c.JSON(200, gin.H{
					"message": "updated",
				})

				// send to ws
				status, _ := dbGetOrderStatus(details.OrderId) 
				postBody, _ := json.Marshal(map[string]string{
					"orderid":  strconv.Itoa(details.OrderId),
					"status": status,
				 })
				 responseBody := bytes.NewBuffer(postBody)

				_, err := http.Post("http://ws.easy-as-pie.fr/update-order", "application/json", responseBody)
				if err != nil {
   					fmt.Println(err)
				}
			}
		}

	} else {
		// send error code
		ret422(c)
	}
}

func editEtabParams(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		var params EtabParams

		c.BindJSON(&params)

		if err != nil || params.Etab_name == "" || !regSiret(params.Siret) || !regPhone(params.Phone) {
			ret422(c)
		} else {
			err = dbUpdateEtabParams(params, etabid)
			if err != nil {
				ret503(c)
			} else {
				getEtabParams(c)
			}
		}
	}
}

func editProfile(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
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

			if !ifExists {
				err = dbUpdateProfile(profile, etabid)
				if err != nil {
					ret503(c)
				} else {
					getProfile(c)
				}
			} else {
				ret404(c)
			}
		} else {
			ret422(c)
		}
	}

}

func editPaymentMethod(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		var pay Payment
		var checkPay Payment

		c.BindJSON(&pay)

		if pay != checkPay && regIban(pay.Iban) && regCP(strconv.Itoa(pay.Fact_cp)) {
			err = dbUpdatePaymentMethod(pay, etabid)

			if err != nil {
				ret503(c)
			} else {
				getPaymentMethod(c)
			}

		} else {
			ret422(c)
		}
	}
}

func editOffers(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		var randomData JSONTODATA
		c.BindJSON(&randomData)

		offerid := randomData.OfferID

		if randomData.OfferID != 0 {
			err = dbUpdateOffer(etabid, offerid)
			if err != nil {
				ret503(c)
			} else {
				getEtabOffer(c)
			}
		} else {
			ret422(c)
		}
	}
}

func putItem(c *gin.Context) {
	_, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		var item Item
		c.BindJSON(&item)

		if item.Name != "" && item.Description != "" {
			err = dbEditItem(item)

			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, "Updated")
			}
		} else {
			ret422(c)
		}
	}
}

func putCategory(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		var empty JSONTODATA
		var category JSONTODATA
		c.BindJSON(&category)

		if category.Category_name != "" && category.Category_id != 0 && category != empty {
			err = dbEditCategory(etabid, category.Category_name, category.Category_id)
			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, gin.H{
					"message": "category updated",
				})
			}
		} else {
			ret422(c)
		}
	}
}

func unsub(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {
		etab, echeance, err := dbUnsub(etabid)

		if err != nil {
			ret503(c)
		} else {
			err = eapMail.AskDeleteAccount(etab, echeance)

			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, gin.H{
					"message": "unsubscription confirmed",
				})
			}
		}
	}
}

func updatePlanning(c *gin.Context) {
	etabid, err := checkAuth(c)
	if err != nil {
		ret401(c)
	} else {
		var planning []*Planning
		c.BindJSON(&planning)

		if planning == nil {
			ret422(c)
		} else {
			err = dbEditPlanning(etabid, planning)
			if err != nil {
				ret503(c)
			} else {
				c.JSON(200, gin.H{
					"message": "new planning inserted",
				})
			}
		}
	}
}


func updateEtabPic(c *gin.Context) {
	etabid, err := checkAuth(c)

	if err != nil {
		ret401(c)
	} else {

		file, err := c.FormFile("pic")
		if err != nil {
			ret422(c)
		} else {
			newName := strconv.FormatInt(etabid, 10)+"_"+uuid.New().String()
			if len(file.Filename) > 0 {
				ext := strings.ToLower(filepath.Ext(file.Filename))



				fmt.Println(ext)
				if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
					ret422(c)
					return
				}
				pic := viper.GetString("links.cdn_pics_dest")+newName+ext
				err := c.SaveUploadedFile(file, pic)

				if err != nil {
					fmt.Println(err)
					ret503(c)
					return
				}

				
				path, err := dbUpdatePic(viper.GetString("links.cdn_pics")+newName+ext, etabid)
				if err != nil {
					fmt.Println(err)
					ret503(c)
					return
				} else {
					c.JSON(200, path)
				}
			}
		}
	}
}