package main

import (
	"errors"
	"fmt"
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
				// insert serveurs token
				serverToken := uuid.New().String()
				_, err = db.Exec("INSERT INTO qr_tokens (etab_id, token, type) VALUES (?, ?, ?) ", etabId, serverToken, 1)
				if err != nil {
					log.Error("failed to insert sample product", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				} else {
					err = CreateQR(serverToken, true)
					if err != nil {
						log.Error("failed to create QRCode")
						fmt.Println(err)
					}
				}
				// insert clients token
				clientToken := uuid.New().String()
				_, err = db.Exec("INSERT INTO qr_tokens (etab_id, token, type) VALUES (?, ?, ?) ", etabId, clientToken, 0)
				if err != nil {
					log.Error("failed to insert sample product", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
					fmt.Println(err)
				} else {
					err = CreateQR(clientToken, false)
					if err != nil {
						log.Error("failed to create QRCode")
						fmt.Println(err)
					}
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

func insertNewPWD(pwdForm PWD) (err error) {
	db := dbConnect()

	ifExists := 0
	var noRow = errors.New("sql: no rows in result set")
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE security_token = ?", pwdForm.Token)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		err = errors.New("no matching row")
		log.Error("no matching row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else if err != nil {
		log.Error("get row failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		err = errors.New("find row failed")
	} else {
		// go insert new data
		_, err = db.Exec("UPDATE etabs SET hash_pwd = ?, security_token = NULL WHERE id = ? ", pwdForm.Password, ifExists)
		if err != nil {
			log.Error("get row failed", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
			err = errors.New("unable to add pwd")
		}
	}
	return err
}

func CliConnect(connForm ClientConn) (err error, token string) {
	db := dbConnect()

	ifExists := 0
	var noRow = errors.New("sql: no rows in result set")
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE mail = ? AND hash_pwd = ?", connForm.Mail, connForm.Password)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		err = errors.New("no matching row")
		log.Error("no matching row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else if err != nil {
		log.Error("get row failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		err = errors.New("find row failed")
	} else {
		// create new auth token
		token = uuid.New().String()

		// insert connect data
		_, err = db.Exec("INSERT INTO conections (etab_id, token) VALUES (?, ?)", ifExists, token)
		if err != nil {
			fmt.Println(err)
			log.Error("insert row failed", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
			err = errors.New("insert connection row failed")
		}
	}
	return err, token
}

func ResetAllConn(etabid int64) (err error) {

	db := dbConnect()

	_, err = db.Exec("DELETE FROM conections WHERE etab_id = ?", etabid)

	fmt.Println(err)
	if err != nil {
		log.Error("delete all connections failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}

func getUserId(auth string) (userid int64, err error) {

	db := dbConnect()

	err = db.Get(&userid, "SELECT etab_id FROM conections WHERE token = ?", auth)

	if err != nil {
		log.Error("auth client failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		err = errors.New("get userid failed")
	}

	return userid, err
}

func dbGetEtabs(mail string) (err error, etabs []*Etab) {

	db := dbConnect()

	etabs = []*Etab{}

	err = db.Select(&etabs, "SELECT id, name, siret, addr, cp, city, country FROM etabs WHERE mail = ?", mail)
	if err != nil {
		fmt.Println(err)
		log.Error("get etabs failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err, etabs
}

func getOwnerInfos(mail string, etabId int64) (ownerInfos Owner, err error) {
	db := dbConnect()

	err = db.Get(&ownerInfos, "SELECT owner_civility, owner_name, owner_surname, mail, name, siret, addr, cp, city, country FROM etabs WHERE id = ?", etabId)
	if err != nil {
		fmt.Println(err)
		log.Error("cannot find owner data", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}
	return ownerInfos, err
}

func AddSecuToken(etabId int64) (temptoken string, err error) {

	db := dbConnect()

	temptoken = uuid.New().String()

	_, err = db.Exec("UPDATE etabs SET security_token = ? WHERE id = ?", temptoken, etabId)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot add new secu token", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}
	return temptoken, err
}

func dbDisconnect(auth string) (err error) {

	db := dbConnect()

	_, err = db.Exec("DELETE FROM conections WHERE token = ?", auth)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot delete conenction", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}

func checkNcreateSession(authToken string) (token string, err error) {

	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT etab_id FROM qr_tokens WHERE token = ? ", authToken)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		fmt.Println(err)
		log.Error("This token doesn't exists", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))

	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot get row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {
		// create new token
		token = uuid.New().String()
		_, err = db.Exec("INSERT INTO conections (etab_id, token, is_admin) VALUES (?, ?, ?)", ifExists, token, 0)
		if err != nil {
			fmt.Println(err)
			log.Error("insert new row failed", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	}

	return token, err
}

func checkCliToken(token string) (err error, etabid int) {

	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&etabid, "SELECT etab_id FROM qr_tokens WHERE token = ? ", token)

	if etabid == 0 || (err != nil && err.Error() == noRow.Error()) {
		fmt.Println(err)
		log.Error("This token doesn't exists", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))

	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot get row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err, etabid
}

func insertCliSess(clientUuid string) (err error) {
	db := dbConnect()

	_, err = db.Exec("INSERT INTO cli_sess (cli_uuid) VALUES (?)", clientUuid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot insert row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}

func getEtabMenu(etabid int) (err error, menu Etab) {

	db := dbConnect()

	menu = Etab{}

	err = db.Get(&menu, "SELECT id, name, siret, addr, cp, city, country FROM etabs WHERE id = ?", etabid)
	if err != nil {
		fmt.Println(err)
		log.Error("get etab failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	err = db.Select(&menu.Items, "SELECT id, in_stock, name, description, price, priceHH, category FROM items WHERE etab_id = ?", etabid)
	if err != nil {
		fmt.Println(err)
		log.Error("get menu failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err, menu
}
