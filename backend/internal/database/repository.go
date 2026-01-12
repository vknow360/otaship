// Package database provides MongoDB repository for updates.
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/vknow360/otaship/backend/internal/models"
)

// UpdateRepository handles CRUD operations for updates.
type UpdateRepository struct {
	collection *mongo.Collection
}

// NewUpdateRepository creates a new update repository.
func NewUpdateRepository(db *MongoDB) *UpdateRepository {
	if db == nil {
		return nil
	}
	return &UpdateRepository{
		collection: db.Updates,
	}
}

// Create inserts a new update into the database.
func (r *UpdateRepository) Create(ctx context.Context, update *models.Update) error {
	if r == nil {
		return fmt.Errorf("database not connected")
	}

	update.CreatedAt = time.Now()
	if update.RolloutPercentage == 0 {
		update.RolloutPercentage = 100
	}
	if update.Channel == "" {
		update.Channel = models.ChannelProduction
	}

	result, err := r.collection.InsertOne(ctx, update)
	if err != nil {
		return fmt.Errorf("failed to insert update: %w", err)
	}

	update.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID retrieves an update by its MongoDB ObjectID.
func (r *UpdateRepository) FindByID(ctx context.Context, id string) (*models.Update, error) {
	if r == nil {
		return nil, fmt.Errorf("database not connected")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid update ID: %w", err)
	}

	var update models.Update
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find update: %w", err)
	}

	return &update, nil
}

// FindByUpdateID retrieves an update by its UUID string (updateId field).
func (r *UpdateRepository) FindByUpdateID(ctx context.Context, updateID string) (*models.Update, error) {
	if r == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var update models.Update
	err := r.collection.FindOne(ctx, bson.M{"updateId": updateID}).Decode(&update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find update: %w", err)
	}

	return &update, nil
}

// FindLatest retrieves the latest active update for a project, runtime version and channel.
func (r *UpdateRepository) FindLatest(ctx context.Context, projectSlug, runtimeVersion, channel, platform string) (*models.Update, error) {
	if r == nil {
		return nil, fmt.Errorf("database not connected")
	}

	filter := bson.M{
		"projectSlug":    projectSlug,
		"runtimeVersion": runtimeVersion,
		"channel":        channel,
		"isActive":       true,
	}

	// Platform filter (match specific or "all")
	if platform != "" {
		filter["$or"] = []bson.M{
			{"platform": platform},
			{"platform": models.PlatformAll},
		}
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	var update models.Update
	err := r.collection.FindOne(ctx, filter, opts).Decode(&update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest update: %w", err)
	}

	return &update, nil
}

// FindAll retrieves all updates with optional filters.
func (r *UpdateRepository) FindAll(ctx context.Context, filter bson.M, limit, offset int64) ([]*models.Update, int64, error) {
	if r == nil {
		return nil, 0, fmt.Errorf("database not connected")
	}

	if filter == nil {
		filter = bson.M{}
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count updates: %w", err)
	}

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find updates: %w", err)
	}
	defer cursor.Close(ctx)

	var updates []*models.Update
	if err := cursor.All(ctx, &updates); err != nil {
		return nil, 0, fmt.Errorf("failed to decode updates: %w", err)
	}

	return updates, total, nil
}

// Update modifies an existing update.
func (r *UpdateRepository) Update(ctx context.Context, id string, update bson.M) error {
	if r == nil {
		return fmt.Errorf("database not connected")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid update ID: %w", err)
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("update not found")
	}

	return nil
}

// IncrementDownloads increments the download counter for an update.
func (r *UpdateRepository) IncrementDownloads(ctx context.Context, updateID string) error {
	if r == nil {
		return nil // Silently ignore if no database
	}

	objectID, err := primitive.ObjectIDFromHex(updateID)
	if err != nil {
		return nil // Ignore invalid IDs
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"downloads": 1}},
	)
	return err
}

// Deactivate marks an update as inactive.
func (r *UpdateRepository) Deactivate(ctx context.Context, id string) error {
	return r.Update(ctx, id, bson.M{"isActive": false})
}

// SetRollout updates the rollout percentage.
func (r *UpdateRepository) SetRollout(ctx context.Context, id string, percentage int) error {
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	return r.Update(ctx, id, bson.M{"rolloutPercentage": percentage})
}

// Delete permanently removes an update.
func (r *UpdateRepository) Delete(ctx context.Context, id string) error {
	if r == nil {
		return fmt.Errorf("database not connected")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid update ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("update not found")
	}

	return nil
}

// DeleteByProjectSlug deletes all updates for a specific project.
func (r *UpdateRepository) DeleteByProjectSlug(ctx context.Context, projectSlug string) error {
	if r == nil {
		return fmt.Errorf("database not connected")
	}

	_, err := r.collection.DeleteMany(ctx, bson.M{"projectSlug": projectSlug})
	if err != nil {
		return fmt.Errorf("failed to delete project updates: %w", err)
	}

	return nil
}
