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
