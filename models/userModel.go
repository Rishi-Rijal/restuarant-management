package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID           bson.ObjectID `bson:"id"`
	FirstName    *string       `json:"first_name" validate:"required"`
	LastName     *string       `json:"last_name" validate:"required,min=2,max=100"`
	Password     *string       `json:"password" validate:"required,min=2"`
	Email        *string       `json:"email" validate:"required,email"`
	Avatar       *string       `json:"avatar"`
	Phone        *string       `json:"phone" validate:"required"`
	Token        *string       `json:"token"`
	RefreshToken *string       `json:"refresh_token"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	UserID       string        `json:"user_id"`
}
