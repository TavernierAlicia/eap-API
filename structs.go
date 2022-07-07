package main

import (
	"github.com/gorilla/websocket"
)

type JSONTODATA struct {
	Mail          string `json:"mail"`
	EtabID        int64  `json:"etab_id"`
	OfferID       int64  `json:"offer_id"`
	ItemID        int64  `json:"item_id"`
	Category_id   int64  `json:"id"`
	Category_name string `json:"name"`
}

type PWD struct {
	Token            string `json:"token"`
	Password         string `json:"password"`
	Confirm_password string `json:"password-confirm"`
}

type ClientConn struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

type Etab struct {
	Id      int    `db:"id"`
	Name    string `db:"name"`
	Siret   string `db:"siret"`
	Addr    string `db:"addr"`
	Cp      int    `db:"cp"`
	City    string `db:"city"`
	Country string `db:"country"`
	Items   []*Menu
}

type ServQRToken struct {
	Token string `json:"token"`
}

type Menu struct {
	Id       int     `db:"id"`
	Stock    bool    `db:"in_stock"`
	Name     string  `db:"name"`
	Desc     string  `db:"description"`
	Price    float64 `db:"price"`
	HHPrice  float64 `db:"priceHH"`
	CategoryID int   `db:"catid"`
	Category string  `db:"category"`
}

type Planning struct {
	Day       int  `db:"day"`
	Start     int  `db:"start"`
	End       int  `db:"end"`
	Is_Active bool `db:"is_active"`
	Is_HH     bool `db:"is_HH"`
}


type CheckOrderItems struct {
	Price    float64 `db:"price"`
	PriceHH float64  `db:"priceHH"`
	Quantity int     `db:"quantity"`
}

type OrderDetails struct {
	OrderId   int    `json:"order_id"`
	Confirmed bool   `json:"confirmed"`
	Ready     bool   `json:"ready"`
	Done      bool   `json:"done"`
}

type ReturnOrders struct {
	Id          int     `db:"id"`
	Cli_uuid    string  `db:"cli_uuid"`
	TotalTTC    float64 `db:"totalTTC"`
	TotalHT     float64 `db:"totalHT"`
	Confirmed   bool    `db:"confirmed"`
	Ready       bool    `db:"ready"`
	Done        bool    `db:"done"`
	Date        string  `db:"created"`
	Order_items []*Items
}

type Items struct {
	Name     string  `db:"name"`
	Quantity int     `db:"quantity"`
	Price    float64 `db:"price"`
	Category string  `db:"category"`
}

type EtabParams struct {
	Etab_name string `db:"name"`
	Addr      string `db:"addr"`
	Cp        int    `db:"cp"`
	City      string `db:"city"`
	Country   string `db:"country"`
	Insta     string `db:"insta"`
	Twitter   string `db:"twitter"`
	Facebook  string `db:"facebook"`
	License   string `db:"licence"`
	Siret     string `db:"siret"`
	Phone     string `db:"phone"`
	Pic string `db:"picture"`
	Horaires  []*Planning
}

type Profile struct {
	Civility string `db:"owner_civility"`
	Name     string `db:"owner_name"`
	Surname  string `db:"owner_surname"`
	Mail     string `db:"mail"`
}

type Payment struct {
	Iban         string `db:"iban"`
	Name_iban    string `db:"name_iban"`
	Fact_addr    string `db:"fact_addr"`
	Fact_cp      int    `db:"fact_cp"`
	Fact_city    string `db:"fact_city"`
	Fact_country string `db:"fact_country"`
}

type Item struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Stock       bool    `json:"in_stock"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	PriceHH     float64 `json:"priceHH"`
	Category    string  `json:"category"`
}

type Categories struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type Factures struct {
	Id int64 `db:"id"`
	Cli_uuid string `db:"cli_uuid"`
	Date string `db:"created"`
	Total float64 `db:"totalTTC"`
	IsDone bool `db:"done"`
	Link string `db:"fact_link"`
}


type Clients struct {
	Uuid string
	OrderId string
	Co *websocket.Conn
}

type WSData struct {
	OrderId string `json:"orderid"`
	Status string `json:"status"`
}