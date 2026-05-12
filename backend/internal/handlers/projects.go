package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/utils"
)

type CreateProjectRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

func toProjectResponse(p database.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Time.UnixMilli(),
	}
}

func GetProjects(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projects, err := queries.ListProjects(r.Context())
		if err != nil {
			jsonError(w, "Failed to fetch projects", http.StatusInternalServerError)
			return
		}
		if projects == nil {
			projects = []database.Project{}
		}

		projectsList := make([]ProjectResponse, len(projects))
		for i, project := range projects {
			projectsList[i] = toProjectResponse(project)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projectsList)
	}
}

func GetProjectByID(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "project_id")
		projectId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		project, err := queries.GetProjectByID(r.Context(), projectId)
		if err != nil {
			jsonError(w, "Project not found", http.StatusNotFound)
			return
		}

		res := ProjectResponse{
			ID:          project.ID.String(),
			Slug:        project.Slug,
			Name:        project.Name,
			Description: project.Description,
			CreatedAt:   project.CreatedAt.Time.UnixMilli(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func CreateProject(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateProjectRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Slug == "" || req.Name == "" {
			jsonError(w, "Slug and name are required", http.StatusBadRequest)
			return
		}

		project, err := queries.CreateProject(r.Context(), database.CreateProjectParams{
			Slug:        req.Slug,
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			jsonError(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ProjectResponse{
			ID:          project.ID.String(),
			Slug:        project.Slug,
			Name:        project.Name,
			Description: project.Description,
			CreatedAt:   project.CreatedAt.Time.UnixMilli(),
		})
	}
}

func DeleteProject(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "project_id")
		projectId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		_, err = queries.DeleteProject(r.Context(), projectId)
		if err != nil {
			jsonError(w, "Failed to delete project", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateProject(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "project_id")
		projectId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		var req UpdateProjectRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		project, err := queries.UpdateProject(r.Context(), database.UpdateProjectParams{
			ID:          projectId,
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			jsonError(w, "Failed to update project", http.StatusInternalServerError)
			return
		}

		res := ProjectResponse{
			ID:          project.ID.String(),
			Slug:        project.Slug,
			Name:        project.Name,
			Description: project.Description,
			CreatedAt:   project.CreatedAt.Time.UnixMilli(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}

func ValidateAPIKey(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" || len(apiKey) < 16 {
			jsonError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		keySuffix := apiKey[len(apiKey)-16:]
		key, err := queries.GetAPIKeyBySuffix(r.Context(), keySuffix)
		if err != nil {
			jsonError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		hash := sha256.Sum256([]byte(apiKey))
		computed := hex.EncodeToString(hash[:])
		if computed != key.KeyHash {
			jsonError(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		project, err := queries.GetProjectByID(r.Context(), key.ProjectID)
		if err != nil {
			jsonError(w, "Project not found", http.StatusNotFound)
			return
		}

		go func() {
			queries.UpdateAPIKeyLastUsed(context.Background(), key.ID)
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"project_id": key.ProjectID.String(),
			"slug":       project.Slug,
			"name":       project.Name,
		})
	}
}

func GetMe(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := utils.GetProjectId(r.Context())

		var id pgtype.UUID = projectId
		project, err := queries.GetProjectByID(r.Context(), id)
		if err != nil {
			jsonError(w, "Project not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":   project.ID.String(),
			"slug": project.Slug,
			"name": project.Name,
		})
	}
}
