package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/vknow360/otaship/backend/internal/models"
)

type APIKeyRepository struct {
	collection *mongo.Collection
}

func NewAPIKeyRepository(db *MongoDB) *APIKeyRepository {
	if db == nil {
		return &APIKeyRepository{collection: nil}
	}
	return &APIKeyRepository{
		collection: db.APIKeys,
	}
}

// Create generates a new API key and stores its hash.
// Returns the plaintext key (only time it's visible) and the model.
func (r *APIKeyRepository) Create(ctx context.Context, name string, scopes []string) (string, *models.APIKey, error) {
	if r.collection == nil {
		return "", nil, fmt.Errorf("database not connected")
	}

	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, err
	}
	plainKey := "ota_" + hex.EncodeToString(keyBytes)

	// Hash key
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := &models.APIKey{
		ID:         primitive.NewObjectID(),
		Name:       name,
		KeyHash:    keyHash,
		Prefix:     plainKey[:8], // "ota_" + 4 chars
		Scopes:     scopes,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Time{}, // Zero time
	}

	_, err := r.collection.InsertOne(ctx, apiKey)
	if err != nil {
		return "", nil, err
	}

	return plainKey, apiKey, nil
}

// FindAll returns all API keys (without hashes).
func (r *APIKeyRepository) FindAll(ctx context.Context) ([]*models.APIKey, error) {
	if r.collection == nil {
		return []*models.APIKey{}, nil
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []*models.APIKey
	if err := cursor.All(ctx, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// Delete removes a key by ID.
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	if r.collection == nil {
		return fmt.Errorf("database not connected")
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// Validate checks if a key is valid and updates LastUsedAt.
// Returns the key model if valid.
func (r *APIKeyRepository) Validate(ctx context.Context, plainKey string) (*models.APIKey, error) {
	if r.collection == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Hash the input key
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	// Find key
	var apiKey models.APIKey
	err := r.collection.FindOne(ctx, bson.M{"keyHash": keyHash}).Decode(&apiKey)
	if err != nil {
		return nil, err // Not found or error
	}

	// Update LastUsedAt (fire and forget / async to not block read)
	go func() {
		ctxBg := context.Background()
		r.collection.UpdateOne(ctxBg, bson.M{"_id": apiKey.ID}, bson.M{
			"$set": bson.M{"lastUsedAt": time.Now()},
		})
	}()

	return &apiKey, nil
}
