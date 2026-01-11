// Package services contains business logic for the application.
package services

import (
	"math/rand"
	"sync"
	"time"
)

// RolloutService handles percentage-based gradual rollouts.
type RolloutService struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewRolloutService creates a new rollout service with a seeded RNG.
func NewRolloutService() *RolloutService {
	return &RolloutService{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldReceiveUpdate determines if a request should receive an update
// based on the rollout percentage.
//
// percentage: 0-100, where 100 means everyone gets the update.
// deviceHash: Optional device identifier for consistent rollouts.
//
// Returns true if the device should receive the update.
func (r *RolloutService) ShouldReceiveUpdate(percentage int, deviceHash string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 100% rollout - everyone gets it
	if percentage >= 100 {
		return true
	}

	// 0% rollout - no one gets it
	if percentage <= 0 {
		return false
	}

	// If device hash is provided, use consistent hashing
	// This ensures the same device always gets the same result
	if deviceHash != "" {
		return r.deterministicRollout(percentage, deviceHash)
	}

	// Random rollout
	return r.rng.Intn(100) < percentage
}

// deterministicRollout provides consistent rollout based on device hash.
// The same device will always get the same result for the same percentage.
func (r *RolloutService) deterministicRollout(percentage int, deviceHash string) bool {
	// Hash the device ID to get a number between 0-99
	var hashSum int
	for _, c := range deviceHash {
		hashSum += int(c)
	}
	bucket := hashSum % 100

	return bucket < percentage
}

// CalculateRolloutBucket returns which rollout bucket (0-99) a device falls into.
// Useful for debugging and analytics.
func CalculateRolloutBucket(deviceHash string) int {
	var hashSum int
	for _, c := range deviceHash {
		hashSum += int(c)
	}
	return hashSum % 100
}
