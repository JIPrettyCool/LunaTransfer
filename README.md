Create User

```
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"test1"}'
```

Login (GET API KEY)
```
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test1","password":"test1"}'
```

Upload File
```
curl -X POST http://localhost:8080/upload \
  -H "Username: test1" \
  -H "API-Key: YOUR_API_KEY" \
  -F "file=@/path/to/your/file.txt"
```

Download File
```
curl -X GET http://localhost:8080/download/file.txt \
  -H "Username: test1" \
  -H "API-Key: YOUR_API_KEY" \
  --output downloaded_file.txt
```

Delete File
```
curl -X DELETE http://localhost:8080/delete/file.txt \
  -H "Username: test1" \
  -H "API-Key: YOUR_API_KEY"
```