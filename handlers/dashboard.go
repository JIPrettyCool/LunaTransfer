package handlers

import (
	"LunaTransfer/config"
	"LunaTransfer/common"
	"LunaTransfer/models"
	"LunaTransfer/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type FileStats struct {
	TotalFiles      int   `json:"totalFiles"`
	TotalSize       int64 `json:"totalSize"`
	AvgFileSize     int64 `json:"avgFileSize"`
	LargestFile     int64 `json:"largestFile"`
	FilesUploaded   int   `json:"filesUploaded"`
	FilesDownloaded int   `json:"filesDownloaded"`
}

type DashboardResponse struct {
	Username       string                   `json:"username"`
	FileStats      FileStats                `json:"fileStats"`
	RecentActivity []models.TransferActivity `json:"recentActivity"`
	StorageUsed    int64                    `json:"storageUsed"`
	StorageLimit   int64                    `json:"storageLimit"`
	StoragePercent float64                  `json:"storagePercent"`
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := r.Context().Value(common.UsernameContextKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userDir := filepath.Join(config.StoragePath, username)
	stats, err := calculateUserStats(userDir)
	if err != nil {
		http.Error(w, "Failed to calculate storage statistics", http.StatusInternalServerError)
		return
	}

	activities, err := utils.GetUserActivity(username, 10)
	if err != nil {
		http.Error(w, "Failed to retrieve activity log", http.StatusInternalServerError)
		return
	}
	
	transferActivities := make([]models.TransferActivity, len(activities))
	for i, activity := range activities {
		transferActivities[i] = models.TransferActivity{
			Timestamp: activity.Timestamp,
			Operation: activity.Operation,
			Filename:  activity.Filename,
			Size:      activity.Size,
		}
	}

	storageLimit := int64(1024 * 1024 * 1024) // 1GB in bytes
	storagePercent := float64(stats.TotalSize) / float64(storageLimit) * 100
	
	response := DashboardResponse{
		Username:       username,
		FileStats:      stats,
		RecentActivity: transferActivities,
		StorageUsed:    stats.TotalSize,
		StorageLimit:   storageLimit,
		StoragePercent: storagePercent,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func calculateUserStats(userDir string) (FileStats, error) {
	stats := FileStats{}
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return stats, err
	}

	err := filepath.Walk(userDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			stats.TotalFiles++
			stats.TotalSize += info.Size()
			if info.Size() > stats.LargestFile {
				stats.LargestFile = info.Size()
			}
		}
		return nil
	})

	if stats.TotalFiles > 0 {
		stats.AvgFileSize = stats.TotalSize / int64(stats.TotalFiles)
	}

	// Get upload/download counts from logs
	// This would be more accurate with a proper database
	// For now, we'll just approximate with dummy values
	stats.FilesUploaded = stats.TotalFiles
	stats.FilesDownloaded = stats.TotalFiles / 2

	return stats, err
}

func GetFileLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement dashboard handler
	fmt.Fprintf(w, "File operation logs will be here.")
}
