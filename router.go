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
	// router.GET(":etabid/orders", GetOrders)

	//POST
	// subscribe
	router.POST("subscribe", Subscribe)
	// connect
	// router.POST("connect", Connect)
	// password creation
	// router.POST("pwd-create/:etabid", createPWD)

	//PUT
	// password edit
	// router.PUT("pwd-edit/:etabid", editPWD)

	//DELETE

	// Cli part

	// Run
	router.Run(":9999")
}
