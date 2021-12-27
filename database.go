package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var log *zap.Logger

func dbConnect() *sqlx.DB {
	log, _ = zap.NewProduction()
	defer log.Sync()
	//// IMPORT CONFIG ////
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Unable to load config file", zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	//// DB CONNECTION ////
	pathSQL := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", viper.GetString("database.user"), viper.GetString("database.pass"), viper.GetString("database.host"), viper.GetInt("database.port"), viper.GetString("database.dbname"))
	db, err := sqlx.Connect("mysql", pathSQL)
	if err != nil {
		log.Error("failed to connect database", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		return db

	} else {
		log.Info("Connexion etablished ", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}
	return db
}

func PostDBSub(subForm Subscription) (err error, temptoken string) {
	db := dbConnect()

	// Verify if user already exists
	ifExists := 0
	var noRow = errors.New("sql: no rows in result set")
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE siret = ?", subForm.Siret)

	if err != nil {
		if err.Error() == noRow.Error() {
			err = nil
			// if no row, user doesn't exists
			// create temptoken to init password
			temptoken = uuid.New().String()
			// insert now etab in db
			insertEtab, err := db.Exec("INSERT INTO etabs (name, owner_civility, owner_name, owner_surname, mail, phone, siret, licence, addr, cp, city, country, offer, Iban, name_Iban, fact_addr, fact_cp, fact_city, fact_country, security_token) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ", subForm.Entname, subForm.Civility, subForm.Name, subForm.Surname, subForm.Mail, subForm.Phone, subForm.Siret, subForm.Licence, subForm.Addr, subForm.Cp, subForm.City, subForm.Country, subForm.Offer, subForm.Iban, subForm.Name_iban, subForm.Fact_addr, subForm.Fact_cp, subForm.Fact_city, subForm.Fact_country, temptoken)
			if err != nil {
				log.Error("failed to insert etab", zap.String("database", viper.GetString("database.dbname")),
					zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
				fmt.Println(err)
			} else {
				etabId, err := insertEtab.LastInsertId()
				if err != nil {
					err = errors.New("something wrong happened")
					return err, temptoken
				}
				// insert other fields in database to make it work baby
				_, err = db.Exec("UPDATE etabs SET qr_code_path = ? WHERE id = ? ", viper.GetString("links.cdn_qr")+fmt.Sprintf("%v", etabId)+".png", etabId)
				if err != nil {
					log.Error("failed to insert planning samples", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				} else {
					err = CreateQR(viper.GetString("links.cdn_qr")+strconv.FormatInt(etabId, 10), etabId)
					if err != nil {
						log.Error("failed to create QRCode")
						fmt.Println(err)
					}
				}
				_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end) VALUES (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?) ", etabId, 0, 540, 800, etabId, 0, 1000, 2000, etabId, 1, 540, 800, etabId, 1, 1000, 2000, etabId, 2, 540, 800, etabId, 2, 1000, 2000, etabId, 3, 540, 800, etabId, 3, 1000, 2000, etabId, 4, 540, 800, etabId, 4, 1000, 2000, etabId, 5, 540, 800, etabId, 5, 1000, 2000)
				if err != nil {
					log.Error("failed to insert planning samples", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				}
				_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end, is_HH) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)", etabId, 5, 1000, 1300, 1, etabId, 3, 1000, 1300, 1)
				if err != nil {
					log.Error("failed to insert planning sample happy hours", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				}
				_, err = db.Exec("INSERT INTO items (etab_id, category) VALUES (?, ?) ", etabId, "Cocktails")
				if err != nil {
					log.Error("failed to insert sample product", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				}
			}
		} else {
			log.Error("failed to request etabs", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	} else {
		err = errors.New("etab already exists")
		log.Error("etab already exists", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err, temptoken

}
