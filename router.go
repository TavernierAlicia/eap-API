package main

import (
	limit "github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/gin"
)

func CORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(200)
	}
}

func Router() {
	// init router
	router := gin.Default()
	router.Use(limit.MaxAllowed(20))
	router.Use(CORS)

	//GET
	router.GET("get-etabs4reset-pwd", getEtabs)
	router.GET("menu", getMenu)
	router.GET("menu-cli/:etab", getMenuCli)
	router.GET("planning/:etab", getPlanning)
	router.GET("orders", getOrders)
	router.GET("order/:etab/:order_id", getOrder)
	router.GET("get-fact", sendFact)
	router.GET("fact-link", getFactLink)
	router.GET("worker-fact", getBossFact) // TODO: make fact
	router.GET("etab-params", getEtabParams)
	router.GET("profile", getProfile)
	router.GET("payment-method", getPaymentMethod)
	router.GET("offers", getEtabOffer)
	router.GET("csv", getCSV)
	router.GET("categories", getCategories)
	router.GET("all-tickets", getAllTickets)
	router.GET("get-qrs", getQRs)


	//POST
	// subscribe
	router.POST("subscribe", Subscribe)
	router.POST("contact", Send)
	// connect
	router.POST("connect", Connect)
	router.POST("bartender", QRConnect)
	router.POST("qr/:token", Cli)
	// password creation
	router.POST("pwd-create", createPWD)
	router.POST("sendMail4reset-pwd", SM4resetPWD)
	// place order
	router.POST("place-order", placeOrder)
	router.POST("item", postItem)
	router.POST("categories", postCategory)

	//PUT
	router.PUT("update-order", updateOrderStatus)
	router.PUT("etab-params", editEtabParams)
	router.PUT("profile", editProfile)
	router.PUT("payment-method", editPaymentMethod)
	router.PUT("offers", editOffers)
	router.PUT("item", putItem)
	router.PUT("categories", putCategory)
	router.PUT("unsubscribe", unsub)
	router.PUT("update-planning", updatePlanning)
	router.PUT("update-pic", updateEtabPic)

	//DELETE
	router.DELETE("reset-all-connections", deleteAllconn)
	router.DELETE("disconnect", disconnect)
	router.DELETE("item", deleteItem)
	router.DELETE("categories", deleteCategory)

	// Run
	router.Run(":9999")
}
