package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// App struct
type App struct {
	ctx        context.Context
	apiBaseURL string
}

// FileItem represents a file or directory
type FileItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
	Size        int64  `json:"size"`
	Modified    string `json:"modified"` // Changed from time.Time to string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		apiBaseURL: "http://localhost:8080", // Your API URL
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// LoginUser handles user authentication via API
func (a *App) LoginUser(username, password string) (map[string]interface{}, error) {
	// Make a login request to your API
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode login data: %w", err)
	}

	resp, err := http.Post(
		a.apiBaseURL+"/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	return map[string]interface{}{
		"token":    result["token"],
		"username": username,
	}, nil
}

// listUserFiles gets files via the API - internal version
func (a *App) listUserFiles(token, path string) ([]FileItem, error) {
	url := a.apiBaseURL + "/api/files"
	if path != "" {
		url += "?path=" + path
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("files request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("files request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Files []map[string]interface{} `json:"files"`
		Path  string                   `json:"path"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse files response: %w", err)
	}

	fileItems := make([]FileItem, 0, len(result.Files))
	for _, file := range result.Files {
		// Parse the modified time from your API response
		var modified time.Time
		if modifiedStr, ok := file["modified"].(string); ok {
			modifiedTime, err := time.Parse(time.RFC3339, modifiedStr)
			if err == nil {
				modified = modifiedTime
			}
		}

		// Get size as int64 carefully
		var size int64
		if sizeFloat, ok := file["size"].(float64); ok {
			size = int64(sizeFloat)
		}

		isDir := false
		if isDirVal, ok := file["isDir"].(bool); ok {
			isDir = isDirVal
		}

		fileItem := FileItem{
			Name:        file["name"].(string),
			Path:        file["path"].(string),
			IsDirectory: isDir,
			Size:        size,
			Modified:    modified.Format(time.RFC3339), // Convert time to string
		}

		fileItems = append(fileItems, fileItem)
	}

	return fileItems, nil
}

// ListUserFiles gets files via the API - TypeScript-safe version
func (a *App) ListUserFiles(token, path string) ([]map[string]interface{}, error) {
	items, err := a.listUserFiles(token, path)
	if err != nil {
		return nil, err
	}
	
	// Convert to TypeScript-friendly format
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = map[string]interface{}{
			"name":        item.Name,
			"path":        item.Path,
			"isDirectory": item.IsDirectory,
			"size":        item.Size,
			"modified":    item.Modified,
		}
	}
	
	return result, nil
}

// Add more API methods as needed
func (a *App) UploadFile(token, path string, fileData []byte, filename string) error {
	// Implement file upload
	return nil
}

func (a *App) CreateDirectory(token, path, folderName string) error {
	// Implement directory creation
	return nil
}

func (a *App) DeleteFile(token, path string) error {
	// Implement file/folder deletion
	return nil
}
