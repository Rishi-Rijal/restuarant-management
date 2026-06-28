package controller

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func getOrderItemByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func ItemsByOrder(id string) (OrderItems []bson.M, err error) {
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
