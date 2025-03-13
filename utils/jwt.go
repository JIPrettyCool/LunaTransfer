package utils

import (
    "LunaTransfer/config"
    "errors"
    "fmt"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var (
    // Default secret key - in production this should come from config
    jwtSecret        = []byte("luna-transfer-secret-key")
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token expired")
    tokenExpiryHours = 24 // Token validity in hours
    blacklistedTokens = make(map[string]time.Time)
    tokenMutex       = &sync.Mutex{}
)

// Claims represents the JWT claims
type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// InitJWT initializes JWT configuration
func InitJWT(cfg *config.AppConfig) {  // Changed from Config to AppConfig
    if cfg.JWTSecret != "" {
        jwtSecret = []byte(cfg.JWTSecret)
    }
    if cfg.TokenExpiryHours > 0 {
        tokenExpiryHours = cfg.TokenExpiryHours
    }
    LogSystem("JWT_INIT", "system", "localhost", fmt.Sprintf("JWT initialized with %d hour expiry", tokenExpiryHours))
}

// GenerateJWT creates a new JWT token for the given username and role
func GenerateJWT(username, role string) (string, error) {
    expirationTime := time.Now().Add(time.Duration(tokenExpiryHours) * time.Hour)
    
    claims := &Claims{
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   username,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", err
    }
    
    return tokenString, nil
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenString string, expiresAt time.Time) {
    tokenMutex.Lock()
    defer tokenMutex.Unlock()
    
    blacklistedTokens[tokenString] = expiresAt
}

// IsTokenBlacklisted checks if a token is blacklisted
func IsTokenBlacklisted(tokenString string) bool {
    tokenMutex.Lock()
    defer tokenMutex.Unlock()
    
    // Cleanup expired tokens periodically
    now := time.Now()
    for token, expires := range blacklistedTokens {
        if now.After(expires) {
            delete(blacklistedTokens, token)
        }
    }
    
    _, found := blacklistedTokens[tokenString]
    return found
}

// ValidateJWT validates and parses a JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
    // Check if token is blacklisted
    if IsTokenBlacklisted(tokenString) {
        return nil, ErrInvalidToken
    }

    claims := &Claims{}
    
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })
    
    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    
    if !token.Valid {
        return nil, ErrInvalidToken
    }
    
    return claims, nil
}