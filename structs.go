package main

type Subscription struct {
	Civility     string `json:"civility"`
	Name         string `json:"name"`
	Surname      string `json:"surname"`
	Mail         string `json:"mail"`
	Phone        string `json:"phone"`
	Offer        int    `json:"offer"`
	Entname      string `json:"entname"`
	Siret        string `json:"siret"`
	Licence      string `json:"licence"`
	Addr         string `json:"addr"`
	Cp           int    `json:"cp"`
	City         string `json:"city"`
	Country      string `json:"country"`
	Iban         string `json:"iban"`
	Name_iban    string `json:"name_iban"`
	Fact_addr    string `json:"fact_addr"`
	Fact_cp      int    `json:"fact_cp"`
	Fact_city    string `json:"fact_city"`
	Fact_country string `json:"fact_country"`
}

type Owner struct {
	Civility string `db:"owner_civility"`
	Name     string `db:"owner_name"`
	Surname  string `db:"owner_surname"`
	Mail     string `db:"mail"`
	Entname  string `db:"name"`
	Siret    string `db:"siret"`
	Addr     string `db:"addr"`
	Cp       int    `db:"cp"`
	City     string `db:"city"`
	Country  string `db:"country"`
}

type PWD struct {
	Id               int    `json:"id"`
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
}
