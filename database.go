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

func PostDBSub(subForm Subscription) (temptoken string, err error) {
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
					return temptoken, err
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

	return temptoken, err

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

func CliConnect(connForm ClientConn) (token string, err error) {
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
	return token, err
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

func dbGetEtabs(mail string) (etabs []*Etab, err error) {

	db := dbConnect()

	etabs = []*Etab{}

	err = db.Select(&etabs, "SELECT id, name, siret, addr, cp, city, country FROM etabs WHERE mail = ?", mail)
	if err != nil {
		fmt.Println(err)
		log.Error("get etabs failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return etabs, err
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

func checkCliToken(token string) (etabid int, err error) {

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

	return etabid, err
}

func checkToken(token string) (etabid int, err error) {

	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&etabid, "SELECT etab_id FROM conections WHERE token = ? ", token)

	if etabid == 0 || (err != nil && err.Error() == noRow.Error()) {
		fmt.Println(err)
		log.Error("This token doesn't exists", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))

	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot get row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return etabid, err
}

func insertCliSess(clientUuid string) (err error) {
	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT id FROM cli_sess WHERE cli_uuid = ? ", clientUuid)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {

		_, err = db.Exec("INSERT INTO cli_sess (cli_uuid) VALUES (?)", clientUuid)

		if err != nil {
			fmt.Println(err)
			log.Error("cannot insert row", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot get row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {
		// clientuuid already here, update date
		_, err = db.Exec("UPDATE cli_sess SET updated = ? WHERE cli_uuid = ?", time.Now(), clientUuid)
		if err != nil {
			fmt.Println(err)
			log.Error("cannot update row", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	}

	return err
}

func checkCliSess(cli_uuid string, orderid int64) (err error) {
	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT id FROM orders WHERE cli_uuid = ? AND id = ?", cli_uuid, orderid)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		fmt.Println(err)
		log.Error("cannot find row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))

	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot request row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}

func getEtabMenu(etabid int) (menu Etab, err error) {

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

	return menu, err
}

func dbGetPlanning(etabid int) (planning []*Planning, err error) {

	db := dbConnect()

	err = db.Select(&planning, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ? ORDER BY day asc", etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("get planning failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return planning, err
}

func dbPlaceOrder(PLOrder Order, etabid int) (orderid int64, err error) {
	db := dbConnect()

	// insert order
	insertOrder, err := db.Exec("INSERT INTO orders (cli_uuid, etab_id, totalTTC, totalHT) VALUES (?, ?, ?, ?)", PLOrder.Cli_uuid, etabid, PLOrder.TotalTTC, PLOrder.TotalHT)
	if err != nil {
		fmt.Println(err)
		log.Error("cannot insert row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {
		// get orderid
		orderid, err = insertOrder.LastInsertId()
		if err != nil {
			fmt.Println(err)
			log.Error("order not inserted", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		} else {
			// insert all items
			for _, item := range PLOrder.Order_items {
				fmt.Println(item.Item_id)
				_, err := db.Exec("INSERT INTO order_items (item_id, order_id, price, quantity) VALUES (?, ?, ?, ?)", item.Item_id, orderid, item.Price, item.Quantity)
				if err != nil {
					fmt.Println(err)
					log.Error("item not inserted", zap.String("database", viper.GetString("database.dbname")),
						zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
				}
			}
		}
	}

	return orderid, err
}

func dbUpdateOrderStatus(details OrderDetails) (err error) {
	db := dbConnect()

	// update only confirmed
	_, err = db.Exec("UPDATE orders SET confirmed = ? WHERE id = ?", details.Confirmed, details.OrderId)
	if err != nil {
		fmt.Println(err)
		log.Error("cannot update order", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}

func dbGetOrders(etabid int) (orders []*ReturnOrders, err error) {

	db := dbConnect()
	orders = []*ReturnOrders{}

	err = db.Select(&orders, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE etab_id = ?", etabid)
	if err != nil {
		fmt.Println(err)
		log.Error("get order failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	for i, _ := range orders {

		err = db.Select(&orders[i].Order_items, "SELECT order_items.quantity, order_items.price, items.category, items.name FROM order_items JOIN items ON order_items.item_id = items.id WHERE order_id = ?", orders[0].Id)
		if err != nil {
			fmt.Println(err)
			log.Error("get order items failed", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	}

	return orders, err
}

func dbGetOrder(orderid int64) (order ReturnOrders, err error) {

	db := dbConnect()
	// orders = ReturnOrders

	err = db.Get(&order, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE id = ?", orderid)
	if err != nil {
		fmt.Println(err)
		log.Error("get order failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	err = db.Select(&order.Order_items, "SELECT order_items.quantity, order_items.price, items.category, items.name FROM order_items JOIN items ON order_items.item_id = items.id WHERE order_id = ?", orderid)
	if err != nil {
		fmt.Println(err)
		log.Error("get order items failed", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return order, err
}

func getOrderFact(orderid int64) (link string, err error) {
	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&link, "SELECT fact_link FROM orders WHERE id = ?", orderid)

	if link == "" || (err != nil && err.Error() == noRow.Error()) {
		fmt.Println(err)
		log.Error("cannot find row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))

	} else if err != nil {
		fmt.Println(err)
		log.Error("cannot request row", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return link, err

}

func getFactEtab(etabid int64) (etab FactEtab, err error) {

	db := dbConnect()
	err = db.Get(&etab, "SELECT name, owner_civility, owner_name, owner_surname, mail, phone, fact_addr, fact_cp, fact_city, fact_country, offer FROM etabs WHERE etabs.id = ?", etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot get etab infos", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {
		// get offer

		err = db.Get(&etab.Etab_offer, "SELECT offers.name, offers.priceHT, offers.priceTTC FROM offers WHERE id = ?", etab.Offer)
		if err != nil {
			fmt.Println(err)
			log.Error("cannot get etab offer", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	}

	return etab, err
}

func dbGetEtabParams(etabid int64) (params EtabParams, err error) {

	db := dbConnect()

	err = db.Get(&params, "SELECT name, phone, addr, cp, city, country, IFNULL(insta, '') AS insta , IFNULL(twitter, '') AS twitter, IFNULL(facebook, '') AS facebook, licence, siret FROM etabs WHERE etabs.id = ?", etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot get etab params", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {

		err = db.Select(&params.Horaires, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ?", etabid)
		if err != nil {
			fmt.Println(err)
			log.Error("cannot get etab planning", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}
	}

	return params, err
}

func dbUpdateEtabParams(params EtabParams, etabid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE etabs SET name = ?, addr = ?, cp = ?, city = ?, country = ?, licence = ?, siret = ?, phone = ?, insta = ?, twitter = ?, facebook = ? WHERE id = ?", params.Etab_name, params.Addr, params.Cp, params.City, params.Country, params.License, params.Siret, params.Phone, params.Insta, params.Twitter, params.Facebook, etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot update etab params", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	} else {
		// update etab planning

		// first delete all rows
		_, err = db.Exec("DELETE FROM planning WHERE etab_id = ?", etabid)

		if err != nil {
			fmt.Println(err)
			log.Error("cannot delete etab planning", zap.String("database", viper.GetString("database.dbname")),
				zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
		}

		// then insert new ones
		for _, planning := range params.Horaires {
			_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end, is_active, is_HH) VALUES (?, ?, ?, ?, ?, ?)", etabid, planning.Day, planning.Start, planning.End, planning.Is_Active, planning.Is_HH)

			if err != nil {
				fmt.Println(err)
				log.Error("cannot insert planning", zap.String("database", viper.GetString("database.dbname")),
					zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
			}
		}
	}
	return err
}

func dbGetProfile(etabid int64) (profile Profile, err error) {
	db := dbConnect()

	err = db.Get(&profile, "SELECT mail, owner_civility, owner_name, owner_surname FROM etabs WHERE id = ?", etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot get profile", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return profile, err
}

func dbUpdateProfile(profile Profile, etabid int64) (err error) {

	db := dbConnect()
	_, err = db.Exec("UPDATE etabs SET owner_civility = ?, owner_name = ?, owner_surname = ?, mail = ? WHERE id = ?", profile.Civility, profile.Name, profile.Surname, profile.Mail, etabid)

	if err != nil {
		fmt.Println(err)
		log.Error("cannot update profile", zap.String("database", viper.GetString("database.dbname")),
			zap.Int("attempt", 3), zap.Duration("backoff", time.Second))
	}

	return err
}
