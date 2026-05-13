package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/storage"
	"github.com/vknow360/otaship/backend/internal/utils"
)

type ExpoMetadata struct {
	Version      int                             `json:"version"`
	Bundler      string                          `json:"bundler"`
	FileMetadata map[string]PlatformFileMetadata `json:"fileMetadata"`
}

type PlatformFileMetadata struct {
	Assets []AssetMetadata `json:"assets"`
	Bundle string          `json:"bundle"`
}

type AssetMetadata struct {
	Path        string `json:"path"`
	Ext         string `json:"ext"`
	ContentType string `json:"contentType"`
}

type UploadedAsset struct {
	FileName    string
	StorageKey  string
	StorageURL  string
	FileHash    string
	Hash        string
	ContentType string
}

func UploadAsset(pool *pgxpool.Pool, queries *database.Queries, providers map[string]storage.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateIdStr := chi.URLParam(r, "update_id")
		projectIdStr := chi.URLParam(r, "project_id")
		if updateIdStr == "" {
			jsonError(w, "Update ID is required", http.StatusBadRequest)
			return
		}
		if projectIdStr == "" {
			jsonError(w, "Project ID is required", http.StatusBadRequest)
			return
		}
		updateId, err := utils.ParseUUID(updateIdStr)
		if err != nil {
			slog.ErrorContext(r.Context(), "Invalid update ID", slog.String("update_id", updateIdStr), slog.Any("error", err))
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}
		projectId, err := utils.ParseUUID(projectIdStr)
		if err != nil {
			slog.ErrorContext(r.Context(), "Invalid project ID", slog.String("project_id", projectIdStr), slog.Any("error", err))
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		if projectId != utils.GetProjectId(r.Context()) {
			slog.WarnContext(r.Context(), "Project ID does not match the authenticated user", slog.String("project_id", projectIdStr))
			jsonError(w, "Project ID does not match the authenticated user", http.StatusForbidden)
			return
		}

		project, err := queries.GetProjectByID(r.Context(), projectId)
		if err != nil {
			slog.ErrorContext(r.Context(), "Project not found", slog.String("project_id", projectIdStr))
			jsonError(w, "Project not found", http.StatusNotFound)
			return
		}

		tx, err := pool.Begin(r.Context())
		if err != nil {
			jsonError(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}

		defer tx.Rollback(r.Context())

		qtx := queries.WithTx(tx)

		update, err := qtx.GetUpdateByID(r.Context(), updateId)
		if err != nil {
			slog.ErrorContext(r.Context(), "Update not found",
				slog.String("update_id", updateIdStr),
				slog.Any("error", err),
			)
			jsonError(w, "Update not found", http.StatusNotFound)
			return
		}

		if update.ProjectID != project.ID {
			slog.WarnContext(r.Context(), "Update does not belong to project",
				slog.String("update_id", update.ID.String()),
				slog.String("project_id", project.ID.String()),
			)
			jsonError(w, "Update does not belong to the specified project", http.StatusBadRequest)
			return
		}

		const maxUploadSize = 50 << 20
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			jsonError(w, "File too large", http.StatusBadRequest)
			return
		}

		platform := r.FormValue("platform")
		if platform == "" {
			jsonError(w, "Platform is required", http.StatusBadRequest)
			return
		}
		if platform != "ios" && platform != "android" {
			jsonError(w, "Invalid platform", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("bundle")
		if err != nil {
			jsonError(w, "Bundle file is required", http.StatusBadRequest)
			return
		}

		defer file.Close()

		if header.Header.Get("Content-Type") != "application/zip" &&
			header.Header.Get("Content-Type") != "application/x-zip-compressed" &&
			header.Header.Get("Content-Type") != "application/octet-stream" {
			slog.WarnContext(r.Context(), "Invalid content type for bundle",
				slog.String("content_type", header.Header.Get("Content-Type")),
			)
			jsonError(w, "Bundle must be a zip file", http.StatusBadRequest)
			return
		}

		os.MkdirAll("./uploads", 0755)
		tempFile, err := os.CreateTemp("./uploads/", "otaship-upload-*.zip")
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to create temp file", slog.Any("error", err))
			jsonError(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		bytesWritten, err := io.Copy(tempFile, file)
		if err != nil {
			jsonError(w, "Failed to write file", http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.NewReader(tempFile, bytesWritten)
		if err != nil {
			jsonError(w, "Failed to read zip file", http.StatusBadRequest)
			return
		}

		metadata, expoConfig, err := parseZipMetadata(r.Context(), zipReader)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to parse metadata", slog.Any("error", err))
			jsonError(w, "Failed to parse metadata", http.StatusBadRequest)
			return
		}

		// Store expoConfig on the update
		if expoConfig != nil {
			err = qtx.UpdateExpoConfig(r.Context(), database.UpdateExpoConfigParams{
				ExpoConfig: expoConfig,
				ID:         update.ID,
			})
			if err != nil {
				slog.WarnContext(r.Context(), "Failed to store expo config", slog.Any("error", err))
			}
		}

		platformMetadata, exists := metadata.FileMetadata[platform]
		if !exists {
			slog.ErrorContext(r.Context(), "Platform metadata not found", slog.String("platform", platform))
			jsonError(w, "Platform metadata not found in bundle", http.StatusBadRequest)
			return
		}

		err = qtx.DeleteAssetByUpdateIDandPlatform(r.Context(), database.DeleteAssetByUpdateIDandPlatformParams{
			UpdateID: update.ID,
			Platform: platform,
		})
		if err != nil {
			slog.WarnContext(r.Context(), "Failed to delete old assets", slog.Any("error", err))
		}

		filesToUpload := make(map[string]bool)
		filesToUpload[normalizeAssetPath(platformMetadata.Bundle)] = true
		for _, asset := range platformMetadata.Assets {
			filesToUpload[normalizeAssetPath(asset.Path)] = true
		}

		providerName, err := qtx.GetSetting(r.Context(), "storage_provider")
		storage, ok := providers[providerName.Value]
		if !ok {
			slog.WarnContext(r.Context(), "Storage provider not found", slog.String("provider", providerName.Value))
			for _, provider := range providers {
				storage = provider
				break
			}
		}

		var uploadedAssets []UploadedAsset

		deleteAsset := func(ctx context.Context, asset UploadedAsset) {
			if err := storage.Delete(ctx, asset.StorageKey, asset.ContentType); err != nil {
				slog.ErrorContext(ctx, "Failed to delete asset", slog.Any("error", err))
			}
		}

		cleanupAssets := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for _, asset := range uploadedAssets {
				go deleteAsset(ctx, asset)
			}
		}

		for _, zipFile := range zipReader.File {
			normalizedZipName := normalizeAssetPath(zipFile.Name)
			if !filesToUpload[normalizedZipName] {
				slog.DebugContext(r.Context(), "Skipping asset",
					slog.String("original", zipFile.Name),
					slog.String("normalized", normalizedZipName),
				)
				continue
			}
			asset, err := uploadZipAssets(r.Context(), storage, platformMetadata, zipFile, project.Slug, update.ID.String(), platform, normalizedZipName)
			if err != nil {
				slog.ErrorContext(r.Context(), "Failed to upload asset",
					slog.String("asset", normalizedZipName),
					slog.Any("error", err),
				)
				// should cleanup already uploaded assets
				cleanupAssets()

				// also delete update from db
				err = qtx.DeleteUpdate(r.Context(), update.ID)
				if err != nil {
					slog.ErrorContext(r.Context(), "Failed to delete update", slog.Any("error", err))
				}
				jsonError(w, "Failed to upload asset", http.StatusInternalServerError)
				return
			}
			uploadedAssets = append(uploadedAssets, asset)
			slog.InfoContext(r.Context(), "Uploaded asset",
				slog.String("asset", normalizedZipName),
				slog.String("key", asset.StorageKey),
			)
		}

		err = saveAssetRecords(r.Context(), qtx, uploadedAssets, platform, update, storage.Name())
		if err != nil {
			cleanupAssets()
			slog.ErrorContext(r.Context(), "Failed to save asset records", slog.Any("error", err))
			jsonError(w, "Failed to save asset records", http.StatusInternalServerError)
			return
		}

		err = qtx.DeactivateUpdates(r.Context(), database.DeactivateUpdatesParams{
			ProjectID:      update.ProjectID,
			Channel:        update.Channel,
			Platform:       update.Platform,
			RuntimeVersion: update.RuntimeVersion,
		})
		if err != nil {
			cleanupAssets()
			slog.WarnContext(r.Context(), "Failed to deactivate updates", slog.Any("error", err))
			jsonError(w, "Failed to deactivate updates", http.StatusInternalServerError)
			return
		}

		err = qtx.ActivateUpdate(r.Context(), update.ID)
		if err != nil {
			cleanupAssets()
			slog.ErrorContext(r.Context(), "Failed to activate update", slog.Any("error", err))
			jsonError(w, "Failed to activate update", http.StatusInternalServerError)
			return
		}

		err = tx.Commit(r.Context())
		if err != nil {
			cleanupAssets()
			slog.ErrorContext(r.Context(), "Failed to commit transaction", slog.Any("error", err))
			jsonError(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		go InvalidateManifestCache(projectIdStr)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":        true,
			"message":        "Assets uploaded successfully",
			"project":        project.Name,
			"update":         update.ID.String(),
			"platform":       platform,
			"uploadedAssets": len(uploadedAssets),
		})
	}
}

func ListUpdateAssets(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "update_id")
		updateId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid update ID", http.StatusBadRequest)
			return
		}

		assets, err := queries.GetAssetsByUpdateID(r.Context(), updateId)
		if err != nil {
			jsonError(w, "Failed to fetch assets", http.StatusInternalServerError)
			return
		}

		if assets == nil {
			assets = []database.Asset{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assets)
	}
}

func saveAssetRecords(ctx context.Context, queries *database.Queries, uploadedAssets []UploadedAsset, platform string, update database.Update, provider string) error {
	for _, asset := range uploadedAssets {
		err := queries.CreateAsset(ctx, database.CreateAssetParams{
			UpdateID:        update.ID,
			Platform:        platform,
			FileName:        asset.FileName,
			MimeType:        asset.ContentType,
			Key:             asset.StorageKey,
			Url:             asset.StorageURL,
			FileHash:        asset.FileHash,
			Hash:            asset.Hash,
			StorageProvider: provider,
		})

		if err != nil {
			slog.ErrorContext(ctx, "Failed to save asset metadata",
				slog.String("file", asset.FileName),
				slog.Any("error", err),
			)
			return errors.New("Failed to save asset metadata")
		}
	}
	return nil
}

func uploadZipAssets(ctx context.Context, storage storage.Provider, platformMetadata PlatformFileMetadata, zipFile *zip.File, projectSlug, updateID, platform, normalizedZipName string) (UploadedAsset, error) {
	slog.DebugContext(ctx, "Processing zip file", slog.String("name", zipFile.Name))

	fileReader, err := zipFile.Open()
	if err != nil {
		return UploadedAsset{}, errors.New("failed to read file from zip " + zipFile.Name)
	}
	defer fileReader.Close()

	headBuffer := make([]byte, 512)

	n, _ := io.ReadFull(fileReader, headBuffer)
	if n == 0 {
		return UploadedAsset{}, errors.New("failed to read file data from zip " + zipFile.Name)
	}
	headerBytes := headBuffer[:n]
	contentType := http.DetectContentType(headerBytes)

	if normalizedZipName == normalizeAssetPath(platformMetadata.Bundle) {
		contentType = inferBundleContentType(normalizedZipName)
	}

	for _, asset := range platformMetadata.Assets {
		if normalizeAssetPath(asset.Path) == normalizedZipName && asset.ContentType != "" {
			contentType = asset.ContentType
			break
		}
	}

	fullStream := io.MultiReader(bytes.NewReader(headerBytes), fileReader)
	hasher := sha256.New()

	teeStream := io.TeeReader(fullStream, hasher)

	storageKey := buildStorageKey(
		projectSlug,
		updateID,
		platform,
		normalizedZipName,
	)

	storageURL, err := storage.Upload(
		ctx,
		storageKey,
		teeStream,
		contentType,
	)

	if err != nil {
		slog.ErrorContext(ctx, "Failed to upload file to storage",
			slog.String("file", zipFile.Name),
			slog.String("storageKey", storageKey),
			slog.Any("error", err),
		)
		return UploadedAsset{}, errors.New("Failed to upload file to storage: " + zipFile.Name)
	}

	hashBytes := hasher.Sum(nil)
	fileHash := hex.EncodeToString(hashBytes)
	hash := base64.RawURLEncoding.EncodeToString(hashBytes)

	return UploadedAsset{
		FileName:    normalizedZipName,
		StorageKey:  storageKey,
		StorageURL:  storageURL,
		FileHash:    fileHash,
		Hash:        hash,
		ContentType: contentType,
	}, nil
}

func parseZipMetadata(ctx context.Context, zipReader *zip.Reader) (ExpoMetadata, json.RawMessage, error) {
	slog.DebugContext(ctx, "Parsing zip metadata")

	var metadata ExpoMetadata
	found := false

	for _, zipFile := range zipReader.File {
		if zipFile.Name == "metadata.json" {
			found = true

			metadataReader, err := zipFile.Open()
			if err != nil {
				slog.ErrorContext(ctx, "Failed to open metadata.json", slog.Any("error", err))
				return ExpoMetadata{}, nil, errors.New("failed to read metadata.json")
			}
			defer metadataReader.Close()

			if err := json.NewDecoder(metadataReader).Decode(&metadata); err != nil {
				slog.ErrorContext(ctx, "Failed to decode metadata.json", slog.Any("error", err))
				return ExpoMetadata{}, nil, errors.New("Invalid metadata.json format")
			}

			break
		}
	}

	if !found {
		slog.WarnContext(ctx, "metadata.json not found in ZIP")
		return ExpoMetadata{}, nil, errors.New("metadata.json not found in bundle")
	}

	// Parse expoConfig.json if present
	var expoConfig json.RawMessage
	for _, zipFile := range zipReader.File {
		if zipFile.Name == "expoConfig.json" {
			configReader, err := zipFile.Open()
			if err != nil {
				slog.WarnContext(ctx, "Failed to open expoConfig.json", slog.Any("error", err))
				break
			}
			configData, err := io.ReadAll(configReader)
			configReader.Close()
			if err != nil {
				slog.WarnContext(ctx, "Failed to read expoConfig.json", slog.Any("error", err))
				break
			}
			expoConfig = json.RawMessage(configData)
			break
		}
	}
	return metadata, expoConfig, nil
}

func inferBundleContentType(fileName string) string {
	lowerName := strings.ToLower(fileName)
	if strings.HasSuffix(lowerName, ".hbc") {
		return "application/octet-stream"
	}
	if strings.HasSuffix(lowerName, ".js") {
		return "application/javascript"
	}
	return "application/octet-stream"
}

func buildStorageKey(projectSlug, updateID, platform, fileName string) string {
	return projectSlug + "/" + updateID + "/" + platform + "/" + fileName
}

func normalizeAssetPath(path string) string {
	return filepath.ToSlash(strings.TrimPrefix(path, "./"))
}
