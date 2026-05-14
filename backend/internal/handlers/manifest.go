package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vknow360/otaship/backend/internal/database"
	"github.com/vknow360/otaship/backend/internal/utils"
)

type ApiKey struct {
	ID        pgtype.UUID `json:"id"`
	ProjectID pgtype.UUID `json:"project_id"`
}

var (
	recentDownloads      = make(map[string]time.Time)
	recentDownloadsMutex sync.RWMutex
)

// Manifest cache: keyed by "projectID:platform:runtime:channel"
type manifestCacheEntry struct {
	data      []byte // pre-built manifest JSON (nil = no update available)
	updateID  string
	createdAt time.Time
}

const manifestCacheTTL = 10 * time.Minute

var (
	manifestCache      = make(map[string]*manifestCacheEntry)
	manifestCacheMutex sync.RWMutex
)

func manifestCacheKey(projectID, platform, runtime, channel string) string {
	return projectID + ":" + platform + ":" + runtime + ":" + channel
}

func getCachedManifest(key string) (*manifestCacheEntry, bool) {
	manifestCacheMutex.RLock()
	defer manifestCacheMutex.RUnlock()
	entry, ok := manifestCache[key]
	if !ok || time.Since(entry.createdAt) > manifestCacheTTL {
		return nil, false
	}
	return entry, true
}

func setCachedManifest(key string, data []byte, updateID string) {
	manifestCacheMutex.Lock()
	defer manifestCacheMutex.Unlock()
	manifestCache[key] = &manifestCacheEntry{
		data:      data,
		updateID:  updateID,
		createdAt: time.Now(),
	}
}

// InvalidateManifestCache clears all cached manifests for a given project.
// Call this when updates are created, deleted, or rolled back.
func InvalidateManifestCache(projectID string) {
	manifestCacheMutex.Lock()
	defer manifestCacheMutex.Unlock()
	for key := range manifestCache {
		if strings.HasPrefix(key, projectID+":") {
			delete(manifestCache, key)
		}
	}
}

func init() {
	go func() {
		for range time.NewTicker(10 * time.Minute).C {
			now := time.Now()
			recentDownloadsMutex.Lock()
			for key, lastTime := range recentDownloads {
				if now.Sub(lastTime) > 10*time.Minute {
					delete(recentDownloads, key)
				}
			}
			recentDownloadsMutex.Unlock()

			manifestCacheMutex.Lock()
			for key, entry := range manifestCache {
				if now.Sub(entry.createdAt) > manifestCacheTTL {
					delete(manifestCache, key)
				}
			}
			manifestCacheMutex.Unlock()
		}
	}()
}

func CheckForUpdates(queries *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "project_id")

		projectId, err := utils.ParseUUID(id)
		if err != nil {
			jsonError(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		protocolVersionStr := r.Header.Get("expo-protocol-version")
		if protocolVersionStr == "" {
			protocolVersionStr = "0"
		}
		protocolVersion, _ := strconv.Atoi(protocolVersionStr)

		platform := r.Header.Get("expo-platform")
		if platform == "" {
			platform = r.URL.Query().Get("platform")
		}
		if platform != "ios" && platform != "android" {
			jsonError(w, "Unsupported platform. Expected either ios or android", http.StatusBadRequest)
			return
		}

		runtimeVersion := r.Header.Get("expo-runtime-version")
		if runtimeVersion == "" {
			runtimeVersion = r.URL.Query().Get("runtime-version")
		}
		if runtimeVersion == "" {
			jsonError(w, "No runtimeVersion provided", http.StatusBadRequest)
			return
		}

		channel := r.Header.Get("expo-channel-name")
		if channel == "" {
			channel = "production"
		}

		currentUpdateID := r.Header.Get("expo-current-update-id")

		slog.InfoContext(r.Context(), "Manifest request",
			slog.String("project_id", id),
			slog.String("platform", platform),
			slog.String("runtime", runtimeVersion),
			slog.String("channel", channel),
		)

		cacheKey := manifestCacheKey(id, platform, runtimeVersion, channel)
		if cached, ok := getCachedManifest(cacheKey); ok {
			if cached.data == nil {
				handleNoUpdateAvailable(w, r, protocolVersion)
				return
			}
			if currentUpdateID == cached.updateID && protocolVersion == 1 {
				handleNoUpdateAvailable(w, r, protocolVersion)
				return
			}

			slog.DebugContext(r.Context(), "Manifest cache hit",
				slog.String("key", cacheKey),
				slog.String("update_id", cached.updateID),
			)

			deviceHash := utils.BuildDeviceHash(r, platform)
			updateId, _ := utils.ParseUUID(cached.updateID)
			go logDownloadEvent(queries, database.Update{ID: updateId}, projectId, deviceHash, platform, channel)

			contentType := "application/json"
			if protocolVersion == 1 {
				contentType = "application/expo+json"
			}
			sendMultipartResponse(w, r, "manifest", cached.data, protocolVersion, contentType, channel)
			return
		}

		update, err := queries.GetLatestActiveUpdate(r.Context(), database.GetLatestActiveUpdateParams{
			ProjectID:      projectId,
			Platform:       platform,
			RuntimeVersion: runtimeVersion,
			Channel:        channel,
		})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Cache the "no update" result too
				setCachedManifest(cacheKey, nil, "")
				handleNoUpdateAvailable(w, r, protocolVersion)
				return
			}
			jsonError(w, "Failed to fetch update", http.StatusInternalServerError)
			return
		}

		if update.IsRollback {
			embeddedUpdateID := r.Header.Get("expo-embedded-update-id")
			if currentUpdateID == embeddedUpdateID {
				handleNoUpdateAvailable(w, r, protocolVersion)
				return
			}

			directive := map[string]interface{}{
				"type": "rollBackToEmbedded",
				"parameters": map[string]interface{}{
					"commitTime": update.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z"),
				},
			}
			directiveJSON, _ := json.Marshal(directive)
			sendMultipartResponse(w, r, "directive", directiveJSON, protocolVersion, "application/json", channel)
			return
		}

		if update.RolloutPercentage < 100 {
			deviceHash := utils.BuildDeviceHash(r, platform)
			if !shouldReceiveUpdate(int(update.RolloutPercentage), deviceHash) {
				handleNoUpdateAvailable(w, r, protocolVersion)
				return
			}
		}

		if currentUpdateID == update.ID.String() && protocolVersion == 1 {
			handleNoUpdateAvailable(w, r, protocolVersion)
			return
		}

		assets, err := queries.GetAssetsByUpdateID(r.Context(), update.ID)
		if err != nil {
			jsonError(w, "Failed to fetch assets", http.StatusInternalServerError)
			return
		}
		if assets == nil {
			assets = []database.Asset{}
		}

		var launchAsset *database.Asset
		var regularAssets []database.Asset

		for i := range assets {
			if isLaunchAsset(assets[i].FileName) {
				launchAsset = &assets[i]
			} else {
				regularAssets = append(regularAssets, assets[i])
			}
		}

		if launchAsset == nil {
			slog.ErrorContext(r.Context(), "No launch asset found for update", slog.String("update_id", update.ID.String()))
			jsonError(w, "Invalid update: missing launch asset", http.StatusInternalServerError)
			return
		}

		// Build expoClient from stored config
		var expoClient interface{} = map[string]interface{}{}
		if update.ExpoConfig != nil {
			err = json.Unmarshal(update.ExpoConfig, &expoClient)
			if err != nil {
				slog.WarnContext(r.Context(), "Failed to unmarshal expo config", slog.Any("error", err))
			}
		}

		manifest := map[string]interface{}{
			"id":             update.ID.String(),
			"createdAt":      update.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z"),
			"runtimeVersion": update.RuntimeVersion,
			"assets":         buildAssetsArray(regularAssets),
			"metadata":       map[string]interface{}{},
			"extra": map[string]interface{}{
				"expoClient": expoClient,
			},
		}

		launchEntry := map[string]interface{}{
			"hash":        launchAsset.Hash,
			"key":         launchAsset.Key,
			"contentType": launchAsset.MimeType,
			"url":         launchAsset.Url,
		}
		if ext := filepath.Ext(launchAsset.FileName); ext != "" {
			launchEntry["fileExtension"] = ext
		}
		manifest["launchAsset"] = launchEntry

		manifestJSON, err := json.Marshal(manifest)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to marshal manifest", slog.Any("error", err))
			jsonError(w, "Failed to create manifest", http.StatusInternalServerError)
			return
		}

		// Cache the built manifest
		setCachedManifest(cacheKey, manifestJSON, update.ID.String())

		slog.InfoContext(r.Context(), "Sending manifest",
			slog.String("update_id", update.ID.String()),
			slog.String("platform", platform),
			slog.String("runtime", runtimeVersion),
		)

		deviceHash := utils.BuildDeviceHash(r, platform)
		go logDownloadEvent(queries, update, projectId, deviceHash, platform, channel)

		contentType := "application/json"
		if protocolVersion == 1 {
			contentType = "application/expo+json"
		}

		sendMultipartResponse(w, r, "manifest", manifestJSON, protocolVersion, contentType, channel)
	}
}

func logDownloadEvent(
	queries *database.Queries,
	update database.Update,
	projectId pgtype.UUID,
	deviceHash, platform, channel string,
) {

	cacheKey := fmt.Sprintf("%s:%s", deviceHash, update.ID.String())

	recentDownloadsMutex.RLock()
	lastLog, ok := recentDownloads[cacheKey]
	recentDownloadsMutex.RUnlock()

	if ok && time.Since(lastLog) < 5*time.Minute {
		return
	}

	event := database.CreateDownloadEventParams{
		UpdateID:   update.ID,
		ProjectID:  projectId,
		DeviceHash: deviceHash,
		Platform:   platform,
		Channel:    channel,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := queries.CreateDownloadEvent(ctx, event)
	if err != nil {
		slog.Error("Failed to log download event", slog.Any("error", err))
		return
	}

	recentDownloadsMutex.Lock()
	recentDownloads[cacheKey] = time.Now()
	recentDownloadsMutex.Unlock()
}

func shouldReceiveUpdate(percentage int, deviceHash string) bool {
	if percentage >= 100 {
		return true
	}
	if percentage <= 0 {
		return false
	}
	var hashSum int
	for _, c := range deviceHash {
		hashSum += int(c)
	}
	return (hashSum % 100) < percentage
}

func sendMultipartResponse(
	w http.ResponseWriter,
	r *http.Request,
	partName string,
	data []byte,
	protocolVersion int,
	partContentType string,
	channel string,
) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", partContentType)
	partHeader.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, partName),
	)

	if r.Header.Get("expo-expect-signature") != "" {
		pvtKey := os.Getenv("EXPO_PRIVATE_KEY")
		if pvtKey != "" {
			sig, err := utils.SignManifest(data, pvtKey)
			if err != nil {
				slog.Error("Code signing error", slog.Any("error", err))
			} else if sig != "" {
				expoSignature := fmt.Sprintf(`sig="%s", keyid="main"`, sig)
				w.Header().Set("expo-signature", expoSignature)
				partHeader.Set("expo-signature", expoSignature)
			}
		}
	}

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		slog.Error("Error creating multipart part", slog.Any("error", err))
		jsonError(w, "Failed to create response", http.StatusInternalServerError)
		return
	}

	_, err = part.Write(data)
	if err != nil {
		jsonError(w, "Failed to write response", http.StatusInternalServerError)
		return
	}

	writer.Close()

	w.Header().Set("expo-protocol-version", strconv.Itoa(protocolVersion))
	w.Header().Set("expo-sfv-version", "0")
	w.Header().Set("cache-control", "private, max-age=0")
	w.Header().Set("Content-Type", "multipart/mixed; boundary="+writer.Boundary())
	if channel != "" {
		w.Header().Set("expo-manifest-filters", fmt.Sprintf(`channel-name="%s"`, channel))
		w.Header().Set("expo-server-defined-headers", fmt.Sprintf(`expo-channel-name="%s"`, channel))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func handleNoUpdateAvailable(
	w http.ResponseWriter,
	r *http.Request,
	protocolVersion int,
) {
	if protocolVersion == 0 {
		jsonError(w, "No update available", http.StatusNotFound)
		return
	}

	directive := map[string]interface{}{
		"type": "noUpdateAvailable",
	}

	directiveJSON, err := json.Marshal(directive)
	if err != nil {
		jsonError(w, "Failed to marshal directive", http.StatusInternalServerError)
		return
	}

	sendMultipartResponse(w, r, "directive", directiveJSON, protocolVersion, "application/json", "")
}

func isLaunchAsset(fileName string) bool {
	return strings.HasPrefix(fileName, "_expo/static/js/")
}

func buildAssetsArray(assets []database.Asset) []map[string]interface{} {
	result := make([]map[string]interface{}, len(assets))
	for i, asset := range assets {
		entry := map[string]interface{}{
			"hash":        asset.Hash,
			"key":         asset.Key,
			"contentType": asset.MimeType,
			"url":         asset.Url,
		}
		if ext := filepath.Ext(asset.FileName); ext != "" {
			entry["fileExtension"] = ext
		}
		result[i] = entry
	}
	return result
}
