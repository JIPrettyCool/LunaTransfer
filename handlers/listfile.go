package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/common"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "strings"
    "fmt"
    "LunaTransfer/utils"
    "LunaTransfer/auth"
)

type FileInfo struct {
    Name      string    `json:"name"`
    Size      int64     `json:"size"`
    IsDir     bool      `json:"isDir"`
    Path      string    `json:"path"`
    Modified  time.Time `json:"modified"`
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    queryPath := r.URL.Query().Get("path")
    if queryPath != "" && strings.Contains(queryPath, "..") {
        utils.LogError("LIST_ERROR", fmt.Errorf("path traversal attempt"), username, queryPath)
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("LIST_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    user, err := auth.GetUserByUsername(username)
    if err != nil {
        utils.LogError("LIST_ERROR", err, username, "Failed to get user details")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    var fileLists [][]map[string]interface{}
    if strings.HasPrefix(queryPath, "groups/") {
        parts := strings.SplitN(queryPath[7:], "/", 2)
        groupID := parts[0]
        
        subPath := ""
        if len(parts) > 1 {
            subPath = parts[1]
        }
        
        hasPermission, err := auth.HasGroupPermission(username, groupID, "read")
        if err != nil {
            utils.LogError("LIST_ERROR", err, username, "Failed to check group permissions")
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }
        
        if !hasPermission {
            utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                fmt.Sprintf("Attempted to list group files without permission: %s", groupID))
            http.Error(w, "Access denied", http.StatusForbidden)
            return
        }
        
        groupDirPath := filepath.Join(appConfig.StorageDirectory, "groups", groupID)
        if subPath != "" {
            groupDirPath = filepath.Join(groupDirPath, subPath)
        }

        groupFiles, err := listFilesInDirectory(groupDirPath, filepath.Join("groups", groupID, subPath))
        if err != nil {
            if os.IsNotExist(err) {
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "files": []interface{}{},
                    "path": queryPath,
                })
                return
            }
            utils.LogError("LIST_ERROR", err, username, fmt.Sprintf("Failed to read group directory: %s", queryPath))
            http.Error(w, "Failed to read directory", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "files": groupFiles,
            "path": queryPath,
        })
        return
    }

    if queryPath == "" {
        userDir := filepath.Join(appConfig.StorageDirectory, username)
        userFiles, err := listFilesInDirectory(userDir, "")
        if err == nil {
            fileLists = append(fileLists, userFiles)
        } else if !os.IsNotExist(err) {
            utils.LogError("LIST_ERROR", err, username, "Failed to read user directory")
        }

        if user.Role == auth.RoleAdmin {
            groups, err := auth.LoadGroups()
            if err == nil {
                for _, group := range groups {
                    groupEntry := map[string]interface{}{
                        "name":     group.Name,
                        "path":     filepath.Join("groups", group.ID),
                        "size":     int64(0),
                        "isDir":    true,
                        "modified": group.CreatedAt,
                        "isGroup":  true,
                    }
                    fileLists = append(fileLists, []map[string]interface{}{groupEntry})
                }
            }
        } else {
            groups, err := auth.LoadGroups()
            if err == nil {
                for _, group := range groups {
                    members, err := auth.GetGroupMembers(group.ID)
                    if err != nil {
                        continue
                    }
                    isMember := false
                    for _, member := range members {
                        if member.Username == username {
                            isMember = true
                            break
                        }
                    }
                    if isMember {
                        groupEntry := map[string]interface{}{
                            "name":     group.Name,
                            "path":     filepath.Join("groups", group.ID),
                            "size":     int64(0),
                            "isDir":    true,
                            "modified": group.CreatedAt,
                            "isGroup":  true,
                        }
                        fileLists = append(fileLists, []map[string]interface{}{groupEntry})
                    }
                }
            }
        }

        userGroups, err := auth.GetUserGroups(username)
        if err == nil && len(userGroups) > 0 {
            sharedSection := []map[string]interface{}{}
            
            allShares, err := auth.LoadSharedFiles()
            if err == nil {
                for _, share := range allShares {
                    for _, group := range userGroups {
                        if share.GroupID == group.ID {
                            sharedInfo := map[string]interface{}{
                                "name":       filepath.Base(share.SourcePath),
                                "path":       share.SourcePath,
                                "sharedBy":   share.SharedBy,
                                "sharedAt":   share.SharedAt,
                                "groupName":  group.Name,
                                "permission": share.Permission,
                                "isShared":   true,
                            }
                            sharedSection = append(sharedSection, sharedInfo)
                        }
                    }
                }
            }
            
            if len(sharedSection) > 0 {
                fileLists = append(fileLists, sharedSection)
            }
        }

        allFiles := []map[string]interface{}{}
        for _, list := range fileLists {
            allFiles = append(allFiles, list...)
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "files": allFiles,
            "path": queryPath,
        })
        return
    }
    userDir := filepath.Join(appConfig.StorageDirectory, username)
    dirPath := filepath.Join(userDir, queryPath)

    if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "files": []interface{}{},
            "path": queryPath,
        })
        return
    }

    userFiles, err := listFilesInDirectory(dirPath, queryPath)
    if err != nil {
        utils.LogError("LIST_ERROR", err, username, fmt.Sprintf("Failed to read directory: %s", queryPath))
        http.Error(w, "Failed to read directory", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "files": userFiles,
        "path": queryPath,
    })
}

func listFilesInDirectory(dirPath, relativePath string) ([]map[string]interface{}, error) {
    files, err := os.ReadDir(dirPath)
    if err != nil {
        return nil, err
    }

    var fileList []map[string]interface{}
    for _, file := range files {
        fileInfo, err := file.Info()
        if err != nil {
            continue
        }
        if strings.HasPrefix(file.Name(), ".") {
            continue
        }
        
        filePath := filepath.Join(relativePath, file.Name())
        fileEntry := map[string]interface{}{
            "name":     file.Name(),
            "path":     filePath,
            "size":     fileInfo.Size(),
            "isDir":    file.IsDir(),
            "modified": fileInfo.ModTime(),
        }
        fileList = append(fileList, fileEntry)
    }
    return fileList, nil
}