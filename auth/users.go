package auth

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "errors"
    "golang.org/x/crypto/bcrypt"
    "os"
    "path/filepath"
)

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrUserExists      = errors.New("user already exists")
    ErrInvalidPassword = errors.New("invalid password")
    ErrInternalServer  = errors.New("internal server error")
)

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
    APIKey   string `json:"apiKey"`
}

func GetUser(username string) (*User, error) {
    file, err := os.Open(UsersDB())
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrUserNotFound
        }
        return nil, ErrInternalServer
    }
    defer file.Close()

    var users []User
    err = json.NewDecoder(file).Decode(&users)
    if err != nil {
        return nil, ErrInternalServer
    }

    for _, user := range users {
        if user.Username == username {
            return &user, nil
        }
    }

    return nil, ErrUserNotFound
}

func LoadUsers() ([]User, error) {
    file, err := os.Open(UsersDB())
    if err != nil {
        if os.IsNotExist(err) {
            return []User{}, nil
        }
        return nil, ErrInternalServer
    }
    defer file.Close()

    var users []User
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&users)
    if err != nil {
        return nil, ErrInternalServer
    }

    return users, nil
}

func SaveUser(user *User) error {
    err := os.MkdirAll(filepath.Dir(UsersDB()), 0755)
    if err != nil {
        return ErrInternalServer
    }

    var users []User
    
    file, err := os.OpenFile(UsersDB(), os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return ErrInternalServer
    }
    defer file.Close()

    fileInfo, err := file.Stat()
    if err != nil {
        return ErrInternalServer
    }

    if fileInfo.Size() > 0 {
        err = json.NewDecoder(file).Decode(&users)
        if err != nil {
            users = []User{}
        }
    }

    userExists := false
    for i, u := range users {
        if u.Username == user.Username {
            users[i] = *user
            userExists = true
            break
        }
    }

    if !userExists {
        users = append(users, *user)
    }

    file.Seek(0, 0)
    file.Truncate(0)
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(users)
}

func CreateUser(username, password string) (*User, error) {
    existingUser, err := GetUser(username)
    if err == nil && existingUser != nil {
        return nil, ErrUserExists
    }

    hashedPassword, err := HashPass(password)
    if err != nil {
        return nil, ErrInternalServer
    }

    apiKey, err := GenerateAPIKey()
    if err != nil {
        return nil, ErrInternalServer
    }

    user := &User{
        Username: username,
        Password: hashedPassword,
        APIKey:   apiKey,
    }

    err = SaveUser(user)
    if err != nil {
        return nil, ErrInternalServer
    }

    return user, nil
}

func AuthUser(username, password string) (*User, error) {
    user, err := GetUser(username)
    if err != nil {
        return nil, ErrUserNotFound
    }

    err = DecodePass(user.Password, password)
    if err != nil {
        return nil, ErrInvalidPassword
    }

    return user, nil
}

func GenerateAPIKey() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", ErrInternalServer
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

func HashPass(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", ErrInternalServer
    }
    return string(hashedBytes), nil
}

func DecodePass(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func UsersDB() string {
    return "users.json"
}