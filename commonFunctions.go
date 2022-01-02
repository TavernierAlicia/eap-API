package main

import (
	"errors"
	"regexp"

	"github.com/gin-gonic/gin"
)

func checkAuth(c *gin.Context) (userid int64, err error) {
	auth := c.Request.Header.Get("Authorization")
	err = nil

	if auth != "" {
		userid, err = getUserId(auth)
		if userid == 0 {
			err = errors.New("no user detected")
		}
	} else {
		err = errors.New("empty token")
	}
	return userid, err
}

func regIban(iban string) (match bool) {
	match, _ = regexp.MatchString("^[a-zA-Z]{2}[0-9]{2}\\s?[a-zA-Z0-9]{4}\\s?[0-9]{4}\\s?[0-9]{3}([a-zA-Z0-9]\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,4}\\s?[a-zA-Z0-9]{0,3})?$", iban)
	return match
}

func regSiret(siret string) (match bool) {
	match, _ = regexp.MatchString("^[0-9]{14}$", siret)
	return match
}

func regMail(mail string) (match bool) {
	match, _ = regexp.MatchString("^((\\w[^\\W]+)[\\.\\-]?){1,}\\@(([0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3})|(([a-zA-Z\\-0-9]+\\.)+[a-zA-Z]{2,}))$", mail)
	return match
}

func regPhone(phone string) (match bool) {
	match, _ = regexp.MatchString("^(0[1-9]{1}[0-9]{8}|\\+?33[1-9][0-9]{8})$", phone)
	return match
}

func regCP(cp string) (match bool) {
	match, _ = regexp.MatchString("^[0-9]{5}$", cp)
	return match
}
