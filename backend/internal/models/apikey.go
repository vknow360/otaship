package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APIKey represents an access key for the OTAShip server.
type APIKey struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	KeyHash    string             `bson:"keyHash" json:"-"` // Not exposed in JSON
	Prefix     string             `bson:"prefix" json:"prefix"`
	Scopes     []string           `bson:"scopes" json:"scopes"` // e.g. "read", "write", "admin"
	LastUsedAt time.Time          `bson:"lastUsedAt" json:"lastUsedAt"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
