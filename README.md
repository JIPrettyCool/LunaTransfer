# **LunaTransfer**

Self Hostable Managed File Transfer app

## API Usage Examples

### Signup

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123","email":"test@example.com","role":"user"}'
```

### Setup (Admin signup)

```bash
curl -X POST http://localhost:8080/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "YourStrongPassword123",
    "email": "admin@example.com"
  }'
```

### Login (Get JWT Token)

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123"}'
```

### Logout

```bash
curl -X POST http://localhost:8080/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Upload File

```bash
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.txt" \
  -F "path=photos/vacation2023"
```

### List Files

```bash
curl -X GET "http://localhost:8080/api/files?path=photos/vacation2023" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Search Files

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

### Create Directory

```bash
curl -X POST http://localhost:8080/api/directory \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"path":"photos", "name":"vacation2023"}'
```

### Download File

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

### Get User Dashboard

```bash
curl -X GET http://localhost:8080/api/dashboard \
  -H "Authorization: Bearer YOUR_JWT_KEY"
```

### List Users (Admin Only)

```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### Delete User (Admin Only)

```bash
curl -X DELETE http://localhost:8080/api/admin/users/username \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### System Stats (Admin Only)

```bash
curl -X GET http://localhost:8080/api/admin/system/stats \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

### WebSocket Connection (for real-time notifications)

```bash
# Using a WebSocket client like wscat
wscat -c "ws://localhost:8080/ws" -H "Authorization: Bearer YOUR_JWT_KEY"
```

## TODO
[View my Notion page](https://jiprettycool.notion.site/)
