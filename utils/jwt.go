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
    jwtSecret        = []byte("luna-transfer-secret-key")
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token expired")
    tokenExpiryHours = 24
    blacklistedTokens = make(map[string]time.Time)
    tokenMutex       = &sync.Mutex{}
)

type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func InitJWT(cfg *config.AppConfig) {
    if cfg.JWTSecret != "" {
        jwtSecret = []byte(cfg.JWTSecret)
    }
    if cfg.TokenExpiryHours > 0 {
        tokenExpiryHours = cfg.TokenExpiryHours
    }
    LogSystem("JWT_INIT", "system", "localhost", fmt.Sprintf("JWT initialized with %d hour expiry", tokenExpiryHours))
}

func GenerateJWT(username, role string) (string, error) {
    expiryTime := time.Now().Add(time.Duration(tokenExpiryHours) * time.Hour)
    
    claims := &Claims{
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiryTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "LunaTransfer",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func BlacklistToken(tokenString string, expiresAt time.Time) {
    tokenMutex.Lock()
    defer tokenMutex.Unlock()
    
    blacklistedTokens[tokenString] = expiresAt
}

func IsTokenBlacklisted(tokenString string) bool {
    tokenMutex.Lock()
    defer tokenMutex.Unlock()
    now := time.Now()
    for token, expires := range blacklistedTokens {
        if now.After(expires) {
            delete(blacklistedTokens, token)
        }
    }
    
    _, found := blacklistedTokens[tokenString]
    return found
}

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