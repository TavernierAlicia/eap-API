package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger *zap.Logger

// auth
func checkAuth(c *gin.Context) (userid int64, err error) {
	auth := c.Request.Header.Get("Authorization")
	err = nil

	if auth != "" {
		userid, err = dbGetUserId(auth)
		if userid == 0 {
			err = errors.New("no user detected")
		}
	} else {
		err = errors.New("empty token")
	}
	return userid, err
}

// regex
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

// codes http return

func ret404(c *gin.Context) {
	c.JSON(404, gin.H{
		"message": "something went wrong",
		"error":   "Not found",
	})
}

func ret401(c *gin.Context) {
	c.JSON(401, gin.H{
		"message": "you must be connected to reach this page",
		"error":   "Unauthorized",
	})
}

func ret422(c *gin.Context) {
	c.JSON(422, gin.H{
		"message": "cannot use this data",
		"error":   "invalid entries",
	})
}

func ret503(c *gin.Context) {
	c.JSON(503, gin.H{
		"message": "this service encounters a problem, please retry",
		"error":   "Unavaillable",
	})
}

// print errors
func printErr(desc string, nomFunc string, err error) {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	if err != nil {
		logger.Error("Cannot "+desc, zap.String("Func", nomFunc), zap.Error(err))
	}
}

func toCSV(content []*RenderCSV, etabid int64, start string, end string) (filepath string, err error) {

	var rows [][]string

	filepath = "media/csvs/" + strconv.FormatInt(etabid, 10) + "_" + strings.ReplaceAll(start, " ", "-") + "_to_" + strings.ReplaceAll(end, " ", "-") + "-export.csv"

	file, err := os.Create(filepath)

	if err != nil {
		printErr("create csv file", "toCSV", err)
	}

	writer := csv.NewWriter(file)

	for _, row := range content {

		fmt.Println(row.Id, row.Name, row.Quantity, row.Price, row.Order_id, row.Order_date)
		rows = append(rows, []string{strconv.Itoa(row.Id), row.Name, strconv.Itoa(row.Quantity), fmt.Sprintf("%.2f", row.Price), strconv.Itoa(row.Order_id), row.Order_date})

	}

	err = writer.WriteAll(rows)
	if err != nil {
		printErr("write csv rows", "toCSV", err)
	}

	return filepath, err

}
