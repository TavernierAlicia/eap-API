package main

import (
	"errors"

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
