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

	fmt.Println(subForm)

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
		fmt.Println("ok on emballe")
		err, temptoken := PostDBSub(subForm)
		if err != nil {
			fmt.Println("et c'est la merde")
			// send error code
		} else {
			err = AddPWD(subForm, temptoken)
		}
	} else {
		fmt.Println("renvoie une erreur")
		// send reset password mail and 200

	}
}

// func Connect(c *gin.Context) {

// }
