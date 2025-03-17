# **LunaTransfer**

Self Hostable Managed File Transfer app

## Configuration

LunaTransfer can be configured by editing the config file before startup. The following settings can be customized:

- Port number
- Storage path
- Log directory
- JWT secret and expiration time
- Rate limiting settings

## API Usage Examples

### Initial Setup and Authentication

#### Initial Setup (Admin signup)

```bash
curl -X POST http://localhost:8080/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "AdminPassword123!",
    "email": "admin@example.com"
  }'
```

#### Signup

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123","email":"test@example.com","role":"user"}'
```

#### Login (Get JWT Token)

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123"}'
```

#### Logout

```bash
curl -X POST http://localhost:8080/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Refresh Token

```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### File Operations

#### Upload File

```bash
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.txt" \
  -F "path=photos/vacation2023"
```

#### List Files

```bash
curl -X GET "http://localhost:8080/api/files?path=photos/vacation2023" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Search Files

```bash
curl -X GET "http://localhost:8080/api/search?term=project" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

  # Search in specific directory
curl -X GET "http://localhost:8080/api/search?term=report&path=documents" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Search by file type (extension)
curl -X GET "http://localhost:8080/api/search?term=data&type=csv" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Search by file size (in bytes)
curl -X GET "http://localhost:8080/api/search?term=video&minSize=1000000&maxSize=5000000" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Search by date modified
curl -X GET "http://localhost:8080/api/search?term=report&after=2005-08-08&before=2005-08-08" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Combining multiple filters
curl -X GET "http://localhost:8080/api/search?term=presentation&path=work&type=pptx&minSize=500000" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Download File

```bash
curl -X GET http://localhost:8080/api/download/file.txt \
  -H "Authorization: Bearer YOUR_JWT_KEY" \
  --output downloaded_file.txt
```

### Delete File

```bash
curl -X DELETE http://localhost:8080/api/delete/file.txt \
  -H "Authorization: Bearer YOUR_JWT_KEY"
```


#### Create Directory

```bash
curl -X POST http://localhost:8080/api/directory \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"photos", "name":"vacation2023"}'
```

### Admin Operations

#### List Users (Admin Only)

```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

#### Delete User (Admin Only)

```bash
curl -X DELETE http://localhost:8080/api/admin/users/username \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

#### System Stats (Admin Only)

```bash
curl -X GET http://localhost:8080/api/admin/system/stats \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Groups

#### Create Group (Admin Only)

```bash
curl -X POST http://localhost:8080/api/admin/groups \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Marketing Team",
    "description": "Group for marketing department files"
  }'
```

#### List Groups (Admin Only)

```bash
curl -X GET http://localhost:8080/api/admin/groups \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

#### Add User to Group (Admin Only)

```bash
curl -X POST "http://localhost:8080/api/admin/groups/YOUR_GROUP_ID/members" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "role": "member"
  }'
```

#### Remove User from Group (Admin Only)

```bash
curl -X DELETE "http://localhost:8080/api/admin/groups/YOUR_GROUP_ID/members/username" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

#### List Group Members (Admin Only)

```bash
curl -X GET "http://localhost:8080/api/admin/groups/YOUR_GROUP_ID/members" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

#### Upload File to Group Directory (Group Members Only)

```bash
curl -X POST http://localhost:8080/api/upload/group \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.txt" \
  -F "groupId=YOUR_GROUP_ID" \
  -F "path=reports/monthly"
```

#### Download File from Group Directory (Group Members Only)

```bash
curl -X GET http://localhost:8080/api/download/groups/YOUR_GROUP_ID/path/to/file.txt \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --output downloaded_file.txt
```

#### List Files in Group Directory (Group Members Only)

```bash
curl -X GET "http://localhost:8080/api/files?path=groups/YOUR_GROUP_ID/reports" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Add User to Group with Specific Role (Admin Only)

```bash
curl -X POST "http://localhost:8080/api/admin/groups/YOUR_GROUP_ID/members" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "role": "contributor"
  }'
```
##### Group Role Permissions

LunaTransfer supports three levels of group roles:

- **admin**: Can manage group members and has full access to all group files
- **contributor**: Can upload, modify, and download files but cannot manage members
- **reader**: Can only view and download files

#### Share a File with Another Group

```bash
curl -X POST http://localhost:8080/api/share \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "path/to/file.txt",
    "source_group": "SOURCE_GROUP_ID",
    "target_group": "TARGET_GROUP_ID",
    "permission": "read"
  }'
```

#### List Files Shared with a Group

```bash
curl -X GET "http://localhost:8080/api/shared?groupId=YOUR_GROUP_ID" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Remove File Sharing

```bash
curl -X DELETE http://localhost:8080/api/share/SHARE_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Misc

#### Get User Dashboard

```bash
curl -X GET http://localhost:8080/api/dashboard \
  -H "Authorization: Bearer YOUR_JWT_KEY"
```

#### WebSocket Connection (for real-time notifications)

```bash
# Using a WebSocket client like wscat
wscat -c "ws://localhost:8080/ws" -H "Authorization: Bearer YOUR_JWT_KEY"
```
##### Notification Types

- **CONNECTED:** Sent when a WebSocket connection is established
- **FILE_UPLOADED:** Sent when a new file is uploaded
- **FILE_DELETED:** Sent when a file is deleted
- **SHARE_CREATED:** Sent when a file is shared with a group
- **SHARE_REMOVED:** Sent when a file share is removed

## TODO
[View my Notion page](https://jiprettycool.notion.site/)
