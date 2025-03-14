package handlers

import (
    "LunaTransfer/common"
    "LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

type SearchResult struct {
    Name        string    `json:"name"`
    Path        string    `json:"path"`
    Size        int64     `json:"size"`
    IsDir       bool      `json:"isDir"`
    Modified    time.Time `json:"modified"`
    MatchReason string    `json:"matchReason"`
}

func SearchFilesHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    query := r.URL.Query()
    searchTerm := query.Get("term")
    if searchTerm == "" {
        http.Error(w, "Search term required", http.StatusBadRequest)
        return
    }

    searchPath := filepath.Clean(query.Get("path"))
    if searchPath == "." {
        searchPath = ""
    }
    fileType := query.Get("type")
    var minSize, maxSize int64
    if minSizeStr := query.Get("minSize"); minSizeStr != "" {
        minSize, _ = strconv.ParseInt(minSizeStr, 10, 64)
    }
    if maxSizeStr := query.Get("maxSize"); maxSizeStr != "" {
        maxSize, _ = strconv.ParseInt(maxSizeStr, 10, 64)
    }

    var dateAfter, dateBefore time.Time
    if afterStr := query.Get("after"); afterStr != "" {
        dateAfter, _ = time.Parse("2006-02-01", afterStr)
    }
    if beforeStr := query.Get("before"); beforeStr != "" {
        dateBefore, _ = time.Parse("2006-02-01", beforeStr)
    }

    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("SEARCH_ERROR", err, username)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    userRootDir := filepath.Join(appConfig.StorageDirectory, username)
    searchDir := userRootDir
    if searchPath != "" {
        if strings.Contains(searchPath, "..") {
            utils.LogError("SEARCH_ERROR", fmt.Errorf("path traversal attempt"), username)
            http.Error(w, "Invalid path", http.StatusBadRequest)
            return
        }
        searchDir = filepath.Join(userRootDir, searchPath)
    }

    searchRegex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(searchTerm))
    if err != nil {
        utils.LogError("SEARCH_ERROR", err, username)
        http.Error(w, "Invalid search term", http.StatusBadRequest)
        return
    }

    var results []SearchResult
    err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }

        relPath, _ := filepath.Rel(userRootDir, path)
        if relPath == "." {
            relPath = ""
        }

        if fileType != "" && !info.IsDir() {
            ext := strings.TrimPrefix(filepath.Ext(info.Name()), ".")
            if !strings.EqualFold(ext, fileType) {
                return nil
            }
        }

        if !info.IsDir() {
            size := info.Size()
            if (minSize > 0 && size < minSize) || 
               (maxSize > 0 && size > maxSize) {
                return nil
            }
        }

        modTime := info.ModTime()
        if (!dateAfter.IsZero() && modTime.Before(dateAfter)) ||
           (!dateBefore.IsZero() && modTime.After(dateBefore)) {
            return nil
        }

        var matchReason string
        if searchRegex.MatchString(info.Name()) {
            matchReason = "filename"
        } else if searchRegex.MatchString(relPath) {
            matchReason = "path"
        } else {
            return nil
        }

        results = append(results, SearchResult{
            Name:        info.Name(),
            Path:        relPath,
            Size:        info.Size(),
            IsDir:       info.IsDir(),
            Modified:    info.ModTime(),
            MatchReason: matchReason,
        })
        return nil
    })

    if err != nil {
        utils.LogError("SEARCH_ERROR", err, username)
        http.Error(w, "Search failed", http.StatusInternalServerError)
        return
    }

    sort.Slice(results, func(i, j int) bool {
        if results[i].IsDir != results[j].IsDir {
            return results[i].IsDir
        }
        return results[i].Name < results[j].Name
    })

    utils.LogSystem("FILE_SEARCH", username, r.RemoteAddr, 
        fmt.Sprintf("Searched for '%s' - found %d results", searchTerm, len(results)))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "term":    searchTerm,
        "path":    searchPath,
        "count":   len(results),
        "results": results,
    })
}