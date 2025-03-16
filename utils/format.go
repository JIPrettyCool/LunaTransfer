package utils

import (
    "fmt"
    "crypto/rand"
    "time"
)

func FormatFileSize(size int64) string {
    const unit = 1024
    if size < unit {
        return fmt.Sprintf("%d B", size)
    }
    
    div, exp := int64(unit), 0
    for n := size / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    
    return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func GenerateUUID() string {
    uuid := make([]byte, 16)
    _, err := rand.Read(uuid)
    if err != nil {
        return fmt.Sprintf("%d", time.Now().UnixNano())
    }
    
    // Set version (4) and variant (RFC4122)
    uuid[6] = (uuid[6] & 0x0f) | 0x40
    uuid[8] = (uuid[8] & 0x3f) | 0x80
    
    return fmt.Sprintf("%x-%x-%x-%x-%x", 
        uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}