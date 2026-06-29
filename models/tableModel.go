package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Table struct {
	ID             bson.ObjectID `bson:"_id"`
	NumberOfGuests *int          `json:"number_of_guests" validate:"required"`
	TableNumber    *int          `json:"table_number" validate:"required"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	TableID        *string       `json:"table_id"`
}
