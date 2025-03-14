package auth

import (
    "LunaTransfer/config"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "sync"
    "time"
    "golang.org/x/crypto/bcrypt"
)

var (
    usersFile     = "users.json"
    usersMutex    = &sync.RWMutex{}
    users         = make(map[string]User)
    apiKeyToUser  = make(map[string]string)
    ErrUserExists = errors.New("user already exists")
    ErrWeakPassword = errors.New("password must be at least 8 characters and contain numbers and letters")
    ErrInvalidCredentials = errors.New("invalid username or password")
    ErrUsersFileNotFound  = errors.New("users file not found")
)

type User struct {
    Username     string    `json:"username"`
    Password     string    `json:"-"`
    PasswordHash string    `json:"password_hash"`// Just for storage
    Email        string    `json:"email"`
    Role         string    `json:"role"`
    APIKey       string    `json:"api_key"`
    CreatedAt    time.Time `json:"created_at"`
    LastLogin    time.Time `json:"last_login"`
}

func ValidatePassword(password string) bool {
    if len(password) < 8 {
        return false
    }
    hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    return hasLetter && hasNumber
}

func GenerateAPIKey() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

func CreateUser(username, password, email, role string) (User, string, error) {
    if UserExists(username) {
        return User{}, "", fmt.Errorf("user already exists")
    }
    
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return User{}, "", err
    }
    
    apiKey, err := GenerateAPIKey()
    if err != nil {
        return User{}, "", err
    }
    
    user := User{
        Username:     username,
        PasswordHash: string(hashedPassword),
        Email:        email,
        Role:         role,
        APIKey:       apiKey,
        CreatedAt:    time.Now(),
        LastLogin:    time.Time{},
    }
    
    usersMutex.Lock()
    users[username] = user
    apiKeyToUser[apiKey] = username
    usersMutex.Unlock()
    
    if err := saveUsers(); err != nil {
        return User{}, "", err
    }
    
    return user, apiKey, nil
}

func Authenticate(username, password string) (string, error) {
    usersMutex.RLock()
    user, exists := users[username]
    usersMutex.RUnlock()

    if (!exists) {
        return "", ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return "", ErrInvalidCredentials
    }

    usersMutex.Lock()
    user.LastLogin = time.Now()
    users[username] = user
    usersMutex.Unlock()

    if err := saveUsers(); err != nil {
        return "", fmt.Errorf("error updating last login: %w", err)
    }

    return user.APIKey, nil
}

func AuthenticateUser(username, password string) (User, string, error) {
    usersMutex.RLock()
    user, exists := users[username]
    usersMutex.RUnlock()

    if (!exists) {
        return User{}, "", ErrInvalidCredentials
    }

    err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    if err != nil {
        return User{}, "", ErrInvalidCredentials
    }

    usersMutex.Lock()
    user.LastLogin = time.Now()
    users[username] = user
    usersMutex.Unlock()

    if err := saveUsers(); err != nil {
        fmt.Printf("Failed to update last login time: %v\n", err)
    }

    return user, user.APIKey, nil
}

func GetUserByAPIKey(apiKey string) (string, bool) {
    usersMutex.RLock()
    defer usersMutex.RUnlock()
    
    username, exists := apiKeyToUser[apiKey]
    return username, exists
}

func LoadUsers() (map[string]User, error) {
    usersMutex.Lock()
    defer usersMutex.Unlock()

    if _, err := os.Stat(usersFile); os.IsNotExist(err) {
        return users, nil
    }

    data, err := os.ReadFile(usersFile)
    if err != nil {
        return nil, fmt.Errorf("error reading users file: %w", err)
    }

    var loadedUsers []User
    if err := json.Unmarshal(data, &loadedUsers); err != nil {
        return nil, fmt.Errorf("error parsing users file: %w", err)
    }

    users = make(map[string]User)
    apiKeyToUser = make(map[string]string)

    for _, u := range loadedUsers {
        users[u.Username] = u
        apiKeyToUser[u.APIKey] = u.Username
    }

    return users, nil
}

func saveUsers() error {
    usersList := []User{}
    for _, user := range users {
        usersList = append(usersList, user)
    }

    data, err := json.MarshalIndent(usersList, "", "  ")
    if err != nil {
        return fmt.Errorf("error encoding users: %w", err)
    }

    tempFile := usersFile + ".tmp"
    if err := os.WriteFile(tempFile, data, 0600); err != nil {
        return fmt.Errorf("error writing users file: %w", err)
    }

    return os.Rename(tempFile, usersFile)
}

func RotateAPIKey(username string) (string, error) {
    usersMutex.Lock()
    defer usersMutex.Unlock()

    user, exists := users[username]
    if (!exists) {
        return "", fmt.Errorf("user not found")
    }

    delete(apiKeyToUser, user.APIKey)

    newKey, err := GenerateAPIKey()
    if err != nil {
        return "", err
    }

    user.APIKey = newKey
    users[username] = user
    apiKeyToUser[newKey] = username

    if err := saveUsers(); err != nil {
        return "", err
    }

    return newKey, nil
}

func UserExists(username string) bool {
    usersMutex.RLock()
    _, exists := users[username]
    usersMutex.RUnlock()
    return exists
}

// DeleteUser removes a user from the system
func DeleteUser(username string) error {
    usersMutex.Lock()
    defer usersMutex.Unlock()

    // Check if user exists
    user, exists := users[username]
    if !exists {
        return fmt.Errorf("user not found: %s", username)
    }

    // Remove user from maps
    delete(users, username)
    delete(apiKeyToUser, user.APIKey)

    // Save changes to disk
    if err := saveUsers(); err != nil {
        return fmt.Errorf("failed to save users after deletion: %w", err)
    }

    // Load config for storage directory
    appConfig, err := config.LoadConfig()
    if err != nil {
        log.Printf("Warning: Failed to load config to delete user storage: %v", err)
        return nil // User is deleted from memory and disk, so return success
    }

    // Delete user storage directory if it exists
    userStorageDir := filepath.Join(appConfig.StorageDirectory, username)
    if _, err := os.Stat(userStorageDir); err == nil {
        if err := os.RemoveAll(userStorageDir); err != nil {
            // Log this error but don't fail the operation
            log.Printf("Warning: Failed to delete user's storage directory: %v", err)
        }
    }

    return nil
}