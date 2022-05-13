package main

import (
	"errors"
	"fmt"
	"time"

	eapFact "github.com/TavernierAlicia/eap-FACT"
	eapMail "github.com/TavernierAlicia/eap-MAIL"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func dbConnect() *sqlx.DB {

	//// IMPORT CONFIG ////
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	printErr("reading config file", "dbConnect", err)

	//// DB CONNECTION ////
	pathSQL := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", viper.GetString("database.user"), viper.GetString("database.pass"), viper.GetString("database.host"), viper.GetInt("database.port"), viper.GetString("database.dbname"))
	db, err := sqlx.Connect("mysql", pathSQL)

	printErr("connect to database", "dbConnect", err)
	return db
}

func dbPostSub(subForm eapMail.Subscription) (temptoken string, err error) {
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
				printErr("insert etab", "dbPostSub", err)
			} else {
				etabId, err := insertEtab.LastInsertId()
				if err != nil {
					printErr("get lastrowid", "PostSubForm", err)
				}

				_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end) VALUES (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?) ", etabId, 0, 540, 800, etabId, 0, 1000, 2000, etabId, 1, 540, 800, etabId, 1, 1000, 2000, etabId, 2, 540, 800, etabId, 2, 1000, 2000, etabId, 3, 540, 800, etabId, 3, 1000, 2000, etabId, 4, 540, 800, etabId, 4, 1000, 2000, etabId, 5, 540, 800, etabId, 5, 1000, 2000)
				printErr("insert planning", "dbPostSub", err)

				_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end, is_HH) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)", etabId, 5, 1000, 1300, 1, etabId, 3, 1000, 1300, 1)
				printErr("insert HH planning", "dbPostSub", err)

				// insert serveurs token
				serverToken := uuid.New().String()
				_, err = db.Exec("INSERT INTO qr_tokens (etab_id, token, type) VALUES (?, ?, ?) ", etabId, serverToken, 1)
				if err != nil {
					printErr("insert serv qr token", "dbPostSub", err)
				} else {
					err = createQR(serverToken, true)
					printErr("create qr", "dbPostSub", err)
				}
				// insert clients token
				clientToken := uuid.New().String()
				_, err = db.Exec("INSERT INTO qr_tokens (etab_id, token, type) VALUES (?, ?, ?) ", etabId, clientToken, 0)
				if err != nil {
					printErr("insert cli qr token", "dbPostSub", err)
				} else {
					err = createQR(clientToken, false)
					printErr("create qr cli", "dbPostSub", err)
				}
				_, err = db.Exec("INSERT INTO items (etab_id, category) VALUES (?, ?) ", etabId, "Cocktails")
				printErr("insert item", "dbPostSub", err)
			}
		} else {
			printErr("insert data, id already exists", "dbPostSub", err)
		}
	} else {
		printErr("insert etab, already existing", "dbPostSub", err)
	}

	return temptoken, err

}

func dbInsertNewPWD(pwdForm PWD) (err error) {
	db := dbConnect()

	ifExists := 0
	var noRow = errors.New("sql: no rows in result set")
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE security_token = ?", pwdForm.Token)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "InsertNewPWD", err)
	} else if err != nil {
		printErr("request", "InsertNewPWD", err)
	} else {
		// go insert new data
		_, err = db.Exec("UPDATE etabs SET hash_pwd = ?, security_token = NULL WHERE id = ? ", pwdForm.Password, ifExists)
		printErr("update password", "InsertNewPWD", err)
	}
	return err
}

func dbCliConnect(connForm ClientConn) (token string, err error) {
	db := dbConnect()

	ifExists := 0
	var noRow = errors.New("sql: no rows in result set")
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE mail = ? AND hash_pwd = ?", connForm.Mail, connForm.Password)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCliConnect", err)
	} else if err != nil {
		printErr("request", "dbCliConnect", err)
	} else {
		// create new auth token
		token = uuid.New().String()

		// insert connect data
		_, err = db.Exec("INSERT INTO conections (etab_id, token) VALUES (?, ?)", ifExists, token)
		printErr("insert connection", "dbCliConnect", err)
	}
	return token, err
}

func dbResetAllConn(etabid int64) (err error) {

	db := dbConnect()

	_, err = db.Exec("DELETE FROM conections WHERE etab_id = ?", etabid)

	printErr("delete connection", "dbResetAllConn", err)

	return err
}

func dbGetUserId(auth string) (userid int64, err error) {

	db := dbConnect()

	err = db.Get(&userid, "SELECT etab_id FROM conections WHERE token = ?", auth)

	printErr("get row", "dbGetUserId", err)

	return userid, err
}

func dbGetEtabs(mail string) (etabs []*Etab, err error) {

	db := dbConnect()

	etabs = []*Etab{}

	err = db.Select(&etabs, "SELECT id, name, siret, addr, cp, city, country FROM etabs WHERE mail = ?", mail)
	printErr("get etabs", "dbGetEtabs", err)

	return etabs, err
}

func dbGetOwnerInfos(mail string, etabId int64) (ownerInfos eapMail.Owner, err error) {
	db := dbConnect()

	err = db.Get(&ownerInfos, "SELECT owner_civility, owner_name, owner_surname, mail, name, siret, addr, cp, city, country FROM etabs WHERE id = ?", etabId)
	printErr("get owner infos", "dbGetOwnerInfos", err)
	return ownerInfos, err
}

func dbAddSecuToken(etabId int64) (temptoken string, err error) {

	db := dbConnect()

	temptoken = uuid.New().String()

	_, err = db.Exec("UPDATE etabs SET security_token = ? WHERE id = ?", temptoken, etabId)

	printErr("update etabs", "dbAddSecuToken", err)
	return temptoken, err
}

func dbDisconnect(auth string) (err error) {

	db := dbConnect()

	_, err = db.Exec("DELETE FROM conections WHERE token = ?", auth)

	printErr("delete row", "dbDisconnect", err)

	return err
}

func dbCheckNcreateSession(authToken string) (token string, err error) {

	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT etab_id FROM qr_tokens WHERE token = ? ", authToken)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckNcreateSession", err)

	} else if err != nil {
		printErr("request", "dbCheckNcreateSession", err)
	} else {
		// create new token
		token = uuid.New().String()
		_, err = db.Exec("INSERT INTO conections (etab_id, token, is_admin) VALUES (?, ?, ?)", ifExists, token, 0)
		printErr("insert row", "dbCheckNcreateSession", err)
	}

	return token, err
}

func dbCheckCliToken(token string) (etabid int, err error) {

	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&etabid, "SELECT etab_id FROM qr_tokens WHERE token = ? ", token)

	if etabid == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckCliToken", err)

	} else if err != nil {
		printErr("request", "dbCheckCliToken", err)
	}

	return etabid, err
}

func dbCheckToken(token string) (etabid int, err error) {

	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&etabid, "SELECT etab_id FROM conections WHERE token = ? ", token)

	if etabid == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckToken", err)

	} else if err != nil {
		printErr("request", "dbCheckToken", err)
	}

	return etabid, err
}

func dbInsertCliSess(clientUuid string) (err error) {
	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT id FROM cli_sess WHERE cli_uuid = ? ", clientUuid)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {

		_, err = db.Exec("INSERT INTO cli_sess (cli_uuid) VALUES (?)", clientUuid)

		printErr("insert row", "dbInsertCliSess", err)

	} else if err != nil {
		printErr("request", "dbInsertCliSess", err)
	} else {
		// clientuuid already here, update date
		_, err = db.Exec("UPDATE cli_sess SET updated = ? WHERE cli_uuid = ?", time.Now(), clientUuid)
		printErr("update row", "dbInsertCliSess", err)
	}

	return err
}

func dbCheckCliSess(cli_uuid string, orderid int64) (err error) {
	db := dbConnect()

	var ifExists int
	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&ifExists, "SELECT id FROM orders WHERE cli_uuid = ? AND id = ?", cli_uuid, orderid)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckCliSess", err)

	} else if err != nil {
		printErr("request", "dbCheckCliSess", err)
	}

	return err
}

func dbGetEtabMenu(etabid int) (menu Etab, err error) {

	db := dbConnect()

	menu = Etab{}

	err = db.Get(&menu, "SELECT id, name, siret, addr, cp, city, country FROM etabs WHERE id = ?", etabid)
	printErr("get row", "dbGetEtabMenu", err)

	err = db.Select(&menu.Items, "SELECT id, in_stock, name, description, price, priceHH, category FROM items WHERE etab_id = ?", etabid)
	printErr("get row", "dbGetEtabMenu", err)

	return menu, err
}

func dbGetPlanning(etabid int) (planning []*Planning, err error) {

	db := dbConnect()

	err = db.Select(&planning, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ? ORDER BY day asc", etabid)

	printErr("get row", "dbGetPlanning", err)

	return planning, err
}

func dbPlaceOrder(PLOrder Order, etabid int) (orderid int64, err error) {
	db := dbConnect()

	// insert order
	insertOrder, err := db.Exec("INSERT INTO orders (cli_uuid, etab_id, totalTTC, totalHT) VALUES (?, ?, ?, ?)", PLOrder.Cli_uuid, etabid, PLOrder.TotalTTC, PLOrder.TotalHT)
	if err != nil {
		printErr("insert row", "dbPlaceOrder", err)
	} else {
		// get orderid
		orderid, err = insertOrder.LastInsertId()
		if err != nil {
			printErr("get lastrowid", "dbPlaceOrder", err)
		} else {
			// insert all items
			for _, item := range PLOrder.Order_items {
				fmt.Println(item.Item_id)
				_, err := db.Exec("INSERT INTO order_items (item_id, order_id, price, quantity) VALUES (?, ?, ?, ?)", item.Item_id, orderid, item.Price, item.Quantity)
				printErr("insert row", "dbPlaceOrder", err)
			}
		}
	}

	return orderid, err
}

func dbUpdateOrderStatus(details OrderDetails) (err error) {
	db := dbConnect()

	// update only confirmed
	_, err = db.Exec("UPDATE orders SET confirmed = ? WHERE id = ?", details.Confirmed, details.OrderId)
	printErr("update row", "dbUpdateOrderStatus", err)

	return err
}

func dbGetOrders(etabid int) (orders []*ReturnOrders, err error) {

	db := dbConnect()
	orders = []*ReturnOrders{}

	err = db.Select(&orders, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE etab_id = ?", etabid)
	printErr("get rows", "dbGetOrders", err)

	for i := range orders {

		err = db.Select(&orders[i].Order_items, "SELECT order_items.quantity, order_items.price, items.category, items.name FROM order_items JOIN items ON order_items.item_id = items.id WHERE order_id = ?", orders[0].Id)
		printErr("get rows", "dbGetOrders", err)
	}

	return orders, err
}

func dbGetOrder(orderid int64) (order ReturnOrders, err error) {

	db := dbConnect()
	// orders = ReturnOrders

	err = db.Get(&order, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE id = ?", orderid)
	printErr("get row", "dbGetOrder", err)

	err = db.Select(&order.Order_items, "SELECT order_items.quantity, order_items.price, items.category, items.name FROM order_items JOIN items ON order_items.item_id = items.id WHERE order_id = ?", orderid)
	printErr("get rows", "dbGetOrder", err)

	return order, err
}

func dbGetOrderFact(orderid int64) (link string, err error) {
	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&link, "SELECT fact_link FROM orders WHERE id = ?", orderid)

	if link == "" || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbGetOrderFact", err)

	} else if err != nil {
		printErr("request", "dbGetOrderFact", err)
	}

	return link, err

}

func dbGetFactEtab(etabid int64) (etab eapFact.FactEtab, err error) {

	db := dbConnect()
	err = db.Get(&etab, "SELECT name, owner_civility, owner_name, owner_surname, mail, phone, fact_addr, fact_cp, fact_city, fact_country, offer FROM etabs WHERE etabs.id = ?", etabid)

	if err != nil {
		printErr("get row", "dbGetFactEtab", err)
	} else {
		// get offer
		err = db.Get(&etab.Etab_offer, "SELECT id, name, priceHT, priceTTC FROM offers WHERE id = ?", etab.Offer)
		printErr("get row", "dbGetFactEtab", err)
	}

	return etab, err
}

func dbCreateBossFirstFact(etabid int64, uuid string, link string) (err error, factId int64) {
	db := dbConnect()
	insertFact, err := db.Exec("INSERT INTO factures (uuid, etab_id, link, created, payed) VALUES (?, ?, ?, NOW(), NOW())", uuid, etabid, link)
	factId, err = insertFact.LastInsertId()
	if err != nil {
		printErr("insert first fact", "dbCreateBossFirstFact", err)
	}

	return err, factId
}

func dbGetEtabParams(etabid int64) (params EtabParams, err error) {

	db := dbConnect()

	err = db.Get(&params, "SELECT name, phone, addr, cp, city, country, IFNULL(insta, '') AS insta , IFNULL(twitter, '') AS twitter, IFNULL(facebook, '') AS facebook, licence, siret FROM etabs WHERE etabs.id = ?", etabid)

	if err != nil {
		printErr("get row", "dbGetEtabParams", err)
	} else {

		err = db.Select(&params.Horaires, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ?", etabid)
		printErr("get rows", "dbGetEtabParams", err)
	}

	return params, err
}

func dbUpdateEtabParams(params EtabParams, etabid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE etabs SET name = ?, addr = ?, cp = ?, city = ?, country = ?, licence = ?, siret = ?, phone = ?, insta = ?, twitter = ?, facebook = ? WHERE id = ?", params.Etab_name, params.Addr, params.Cp, params.City, params.Country, params.License, params.Siret, params.Phone, params.Insta, params.Twitter, params.Facebook, etabid)

	if err != nil {
		printErr("update row", "dbUpdateEtabParams", err)
	} else {
		// update etab planning

		// first delete all rows
		_, err = db.Exec("DELETE FROM planning WHERE etab_id = ?", etabid)
		printErr("delete rows", "dbUpdateEtabParams", err)

		// then insert new ones
		for _, planning := range params.Horaires {
			_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end, is_active, is_HH) VALUES (?, ?, ?, ?, ?, ?)", etabid, planning.Day, planning.Start, planning.End, planning.Is_Active, planning.Is_HH)
			printErr("insert row", "dbUpdateEtabParams", err)
		}
	}
	return err
}

func dbGetProfile(etabid int64) (profile Profile, err error) {
	db := dbConnect()

	err = db.Get(&profile, "SELECT mail, owner_civility, owner_name, owner_surname FROM etabs WHERE id = ?", etabid)

	printErr("get row", "dbGetProfile", err)

	return profile, err
}

func dbUpdateProfile(profile Profile, etabid int64) (err error) {

	db := dbConnect()
	_, err = db.Exec("UPDATE etabs SET owner_civility = ?, owner_name = ?, owner_surname = ?, mail = ? WHERE id = ?", profile.Civility, profile.Name, profile.Surname, profile.Mail, etabid)

	printErr("update row", "dbUpdateProfile", err)

	return err
}

func dbGetPaymentMethods(etabid int64) (pay Payment, err error) {

	db := dbConnect()
	err = db.Get(&pay, "SELECT iban, name_iban, fact_addr, fact_cp, fact_city, fact_country FROM etabs WHERE id = ?", etabid)

	printErr("get row", "dbGetPaymentMethods", err)

	return pay, err
}

func dbUpdatePaymentMethod(pay Payment, etabid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE etabs SET iban = ?, name_iban = ?, fact_addr = ?, fact_cp = ?, fact_city = ?, fact_country = ? WHERE id = ?", pay.Iban, pay.Name_iban, pay.Fact_addr, pay.Fact_cp, pay.Fact_city, pay.Fact_country, etabid)
	printErr("update row", "dbUpdatePaymentMethods", err)

	return err
}

func dbGetOffer(etabid int64) (offer eapFact.Offer, err error) {
	db := dbConnect()
	err = db.Get(&offer, "SELECT offers.id, offers.name, offers.priceHT, offers.priceTTC FROM etabs JOIN offers ON etabs.offer = offers.id WHERE etabs.id = ?", etabid)

	printErr("get row", "dbGetOffer", err)

	return offer, err
}

func dbUpdateOffer(etabid int64, offerid int64) (err error) {
	db := dbConnect()

	var ifExists int
	// check if offer exists before
	err = db.Get(&ifExists, "SELECT id FROM offers WHERE id = ?", offerid)

	if err == nil && ifExists != 0 {
		_, err = db.Exec("UPDATE etabs SET offer = ? WHERE id = ?", offerid, etabid)
		printErr("update row", "dbUpdateOffer", err)
	} else {
		printErr("get row", "dbUpdateOffer", err)
	}

	return err
}

func dbInsertItem(item Item, etabid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("INSERT INTO items (etab_id, in_stock, name, description, price, priceHH, category) VALUES (?, ?, ?, ?, ?, ?, ?)", etabid, item.Stock, item.Name, item.Description, item.Price, item.PriceHH, item.Category)
	printErr("insert row", "dbInsertItem", err)
	return err
}

func dbEditItem(item Item) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE items SET in_stock = ?, name = ?, description = ?, price = ?, priceHH = ?, category = ?, modified = ? WHERE id = ?", item.Stock, item.Name, item.Description, item.Price, item.PriceHH, item.Category, time.Now(), item.Id)
	printErr("update row", "dbEditItem", err)

	return err
}

func dbDeleteItem(itemid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("DELETE FROM items WHERE id = ?", itemid)
	printErr("delete row", "dbDeleteItem", err)

	return err
}

func dbGetCSV(start string, end string, etabid int64) (result []*RenderCSV, err error) {
	db := dbConnect()

	err = db.Select(&result, "SELECT order_items.id, order_items.order_id, order_items.quantity, order_items.created, order_items.price, items.name FROM `order_items` JOIN items ON items.id = order_items.item_id WHERE items.etab_id = ? and order_items.created BETWEEN ? and ?", etabid, start, end)
	printErr("getting csv content", "dbGetCSV", err)

	return result, err

}
