package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := menuCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while listing the menu item"})
			return
		}
		var allMenus []bson.M

		if err = result.All(ctx, &allMenus); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allMenus)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		menuID := c.Param("menu_id")
		var menu models.Menu

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuID}).Decode(&menu)
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		valitationErr := validate.Struct(menu)

		if valitationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": valitationErr.Error()})
			return
		}

		menu.CreatedAt = time.Now()
		menu.UpdatedAt = time.Now()
		menu.ID = bson.NewObjectID()

		menu.MenuID = menu.ID.Hex()

		result, insertErr := menuCollection.InsertOne(ctx, menu)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Menu item not created"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menuID := c.Param("menu_id")

		filter := bson.M{"menu_id": menuID}

		var updateObj bson.D

		if menu.StartDate != nil && menu.EndDate != nil {
			if !inTimeSpan(*menu.StartDate, *menu.EndDate, time.Now()) {
				msg := "kindly send the time"

				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "startDate", Value: menu.StartDate})
			updateObj = append(updateObj, bson.E{Key: "endDate", Value: menu.EndDate})

		}

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		}

		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
		}

		menu.UpdatedAt = time.Now()
		updateObj = append(updateObj, bson.E{Key: "updatedAt", Value: menu.UpdatedAt})

		opt := options.UpdateOne().SetUpsert(true)

		result, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "menu update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func inTimeSpan(startDate time.Time, endDate time.Time, baseTime time.Time) bool {
	return !baseTime.Before(startDate) && !baseTime.After(endDate)
}
