// Package database provides analytics repository.
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/vknow360/otaship/backend/internal/models"
)

// AnalyticsRepository handles download logging and statistics.
type AnalyticsRepository struct {
	collection *mongo.Collection
}

// NewAnalyticsRepository creates a new analytics repository.
func NewAnalyticsRepository(db *MongoDB) *AnalyticsRepository {
	if db == nil {
		return nil
	}
	return &AnalyticsRepository{
		collection: db.Downloads,
	}
}

// LogDownload records a download event.
func (r *AnalyticsRepository) LogDownload(ctx context.Context, log *models.DownloadLog) error {
	if r == nil {
		return nil // Silently ignore if no database
	}

	log.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// GetSummary returns aggregated analytics.
func (r *AnalyticsRepository) GetSummary(ctx context.Context) (*models.AnalyticsSummary, error) {
	if r == nil {
		return &models.AnalyticsSummary{
			ByPlatform:       make(map[string]int64),
			ByChannel:        make(map[string]int64),
			ByRuntimeVersion: make(map[string]int64),
		}, nil
	}

	summary := &models.AnalyticsSummary{
		ByPlatform:       make(map[string]int64),
		ByChannel:        make(map[string]int64),
		ByRuntimeVersion: make(map[string]int64),
	}

	// Total downloads
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count downloads: %w", err)
	}
	summary.TotalDownloads = total

	// Today's downloads
	startOfDay := time.Now().Truncate(24 * time.Hour)
	todayCount, err := r.collection.CountDocuments(ctx, bson.M{
		"timestamp": bson.M{"$gte": startOfDay},
	})
	if err == nil {
		summary.TodayDownloads = todayCount
	}

	// This week's downloads
	startOfWeek := time.Now().AddDate(0, 0, -7)
	weekCount, err := r.collection.CountDocuments(ctx, bson.M{
		"timestamp": bson.M{"$gte": startOfWeek},
	})
	if err == nil {
		summary.WeekDownloads = weekCount
	}

	// Group by platform
	platformPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":   "$platform",
			"count": bson.M{"$sum": 1},
		}}},
	}
	platformCursor, err := r.collection.Aggregate(ctx, platformPipeline)
	if err == nil {
		var results []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := platformCursor.All(ctx, &results); err == nil {
			for _, r := range results {
				summary.ByPlatform[r.ID] = r.Count
			}
		}
	}

	return summary, nil
}

// GetDownloadsForUpdate returns download count for a specific update.
func (r *AnalyticsRepository) GetDownloadsForUpdate(ctx context.Context, updateID string) (int64, error) {
	if r == nil {
		return 0, nil
	}

	return r.collection.CountDocuments(ctx, bson.M{"updateId": updateID})
}
