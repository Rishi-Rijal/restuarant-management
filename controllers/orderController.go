package controller

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rishi-rijal/restuarant-management/database"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")
	tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")
)

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := orderCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occured while listing the orders"})
			return
		}

		var allOrders []bson.M

		if err := result.All(ctx, &allOrders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderID := c.Param("order_id")
		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"orderId": orderID}).Decode(&order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fteching the order"})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return

		}

		if order.TableID != nil {
			err := tableCollection.FindOne(ctx, bson.M{"tableID": order.TableID}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "table not found"})
				return
			}
		}

		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()

		order.ID = bson.NewObjectID()

		order.OrderID = order.ID.Hex()

		result, insertErr := orderCollection.InsertOne(ctx, order)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating a order"})
			return
		}

		c.JSON(http.StatusInternalServerError, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var table models.Table
		var order models.Order

		var updateObj bson.D

		orderID := c.Param("order_id")

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if order.TableID != nil {
			err := orderCollection.FindOne(ctx, bson.M{"tableId": order.TableID}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "table not found"})
				return
			}

			updateObj = append(updateObj, bson.E{"order", order.TableID})
		}

		order.UpdatedAt = time.Now()
		updateObj = append(updateObj, bson.E{"updatedAt", order.UpdatedAt})

		filter := bson.M{"orderID": orderID}

		opt := options.UpdateOne().SetUpsert(true)

		result, err := orderCollection.UpdateOne(ctx, filter, bson.D{{"$st", updateObj}}, opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order updation failed"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) (string, error) {
	ctx, err := context.WithTimeout(context.Background(), 10*time.Second)
	if err != nil {
		return "", errors.New("something went wrong")
	}

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.ID = bson.NewObjectID()
	order.OrderID = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)

	return order.OrderID, nil
}
