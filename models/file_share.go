package models

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sync"
    "time"
    
    "LunaTransfer/config"
)

type FileShare struct {
    ID          string    `json:"id"`
    FilePath    string    `json:"file_path"`
    SourceGroup string    `json:"source_group"`
    TargetGroup string    `json:"target_group"`
    Permission  string    `json:"permission"`
    SharedBy    string    `json:"shared_by"`
    SharedAt    time.Time `json:"shared_at"`
}

var (
    fileSharesMutex sync.RWMutex
    fileSharesFile  = "file_shares.json"
)

func GetFileSharePath() (string, error) {
    cfg, err := config.LoadConfig()
    if err != nil {
        return "", err
    }
    return filepath.Join(cfg.StorageDirectory, "file_shares.json"), nil
}

func LoadFileShares() ([]FileShare, error) {
    fileSharesMutex.RLock()
    defer fileSharesMutex.RUnlock()

    path, err := GetFileSharePath()
    if err != nil {
        return nil, err
    }

    if _, err := os.Stat(path); os.IsNotExist(err) {
        return []FileShare{}, nil
    }

    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var shares []FileShare
    err = json.Unmarshal(data, &shares)
    if err != nil {
        return nil, err
    }

    return shares, nil
}

func SaveFileShares(shares []FileShare) error {
    fileSharesMutex.Lock()
    defer fileSharesMutex.Unlock()

    path, err := GetFileSharePath()
    if err != nil {
        return err
    }

    data, err := json.MarshalIndent(shares, "", "  ")
    if err != nil {
        return err
    }

    return ioutil.WriteFile(path, data, 0644)
}

func SaveFileShare(share FileShare) error {
    shares, err := LoadFileShares()
    if err != nil {
        return err
    }

    shares = append(shares, share)
    return SaveFileShares(shares)
}

func GetFileSharesForGroup(groupID string) ([]FileShare, error) {
    shares, err := LoadFileShares()
    if err != nil {
        return nil, err
    }

    var groupShares []FileShare
    for _, share := range shares {
        if share.TargetGroup == groupID {
            groupShares = append(groupShares, share)
        }
    }

    return groupShares, nil
}

func GetFileShareByID(shareID string) (FileShare, error) {
    shares, err := LoadFileShares()
    if err != nil {
        return FileShare{}, err
    }

    for _, share := range shares {
        if share.ID == shareID {
            return share, nil
        }
    }

    return FileShare{}, fmt.Errorf("share not found")
}

func DeleteFileShare(shareID string) error {
    shares, err := LoadFileShares()
    if err != nil {
        return err
    }

    var newShares []FileShare
    for _, share := range shares {
        if share.ID != shareID {
            newShares = append(newShares, share)
        }
    }

    return SaveFileShares(newShares)
}