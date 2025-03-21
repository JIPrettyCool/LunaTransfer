package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type App struct {
	ctx        context.Context
	apiBaseURL string
}

type FileItem struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
	Size        int64  `json:"size"`
	Modified    string `json:"modified"`
}

func NewApp() *App {
	return &App{
		apiBaseURL: "http://localhost:8080",
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("Application started with API URL:", a.apiBaseURL)
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoginUser(username, password string) (map[string]interface{}, error) {
	fmt.Println("Attempting login for user:", username)
	
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

	fmt.Println("Login response status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed with status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	fmt.Println("Login successful for user:", username)

	var role string
	if roleVal, ok := result["role"].(string); ok {
		role = roleVal
	} else {
		role = "user"
	}

	return map[string]interface{}{
		"token":    result["token"],
		"username": username,
		"role":     role,
	}, nil
}

func (a *App) listUserFiles(token, path string) ([]FileItem, error) {
	fmt.Printf("Listing files at path: %q\n", path)

	encodedPath := ""
	if path != "" {
		encodedPath = "?path=" + url.QueryEscape(path)
	}
	
	apiURL := a.apiBaseURL + "/api/files" + encodedPath
	fmt.Println("API URL:", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
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

	fmt.Println("List files response status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("files request failed with status: %d, response: %s", 
			resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Files []map[string]interface{} `json:"files"`
		Path  string                   `json:"path"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		var fileArray []map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &fileArray); err != nil {
			return nil, fmt.Errorf("failed to parse files response: %w", err)
		}
		result.Files = fileArray
	}

	fileItems := make([]FileItem, 0, len(result.Files))
	for _, file := range result.Files {
		name, ok := file["name"].(string)
		if !ok {
			continue
		}

		filePath, ok := file["path"].(string)
		if !ok {
			filePath = path + "/" + name
			if path == "" {
				filePath = name
			}
		}

		isDir := false
		if isDirVal, ok := file["isDirectory"].(bool); ok {
			isDir = isDirVal
		} else if isDirVal, ok := file["isDir"].(bool); ok {
			isDir = isDirVal
		}

		var size int64
		if sizeFloat, ok := file["size"].(float64); ok {
			size = int64(sizeFloat)
		}

		modified := time.Now().Format(time.RFC3339)
		if modifiedStr, ok := file["modified"].(string); ok {
			modified = modifiedStr
		}

		fileItem := FileItem{
			Name:        name,
			Path:        filePath,
			IsDirectory: isDir,
			Size:        size,
			Modified:    modified,
		}
		
		fileItems = append(fileItems, fileItem)
	}

	return fileItems, nil
}

func (a *App) ListUserFiles(token, path string) ([]FileItem, error) {
	return a.listUserFiles(token, path)
}

func (a *App) UploadFile(token, path string, fileData []byte, filename string) error {
	fmt.Printf("Uploading file: %s to path: %s\n", filename, path)
	
	url := a.apiBaseURL + "/api/upload"
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	err := writer.WriteField("path", path)
	if err != nil {
		return fmt.Errorf("failed to add path field: %w", err)
	}
	
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	
	_, err = part.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}
	
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer " + token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("file upload request failed: %w", err)
	}
	defer resp.Body.Close()
	
	fmt.Println("Upload response status:", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("file upload failed with status: %d, response: %s", 
			resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

func (a *App) CreateDirectory(token, path, folderName string) error {
	fmt.Printf("Creating directory: %s in path: %s\n", folderName, path)
	
	url := a.apiBaseURL + "/api/directory"
	
	reqData := map[string]string{
		"path": path,
		"name": folderName,
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to encode request data: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("directory creation request failed: %w", err)
	}
	defer resp.Body.Close()
	
	fmt.Println("Directory creation response status:", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("directory creation failed with status: %d, response: %s", 
			resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

func (a *App) DeleteFile(token, path string) error {
    fmt.Printf("Deleting item at path: %s\n", path)
    sanitizedPath := strings.TrimPrefix(path, "/")
    deleteURL := a.apiBaseURL + "/api/delete"
    reqData := map[string]string{
        "path": sanitizedPath,
    }
    jsonData, err := json.Marshal(reqData)
    if err != nil {
        return fmt.Errorf("failed to encode request data: %w", err)
    }
    
    req, err := http.NewRequest("POST", deleteURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + token)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("delete request failed: %w", err)
    }
    defer resp.Body.Close()
    
    fmt.Printf("Delete response status: %d\n", resp.StatusCode)
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("delete failed with status: %d, response: %s", 
            resp.StatusCode, string(bodyBytes))
    }
    
    return nil
}
