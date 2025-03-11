package models

import "time"

type TransferActivity struct {
    Timestamp int64  `json:"timestamp"`
    Operation string `json:"operation"`
    Filename  string `json:"filename"`
    Size      int64  `json:"size,omitempty"`
}

type FileActivity struct {
    Timestamp int64  `json:"timestamp"`
    Operation string `json:"operation"`
    Filename  string `json:"filename"`
    Username  string `json:"username"`
    RemoteIP  string `json:"remoteIP"`
    Size      int64  `json:"size,omitempty"`
}

type NotificationType string

const (
    NoteFileUploaded NotificationType = "FILE_UPLOADED"
    NoteFileDeleted  NotificationType = "FILE_DELETED"
    NoteFileAccessed NotificationType = "FILE_ACCESSED"
)

type Notification struct {
    Type      NotificationType `json:"type"`
    Timestamp time.Time        `json:"timestamp"`
    Filename  string           `json:"filename,omitempty"`
    Message   string           `json:"message"`
}