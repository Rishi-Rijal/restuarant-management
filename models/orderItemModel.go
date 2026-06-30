package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderItem struct {
	ID          bson.ObjectID `bson:"_id"`
	Quantity    *int          `json:"quantity" validate:"required,eq=S|eq=M|eq=L"`
	UnitPrice   *float64      `json:"unit_price" validate:"required"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	FoodID      *string       `json:"food_id" validate:"required"`
	OrderItemID string        `json:"order_item_id"`
	OrderID     string        `json:"order_id" validate:"required"`
}
