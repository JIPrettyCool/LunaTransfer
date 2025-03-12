# **LunaTransfer**

Self Hostable Managed File Transfer app

## API Usage Examples

### Create User

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123","email":"test@example.com","role":"user"}'
```

### Login (Get API Key)

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"Test1Password123"}'
```

### Upload File

```bash
curl -X POST http://localhost:8080/api/upload \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "file=@/path/to/your/file.txt"
```

### List Files

```bash
curl -X GET http://localhost:8080/api/files \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Download File

```bash
curl -X GET http://localhost:8080/api/download/file.txt \
  -H "Authorization: Bearer YOUR_API_KEY" \
  --output downloaded_file.txt
```

### Delete File

```bash
curl -X DELETE http://localhost:8080/api/delete/file.txt \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Get User Dashboard

```bash
curl -X GET http://localhost:8080/api/dashboard \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### WebSocket Connection (for real-time notifications)

```bash
# Using a WebSocket client like wscat
wscat -c "ws://localhost:8080/ws" -H "Authorization: Bearer YOUR_API_KEY"
```

## TODO
[View my Notion page](https://jiprettycool.notion.site/)
