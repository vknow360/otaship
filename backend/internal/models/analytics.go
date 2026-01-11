// Package models defines the data structures for the application.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DownloadLog tracks individual update downloads for analytics.
type DownloadLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UpdateID  string             `bson:"updateId" json:"updateId"`
	Platform  string             `bson:"platform" json:"platform"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Success   bool               `bson:"success" json:"success"`

	// Optional device info (hashed for privacy)
	DeviceHash string `bson:"deviceHash,omitempty" json:"deviceHash,omitempty"`
}

// AnalyticsSummary provides aggregated statistics.
type AnalyticsSummary struct {
	TotalDownloads   int64            `json:"totalDownloads"`
	TodayDownloads   int64            `json:"todayDownloads"`
	WeekDownloads    int64            `json:"weekDownloads"`
	ByPlatform       map[string]int64 `json:"byPlatform"`
	ByChannel        map[string]int64 `json:"byChannel"`
	ByRuntimeVersion map[string]int64 `json:"byRuntimeVersion"`
}
