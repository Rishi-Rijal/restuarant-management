package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rishi-rijal/restuarant-management/database"
	"github.com/rishi-rijal/restuarant-management/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrderItemPack struct {
	TableID    *string
	OrderItems []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := orderItemCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching list items"})
			return
		}

		var orderItems []bson.M

		if err := result.All(ctx, &orderItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching itesm"})
			return
		}

		c.JSON(http.StatusOK, orderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		orderItemID := c.Param("order_item_id")

		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctx, bson.M{"orderItemID": orderItemID}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the order item"})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

func GetOrderItemByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error occured while fetching orders by order id"})
			return
		}

		c.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []bson.M, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "order_id", Value: id},
	}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "food"},
		{Key: "localField", Value: "food_id"},
		{Key: "foreignField", Value: "food_id"},
		{Key: "as", Value: "food"},
	}}}
	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "$path", Value: "$food"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "order"},
		{Key: "localField", Value: "order_id"},
		{Key: "foreignField", Value: "order_id"},
		{Key: "as", Value: "order"},
	}}}
	unwindOrderStage := bson.D{
		{
			Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$order"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			},
		},
	}
	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "table"},
		{Key: "localField", Value: "order.table_id"},
		{Key: "foreignField", Value: "table_id"},
		{Key: "as", Value: "table"},
	}}}
	unwindTableStage := bson.D{{Key: "unwind", Value: bson.D{
		{Key: "$path", Value: "$table"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	projectStage := bson.D{
		{
			Key: "$project", Value: bson.D{
				{Key: "id", Value: 0},
				{Key: "amount", Value: "$food.price"},
				{Key: "total_count", Value: 1},
				{Key: "food_name", Value: "$food.name"},
				{Key: "food_image", Value: "$food.food_image"},
				{Key: "table_number", Value: "$table.table_number"},
				{Key: "table_id", Value: "$table.table_id"},
				{Key: "order_id", Value: "$order.order_id"},
				{Key: "price", Value: "$food.price"},
				{Key: "quantity", Value: 1},
			},
		},
	}

	groupStage := bson.D{
		{
			Key: "$group", Value: bson.D{
				{
					Key: "_id", Value: bson.D{
						{Key: "order_id", Value: "$order_id"},
						{Key: "table_id", Value: "$table_id"},
						{Key: "table_number", Value: "$table_number"},
					},
				},
				{
					Key: "payment_due",
					Value: bson.D{
						{Key: "$sum", Value: "$amount"},
					},
				},
				{
					Key: "total_count",
					Value: bson.D{
						{Key: "$sum", Value: 1},
					},
				},
				{
					Key: "order_items",
					Value: bson.D{
						{Key: "$push", Value: 1},
					},
				},
			},
		},
	}

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		}},
	}

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	})
	if err != nil {
		return []bson.M{}, err
	}

	if err := result.All(ctx, &OrderItems); err != nil {
		return []bson.M{}, err
	}

	return OrderItems, err
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		var OrderItem models.OrderItem
		var orderItemPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.OrderDate = time.Now()

		var orderItemsToBeInserted []interface{}
		order.TableID = orderItemPack.TableID

		orderID, err := OrderItemOrderCreator(order)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "someting went wrong"})
		}
		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = orderID
			validationErr := validate.Struct(orderItem)

			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}
			OrderItem.ID = bson.NewObjectID()
			orderItem.CreatedAt = time.Now()
			orderItem.UpdatedAt = time.Now()
			orderItem.OrderItemID = OrderItem.ID.Hex()

			num := toFixed(float64(*orderItem.UnitPrice), 2)
			OrderItem.UnitPrice = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)

		}

		insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong while creating orderItem"})
			return
		}

		c.JSON(http.StatusOK, insertedOrderItems)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var orderItem models.OrderItem

		orderItemID := c.Param("order_item_id")

		filter := bson.M{"order_item_id": orderItemID}

		var updateObj bson.D

		if orderItem.UnitPrice != nil {
			updateObj = append(updateObj, bson.E{"unit_price", orderItem.UnitPrice})
		}

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", orderItem.Quantity})
		}

		if orderItem.FoodID != nil {
			updateObj = append(updateObj, bson.E{"food_id", orderItem.FoodID})
		}

		orderItem.UpdatedAt = time.Now()

		updateObj = append(updateObj, bson.E{"updated_at", orderItem.UpdatedAt})

		opt := options.UpdateOne().SetUpsert(true)

		result, err := orderItemCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}},
			opt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "order item update failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
