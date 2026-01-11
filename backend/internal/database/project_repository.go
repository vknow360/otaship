package database

import (
	"context"
	"errors"
	"time"

	"github.com/vknow360/otaship/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNoDatabase is returned when database is not connected.
var ErrNoDatabase = errors.New("database not connected")

// ProjectRepository handles project data.
type ProjectRepository struct {
	collection *mongo.Collection
}

// NewProjectRepository creates a new project repository.
func NewProjectRepository(db *MongoDB) *ProjectRepository {
	if db == nil {
		return &ProjectRepository{}
	}
	return &ProjectRepository{
		collection: db.Database.Collection("projects"),
	}
}

// Create inserts a new project.
func (r *ProjectRepository) Create(project *models.Project) error {
	if r.collection == nil {
		return ErrNoDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	project.CreatedAt = time.Now()
	project.UpdateCount = 0

	result, err := r.collection.InsertOne(ctx, project)
	if err != nil {
		return err
	}

	project.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindAll returns all projects.
func (r *ProjectRepository) FindAll() ([]models.Project, error) {
	if r.collection == nil {
		return []models.Project{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []models.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// FindBySlug returns a project by its slug.
func (r *ProjectRepository) FindBySlug(slug string) (*models.Project, error) {
	if r.collection == nil {
		return nil, ErrNoDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var project models.Project
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// Delete removes a project by slug.
func (r *ProjectRepository) Delete(slug string) error {
	if r.collection == nil {
		return ErrNoDatabase
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"slug": slug})
	return err
}

// IncrementUpdateCount increments the update count for a project.
func (r *ProjectRepository) IncrementUpdateCount(slug string, delta int) error {
	if r.collection == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"slug": slug},
		bson.M{"$inc": bson.M{"updateCount": delta}},
	)
	return err
}

// EnsureProjectExists creates a project if it doesn't exist (upsert).
func (r *ProjectRepository) EnsureProjectExists(slug, name string) error {
	if r.collection == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"slug": slug},
		bson.M{
			"$setOnInsert": bson.M{
				"slug":        slug,
				"name":        name,
				"createdAt":   time.Now(),
				"updateCount": 0,
			},
		},
		opts,
	)
	return err
}
