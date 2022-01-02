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
	router.GET("menu", getMenu)
	router.GET("planning", getPlanning)
	router.GET("orders", GetOrders)
	router.GET("order", GetOrder)
	router.GET("sendmail-fact", sendFact)
	router.GET("fact-link", factLink)
	router.GET("worker-fact", bossFact) // TODO: make fact
	router.GET("etab-params", getEtabParams)
	router.GET("profile", getProfile)
	router.GET("payment-method", getPaymentMethod)

	//POST
	// subscribe
	router.POST("subscribe", Subscribe) // TODO: make fact
	// connect
	router.POST("connect", Connect)
	router.POST("bartender", QRConnect)
	// password creation
	router.POST("pwd-create", createPWD)
	router.POST("sendMail4reset-pwd", SM4resetPWD)
	// place order
	router.POST("place-order", placeOrder) // TODO: make fact

	//PUT
	router.PUT("update-order", updateOrderStatus)
	router.PUT("etab-params", EditEtabParams)
	router.PUT("profile", EditProfile)
	router.PUT("payment-method", EditPaymentMethod)

	//DELETE
	router.DELETE("reset-all-connections", deleteAllconn)
	router.DELETE("disconnect", disconnect)

	// Cli part

	// Run
	router.Run(":9999")
}
