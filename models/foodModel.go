package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Food struct {
	ID        bson.ObjectID `bson:"_id"`
	Name      *string       `json:"name" validate:"required,min=2,max=100"`
	Price     *float64      `json:"price" validate:"required"`
	FoodImage *string       `json:"food_image" validate:"required"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	FoodID    string        `json:"food_id"`
	MenuID    *string       `json:"menu_id" validate:"required"`
}
