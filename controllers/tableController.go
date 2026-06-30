package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := tableCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching the tables"})
			return
		}

		var tables []bson.M
		if err := result.All(ctx, &tables); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error while getting the data from cursor"})
			return
		}

		c.JSON(http.StatusOK, tables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tableID := c.Param("table_id")

		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"order_id": tableID}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching the table"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		table.CreatedAt = time.Now()
		table.UpdatedAt = time.Now()
		table.ID = bson.NewObjectID()
		table.TableID = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctx, table)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create a table"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var table models.Table

		tableID := c.Param("table_id")

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updatedObj bson.D

		if table.NumberOfGuests != nil {
			updatedObj = append(updatedObj, bson.E{"number_of_guests", table.NumberOfGuests})
		}

		if table.TableNumber != nil {
			updatedObj = append(updatedObj, bson.E{"table_number", table.TableNumber})
		}

		table.UpdatedAt = time.Now()

		opt := options.UpdateOne().SetUpsert(true)

		filter := bson.M{"table_id": tableID}

		result, err := tableCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updatedObj},
			},
			opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update the table"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
