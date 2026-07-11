// Package controller provides all the controller functions that are required for the
// food endpoints
package controller

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rishi-rijal/restuarant-management/database"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
	validate                         = validator.New()
)
var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("record_per_page"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		matchState := bson.D{{
			Key: "$match", Value: bson.D{{}},
		}}
		groupState := bson.D{
			{
				Key: "$group", Value: bson.D{
					{
						Key: "_id", Value: bson.D{
							{Key: "_id", Value: nil},
						},
					},
					{
						Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}},
					},
					{
						Key: "data", Value: bson.D{
							{Key: "$push", Value: "$$ROOT"},
						},
					},
				},
			},
		}

		projectStage := bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "total_count", Value: 1},
			{Key: "food_items", Value: bson.D{
				{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
			}},
		}}}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchState, groupState, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while lising food items"})
			return
		}

		var allFoods []bson.M

		if err := result.All(ctx, &allFoods); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error listing food itesm"})
			return
		}

		c.JSON(http.StatusOK, allFoods)
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		foodID := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the food"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
		if err != nil {
			msg := "menu was not found"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		food.CreatedAt = time.Now()
		food.UpdatedAt = time.Now()

		food.ID = bson.NewObjectID()
		food.FoodID = food.ID.Hex()
		toFixed(*food.Price, 2)

		result, insertErr := foodCollection.InsertOne(ctx, food)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": insertErr.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var food models.Food
		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		foodID := c.Param("food_id")

		filter := bson.M{"food_id": foodID}

		var updateObj bson.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.FoodImage != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.FoodImage})
		}

		if food.MenuID != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.MenuID}).Decode(&menu)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Menu was not found"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: menu.ID})
		}

		food.UpdatedAt = time.Now()
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.UpdatedAt})

		opt := options.UpdateOne().SetUpsert(true)

		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "food update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
