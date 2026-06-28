package routes

import "github.com/gin-gonic/gin"

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/orders", controller.GetTables())
	incomingRoutes.GET("/orders/order_id", controller.GetTable())
	incomingRoutes.POST("/orders", controller.CreateTable())
	incomingRoutes.PATCH("/orders/order_id", controller.UpdateTable())
}
