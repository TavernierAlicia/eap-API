package main

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/google/uuid"

)

var clients = []Clients{}
var boss = []Clients{}

func Websocket() {

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context){
		cliId := c.Query("cliid")
		orderId := c.Query("orderid")

		if orderId == "" || cliId == "" {
			return
		} else {
			err := wsCliAuth(cliId, orderId)
			if err != nil {
				return
			} else {
				wshandler(c.Writer, c.Request, cliId, orderId)
			}
		}
	})

	r.GET("/orders", func(c *gin.Context){
		etabid := c.Query("etabid")
		token := c.Query("authorization")
		_, err := dbCheckToken(token)
		if err != nil {
			ret401(c)
		}

		uuid := uuid.New().String()

		wshandler(c.Writer, c.Request, uuid, etabid)

	})

	r.POST("/new-order/:eid", func (c *gin.Context){
		etabid := c.Param("eid")

		for _, cli := range clients {
			if cli.OrderId == etabid {
				cli.Co.WriteMessage(1, []byte("{\"event\": \"NEW\"}"))
				c.JSON(200, "new")
			}
		} 

	})

	r.POST("/update-order", func(c *gin.Context){
		var orderData WSData
		c.BindJSON(&orderData)
		
		for _, cli := range clients {
			if cli.OrderId == orderData.OrderId {
				cli.Co.WriteMessage(1, []byte("{\"status\": \""+orderData.Status+"\"}"))
				c.JSON(200, "send")
			}
		}
	})

	r.Run(":9998")
}


var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request, cliId string, orderId string) {
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}

	client := Clients {Uuid: cliId, OrderId: orderId, Co: conn}
	clients = append(clients, client)
	

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			remove(cliId)
			break
		}
	}
}



func remove(cliId string) {

	for i, cli := range clients {
		if cli.Uuid == cliId {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

