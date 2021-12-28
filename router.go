package main

import (
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
)

func Router() {
	// init router
	router := gin.Default()
	router.Use(limit.MaxAllowed(20))

	// Bar part

	//GET
	// orders
	router.GET("get-etabs4reset-pwd", getEtabs)
	// router.GET(":etabid/orders", GetOrders)

	//POST
	// subscribe
	router.POST("subscribe", Subscribe)
	// connect
	router.POST("connect", Connect)
	// router.POST("QRconnect", Connect)
	// password creation
	router.POST("pwd-create", createPWD)
	router.POST("sendMail4reset-pwd", SM4resetPWD)

	//PUT
	// password edit
	//DELETE
	router.DELETE("reset-all-connections", deleteAllconn)
	router.DELETE("disconnect", disconnect)

	// Cli part

	// Run
	router.Run(":9999")
}
