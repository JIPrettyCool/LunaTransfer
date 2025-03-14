package middleware

import (
    "LunaTransfer/utils"
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/gorilla/mux"
    "io"
    "net/http"
    "path/filepath"
    "regexp"
    "strings"
    "strconv"
    "time"
)

type ValidationFunc func(body []byte) error
type ParamValidationFunc func(r *http.Request) error

func ValidationMiddleware(validationFunc ValidationFunc) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            var bodyBytes []byte
            var err error
            
            if r.Body != nil && r.Header.Get("Content-Type") == "application/json" {
                bodyBytes, err = io.ReadAll(r.Body)
                if err != nil {
                    utils.LogError("VALIDATION_ERROR", err, r.RemoteAddr)
                    http.Error(w, "Failed to read request body", http.StatusBadRequest)
                    return
                }
                r.Body.Close()
                
                if err := validationFunc(bodyBytes); err != nil {
                    utils.LogError("VALIDATION_ERROR", err, r.RemoteAddr)
                    http.Error(w, err.Error(), http.StatusBadRequest)
                    return
                }
                
                r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
            } else {
                if err := validationFunc(nil); err != nil {
                    utils.LogError("VALIDATION_ERROR", err, r.RemoteAddr)
                    http.Error(w, err.Error(), http.StatusBadRequest)
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

func ParamValidationMiddleware(validationFunc func(*http.Request) error) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if err := validationFunc(r); err != nil {
                utils.LogError("PARAM_VALIDATION_ERROR", err, r.RemoteAddr)
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Login request validation
func ValidateLoginRequest(body []byte) error {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    
    if err := json.Unmarshal(body, &req); err != nil {
        return errors.New("invalid JSON format")
    }
    
    if req.Username == "" {
        return errors.New("username is required")
    }
    
    if req.Password == "" {
        return errors.New("password is required")
    }
    
    return nil
}

// Validate signup/create user request
func ValidateSignupRequest(body []byte) error {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
        Role     string `json:"role"`
    }
    
    if err := json.Unmarshal(body, &req); err != nil {
        return errors.New("invalid JSON format")
    }
    
    if req.Username == "" {
        return errors.New("username is required")
    }
    
    // Username format validation
    if match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,32}$`, req.Username); !match {
        return errors.New("username must be 3-32 characters and contain only letters, numbers, and underscores")
    }
    
    if req.Password == "" {
        return errors.New("password is required")
    }
    
    // Password strength validation
    if len(req.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    hasUpper := strings.ContainsAny(req.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    hasLower := strings.ContainsAny(req.Password, "abcdefghijklmnopqrstuvwxyz")
    hasNumber := strings.ContainsAny(req.Password, "0123456789")
    
    if !hasUpper || !hasLower || !hasNumber {
        return errors.New("password must contain uppercase, lowercase, and numbers")
    }
    
    // Optional email validation
    if req.Email != "" {
        if match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, req.Email); !match {
            return errors.New("invalid email format")
        }
    }
    
    return nil
}

// ValidateFileDeleteRequest ensures the filename parameter is safe
func ValidateFileDeleteRequest(r *http.Request) error {
    vars := mux.Vars(r)
    filename, ok := vars["filename"]
    if !ok || filename == "" {
        return errors.New("filename is required")
    }
    
    // Check for path traversal attempts
    if strings.Contains(filename, "..") || 
       strings.Contains(filename, "/") || 
       strings.Contains(filename, "\\") {
        return errors.New("invalid filename")
    }
    
    return nil
}

// ValidateFileDownloadRequest ensures the filename parameter is safe
func ValidateFileDownloadRequest(r *http.Request) error {
    // Same as delete validation
    vars := mux.Vars(r)
    filename, ok := vars["filename"]
    if !ok || filename == "" {
        return errors.New("filename is required")
    }
    
    if strings.Contains(filename, "..") || 
       strings.Contains(filename, "/") || 
       strings.Contains(filename, "\\") {
        return errors.New("invalid filename")
    }
    
    return nil
}

func ValidateUploadRequest(r *http.Request) error {
    if (!strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")) {
        return errors.New("content type must be multipart/form-data")
    }
    
    // The actual file validation will happen in the handler when we parse the form
    return nil
}

func ValidateRefreshRequest(r *http.Request) error {
    return nil
}

func CreateURLParamValidator(paramName string, validationFn func(string) error) ParamValidationFunc {
    return func(r *http.Request) error {
        vars := mux.Vars(r)
        param, ok := vars[paramName]
        if !ok || param == "" {
            return fmt.Errorf("%s parameter is required", paramName)
        }
        
        return validationFn(param)
    }
}

func ValidateFilename(filename string) error {
    if filename == "" {
        return errors.New("filename is required")
    }
    
    // Check for path traversal attempts
    if strings.Contains(filename, "..") || 
       strings.Contains(filename, "/") || 
       strings.Contains(filename, "\\") {
        return errors.New("invalid filename")
    }
    
    // Check filename extension for common files
    ext := filepath.Ext(filename)
    allowed := []string{".txt", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".jpg", ".jpeg", ".png", ".zip", ".rar"}
    
    isAllowed := false
    for _, a := range allowed {
        if strings.EqualFold(ext, a) {
            isAllowed = true
            break
        }
    }
    
    // Only warn about unusual extensions, not block them
    if !isAllowed {
        utils.LogSystem("UNUSUAL_FILE", "system", "unknown", fmt.Sprintf("Unusual file extension: %s", ext))
    }
    
    return nil
}

func ValidateFilenameParam(r *http.Request) error {
    vars := mux.Vars(r)
    filename, ok := vars["filename"]
    if !ok || filename == "" {
        return errors.New("filename is required")
    }
    filename = strings.Replace(filename, "%2F", "/", -1)
    if strings.Contains(filename, "..") {
        return errors.New("invalid filename format - path traversal detected")
    }
    
    validFilename := regexp.MustCompile(`^[a-zA-Z0-9_\-./() ]+$`)
    if !validFilename.MatchString(filename) {
        return errors.New("filename contains invalid characters")
    }
    
    return nil
}

func ValidateListFilesRequest(r *http.Request) error {
    // Optional validation for sort, filter, or pagination parameters
    return nil
}

func ValidateDirectoryRequest(body []byte) error {
    var req struct {
        Path string `json:"path"`
        Name string `json:"name"`
    }
    
    if err := json.Unmarshal(body, &req); err != nil {
        return errors.New("invalid JSON format")
    }
    
    if req.Name == "" {
        return errors.New("directory name is required")
    }
    
    invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
    for _, char := range invalidChars {
        if strings.Contains(req.Name, char) {
            return errors.New("directory name contains invalid characters")
        }
    }
    
    if strings.Contains(req.Path, "..") || strings.Contains(req.Name, "..") {
        return errors.New("invalid path - path traversal detected")
    }
    
    return nil
}

func ValidateSearchRequest(r *http.Request) error {
    query := r.URL.Query()
    term := query.Get("term")
    if term == "" {
        return errors.New("search term is required")
    }
    path := query.Get("path")
    if path != "" && strings.Contains(path, "..") {
        return errors.New("invalid path - contains path traversal")
    }
    if minSize := query.Get("minSize"); minSize != "" {
        if _, err := strconv.ParseInt(minSize, 10, 64); err != nil {
            return errors.New("minSize must be a valid number")
        }
    }
    if maxSize := query.Get("maxSize"); maxSize != "" {
        if _, err := strconv.ParseInt(maxSize, 10, 64); err != nil {
            return errors.New("maxSize must be a valid number")
        }
    }
    if after := query.Get("after"); after != "" {
        if _, err := time.Parse("2006-01-02", after); err != nil {
            return errors.New("after date must be in YYYY-MM-DD format")
        }
    }
    if before := query.Get("before"); before != "" {
        if _, err := time.Parse("2006-01-02", before); err != nil {
            return errors.New("before date must be in YYYY-MM-DD format")
        }
    }
    return nil
}

