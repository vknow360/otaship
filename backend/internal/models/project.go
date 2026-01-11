package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Project represents an Expo app project.
type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug        string             `bson:"slug" json:"slug"`               // URL-safe identifier from app.json expo.slug
	Name        string             `bson:"name" json:"name"`               // Display name
	Description string             `bson:"description" json:"description"` // Optional description
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdateCount int                `bson:"updateCount" json:"updateCount"` // Cached count
}
