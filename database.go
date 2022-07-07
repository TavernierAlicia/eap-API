package main

import (
	"errors"
	"time"
	"fmt"

	eapFact "github.com/TavernierAlicia/eap-FACT"
	eapMail "github.com/TavernierAlicia/eap-MAIL"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var noRow = errors.New("sql: no rows in result set")
var invalidData = errors.New("incorrect data")


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

				_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end) VALUES (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ? , ?), (?, ?, ?, ?) ", etabId, 0, 540, 800, etabId, 0, 1000, 2000, etabId, 1, 540, 800, etabId, 1, 1000, 2000, etabId, 2, 540, 800, etabId, 2, 1000, 2000, etabId, 3, 540, 800, etabId, 3, 1000, 2000, etabId, 4, 540, 800, etabId, 4, 1000, 2000, etabId, 5, 540, 800, etabId, 5, 1000, 2000, etabId, 6, 1000, 2000)
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
				insertCat, _ := db.Exec("INSERT INTO categories (name, etab_id) VALUES (?, ?)", "Cocktails", etabId)
				catId, err := insertCat.LastInsertId()
				if err != nil {
					printErr("insert new category", "dbPostSub", err)
				} else {
					_, err = db.Exec("INSERT INTO items (etab_id, category_id, description, created) VALUES (?, ?, ?, NOW()) ", etabId, catId, "Un super cockail de bienvenue, car il y a une première fois à tout!")
					if err != nil {
						printErr("insert item", "dbPostSub", err)
					}
				}
			}
		} else {
			err = errors.New("insert etab already existing")
			printErr("insert data, id already exists", "dbPostSub", err)
		}
	} else {
		err = errors.New("insert etab already existing")
		printErr("insert etab, already existing", "dbPostSub", err)
	}

	return temptoken, err

}

func dbInsertNewPWD(pwdForm PWD) (err error) {
	db := dbConnect()

	ifExists := 0
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
	err = db.Get(&ifExists, "SELECT id FROM etabs WHERE mail = ? AND hash_pwd = ?", connForm.Mail, connForm.Password)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCliConnect", err)
	} else if err != nil {
		printErr("request", "dbCliConnect", err)
	} else {
		// check if suspended account
		var suspended bool
		err = db.Get(&suspended, "SELECT suspended FROM etabs WHERE id = ?", ifExists)
		if err != nil {
			printErr("insert connection", "dbCliConnect", err)
		}

		// create new auth token
		if !suspended {
			token = uuid.New().String()
			// insert connect data
			_, err = db.Exec("INSERT INTO conections (etab_id, token) VALUES (?, ?)", ifExists, token)
			printErr("insert connection", "dbCliConnect", err)
		} else {
			err = errors.New("suspended account")
		}

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

	err = db.Get(&ifExists, "SELECT etab_id FROM qr_tokens WHERE token = ? ", authToken)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckNcreateSession", err)

	} else if err != nil {
		printErr("request", "dbCheckNcreateSession", err)
	} else {

		// check if suspended account
		var suspended bool
		err = db.Get(&suspended, "SELECT suspended FROM etabs WHERE id = ?", ifExists)
		if err != nil {
			printErr("insert connection", "dbCliConnect", err)
		}

		if !suspended {
			// create new token
			token = uuid.New().String()
			_, err = db.Exec("INSERT INTO conections (etab_id, token, is_admin) VALUES (?, ?, ?)", ifExists, token, 0)
			printErr("insert row", "dbCheckNcreateSession", err)
		} else {
			err = errors.New("suspended account")
		}

	}

	return token, err
}

func dbCheckCliToken(token string) (etabid int, err error) {

	db := dbConnect()

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

func dbCheckOrderCliSess(cli_uuid string, orderid int64) (err error) {
	db := dbConnect()

	var ifExists int

	err = db.Get(&ifExists, "SELECT id FROM orders WHERE cli_uuid = ? AND id = ?", cli_uuid, orderid)

	if ifExists == 0 || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbCheckOrderCliSess", err)

	} else if err != nil {
		printErr("request", "dbCheckOrderCliSess", err)
	}

	return err
}

func dbCheckCliSess(cli_uuid string) (err error) {
	db := dbConnect()

	var ifExists int

	err = db.Get(&ifExists, "SELECT id FROM cli_sess WHERE cli_uuid = ?", cli_uuid)

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

	err = db.Select(&menu.Items, "SELECT items.id, in_stock, items.name, description, price, priceHH, categories.id AS catid, categories.name AS category FROM items JOIN categories on items.category_id = categories.id WHERE items.etab_id = ?", etabid)
	printErr("get row", "dbGetEtabMenu", err)

	return menu, err
}

func dbGetPlanning(etabid int) (planning []*Planning, err error) {

	db := dbConnect()

	err = db.Select(&planning, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ? ORDER BY day asc", etabid)

	printErr("get row", "dbGetPlanning", err)

	return planning, err
}

func dbPlaceOrder(PLOrder eapFact.Order, etabid int, link string) (orderid int64, uuid_fact string, err error) {
	db := dbConnect()

	// check prices
	var count float64
	for _, i := range PLOrder.Order_items {
		var compare CheckOrderItems
		err = db.Get(&compare, "SELECT priceHH, price FROM items WHERE id = ? AND etab_id = ?", i.Item_id, etabid)

		if err != nil {
			printErr("get item", "dbPlaceOrder", err)
			return 0, "", invalidData
		}

		if compare.PriceHH == i.Price {
			count = count + (compare.PriceHH * float64(i.Quantity))
		} else if compare.Price == i.Price {
			count = count + (compare.Price * float64(i.Quantity))
		} else {
			return 0, "", invalidData
		}
	}

	if count != PLOrder.TotalTTC {
		return 0, "", invalidData
	}

	// insert order
	uuid_fact = uuid.New().String()
	insertOrder, err := db.Exec("INSERT INTO orders (cli_uuid, etab_id, fact_link, totalTTC, totalHT) VALUES (?, ?, ?, ?, ?)", PLOrder.Cli_uuid, etabid, link+uuid_fact+".pdf", PLOrder.TotalTTC, PLOrder.TotalHT)
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
				_, err := db.Exec("INSERT INTO order_items (item_id, order_id, price, quantity) VALUES (?, ?, ?, ?)", item.Item_id, orderid, item.Price, item.Quantity)
				printErr("insert row", "dbPlaceOrder", err)
			}
		}
	}

	return orderid, uuid_fact, err
}

func dbUpdateOrderStatus(details OrderDetails) (err error) {
	db := dbConnect()

	// update only confirmed
	_, err = db.Exec("UPDATE orders SET confirmed = ?, ready = ?, done = ? WHERE id = ?", details.Confirmed, details.Ready, details.Done, details.OrderId)
	printErr("update row", "dbUpdateOrderStatus", err)

	return err
}

func dbGetOrders(etabid int) (orders []*ReturnOrders, err error) {

	db := dbConnect()
	orders = []*ReturnOrders{}

	err = db.Select(&orders, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE etab_id = ? AND created > NOW() - interval 2 HOUR", etabid)
	printErr("get rows", "dbGetOrders", err)

	for i := range orders {

		err = db.Select(&orders[i].Order_items, "SELECT order_items.quantity, order_items.price, categories.name AS category, items.name FROM order_items JOIN items ON order_items.item_id = items.id JOIN categories ON items.category_id = categories.id WHERE order_id = ?", orders[i].Id)
		printErr("get rows", "dbGetOrders", err)
	}

	return orders, err
}

func dbGetOrder(orderid int64) (order ReturnOrders, err error) {

	db := dbConnect()
	// orders = ReturnOrders

	err = db.Get(&order, "SELECT id, cli_uuid, totalTTC, totalHT, confirmed, ready, done, created FROM orders WHERE id = ?", orderid)
	printErr("get row", "dbGetOrder", err)

	err = db.Select(&order.Order_items, "SELECT order_items.quantity, order_items.price, categories.name AS category, items.name FROM order_items JOIN items ON order_items.item_id = items.id JOIN categories ON items.category_id = categories.id WHERE order_id = ?", orderid)
	printErr("get rows", "dbGetOrder", err)

	return order, err
}

func dbGetOrderFact(orderid int64) (link string, err error) {
	db := dbConnect()

	var noRow = errors.New("sql: no rows in result set")

	err = db.Get(&link, "SELECT IFNULL(fact_link, '') FROM orders WHERE id = ?", orderid)

	if link == "" || (err != nil && err.Error() == noRow.Error()) {
		printErr("get row", "dbGetOrderFact", err)

	} else if err != nil {
		printErr("request", "dbGetOrderFact", err)
	}

	return link, err

}

func dbGetAllTickets(etabid int64, datemin string, datemax string) (factures []*Factures, err error) {

	db := dbConnect()
	err = db.Select(&factures, "SELECT id, cli_uuid, created, totalTTC, done, IFNULL(fact_link, '') AS fact_link FROM orders WHERE etab_id = ? AND created BETWEEN ? AND ?", etabid, datemin, datemax)

	if err != nil {
		printErr("get tickets", "dbGetAllTickets", err)
	}

	return factures, err
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


	err = db.Get(&params, "SELECT name, phone, addr, cp, city, country, IFNULL(picture, '') AS picture, IFNULL(insta, '') AS insta , IFNULL(twitter, '') AS twitter, IFNULL(facebook, '') AS facebook, licence, siret FROM etabs WHERE etabs.id = ?", etabid)

	if err != nil {
		printErr("get row", "dbGetEtabParams", err)
	} else {

		err = db.Select(&params.Horaires, "SELECT day, start, end, is_active, is_HH FROM planning WHERE etab_id = ?", etabid)
		printErr("get rows", "dbGetEtabParams", err)
	}

	return params, err
}

func dbGetQRs(etabid int64) (qr1 string, qr0 string, err error) {
	db := dbConnect()

	err = db.Get(&qr1, "SELECT CONCAT('"+viper.GetString("links.cdn_qr")+"/bartender/', token, '.png') FROM qr_tokens WHERE etab_id = ? AND type = 1 ", etabid)

	if err != nil {
		printErr("get QR1", "dbGetQRs", err)
	}

	err = db.Get(&qr0, "SELECT CONCAT('"+viper.GetString("links.cdn_qr")+"/menu_qr/', token, '.png') FROM qr_tokens WHERE etab_id = ? AND type = 0 ", etabid)

	if err != nil {
		printErr("get QR0", "dbGetQRs", err)
	}

	return qr0, qr1, err

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

func dbUpdatePic(link string, etabid int64) (path string, err error) {
	db := dbConnect()

	oldpic := ""
	err = db.Get(&oldpic, "SELECT picture FROM etabs WHERE id = ?", etabid)
	if err != nil {
		printErr("get old pic", "dbupdatePic", err)
	}

	err = deleteOldPic(oldpic)
	if err != nil {
		printErr("delete old pic", "dbUpdatePic", err)
	}

	_, err = db.Exec("UPDATE etabs SET picture = ? WHERE id = ?", link, etabid)

	if err != nil {
		printErr("update picture", "dbUpdatePic", err)
	}
	return link, err
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

	_, err = db.Exec("INSERT INTO items (etab_id, in_stock, name, description, price, priceHH, category_id, created) VALUES (?, ?, ?, ?, ?, ?, ?,NOW())", etabid, item.Stock, item.Name, item.Description, item.Price, item.PriceHH, item.Category)
	printErr("insert row", "dbInsertItem", err)
	return err
}

func dbEditItem(item Item) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE items SET in_stock = ?, name = ?, description = ?, price = ?, priceHH = ?, category_id = ?, modified = ? WHERE id = ?", item.Stock, item.Name, item.Description, item.Price, item.PriceHH, item.Category, time.Now(), item.Id)
	printErr("update row", "dbEditItem", err)

	return err
}

func dbDeleteItem(itemid int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("DELETE FROM items WHERE id = ?", itemid)
	printErr("delete row", "dbDeleteItem", err)

	return err
}

func dbGetCategories(etabid int64) (categories []*Categories, err error) {
	db := dbConnect()

	err = db.Select(&categories, "SELECT id, name FROM categories WHERE etab_id = ?", etabid)

	if err != nil {
		printErr("get categories", "dbGetCategories", err)
	}

	return categories, err
}

func dbInsertCategory(etabid int64, category string) (err error) {
	db := dbConnect()

	_, err = db.Exec("INSERT INTO categories (name, etab_id) VALUES (?, ?)", category, etabid)

	if err != nil {
		printErr("insert category", "dbInsertCategory", err)
	}
	return err
}

func dbEditCategory(etabid int64, categoryName string, categoryId int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("UPDATE categories SET name = ? WHERE id = ? AND etab_id = ?", categoryName, categoryId, etabid)

	if err != nil {
		printErr("update category", "dbEditCategory", err)
	}

	return err
}

func dbDeleteCategory(etabid int64, categoryId int64) (err error) {
	db := dbConnect()

	_, err = db.Exec("DELETE FROM categories WHERE id = ? and etab_id = ?", categoryId, etabid)

	if err != nil {
		printErr("delete category", "dbDeleteCategory", err)
	}

	return err
}

func dbUnsub(etabId int64) (data eapFact.FactEtab, date string, err error) {
	db := dbConnect()

	// first get account data
	err = db.Get(&data, "SELECT owner_civility, owner_name, owner_surname, mail, phone, name, fact_addr, fact_cp, fact_country, offer FROM etabs WHERE id = ?", etabId)
	if err != nil {
		printErr("get user data", "dbUnsub", err)

	} else {
		// now unsub
		_, err = db.Exec("UPDATE etabs SET offer = NULL WHERE id = ?", etabId)
		if err != nil {
			printErr("unsubscribe account", "dbUnsub", err)
		}

		// and get delete date
		err = db.Get(&date, "SELECT created_at FROM etabs WHERE id = ?", etabId)
		if err != nil {
			printErr("get account deletion date", "dbUnsub", err)
		}
	}

	return data, date, err
}

func dbEditPlanning(etabid int64, planning []*Planning) (err error) {
	db := dbConnect()


	_, err = db.Exec("DELETE FROM planning WHERE etab_id = ?", etabid)

	if err != nil {
		printErr("delete all rows in planning", "dbEditPlanning", err)

	} else {

		for _, day := range planning {
			_, err = db.Exec("INSERT INTO planning (etab_id, day, start, end, is_active, is_HH) VALUES (?, ?, ? , ?, ?, ?)", etabid, day.Day, day.Start, day.End, day.Is_Active, day.Is_HH)

			if err != nil {
				printErr("insert new row in planning", "dbEditPlanning", err)
				return err
			}
		}
	}
	return err
}

func dbGetEtabInfos(etabid int) (etab eapFact.Infos, err error) {
	db := dbConnect()

	err = db.Get(&etab, "SELECT name, addr, cp, city, country, picture FROM etabs WHERE id = ?", etabid)

	if err != nil {
		printErr("get etab infos", "dbGetEtabInfos", err)
	}
	return etab, err
}

func wsCliAuth(cliId string, orderid string) (err error) {
	db := dbConnect()

	var id int64 

	err = db.Get(&id, "SELECT id FROM orders WHERE cli_uuid = ? AND id = ?", cliId, orderid)

	if err != nil {
		printErr("auth cli", "wsCliAuth", err)
	}
	return err
}

func dbGetOrderStatus(orderid int) (status string, err error) {
	db := dbConnect()

	err = db.Get(&status, "SELECT (CASE WHEN confirmed = 1 AND ready = 0 AND done = 0 THEN 'confirmed' WHEN ready = 1 AND done = 0 THEN 'ready' WHEN done = 1 THEN 'done' END) AS status FROM orders WHERE id = ?", orderid)

	if err != nil {
		printErr("get order status", "dbGetOrderStatus", err)
	}
	return status, err
}