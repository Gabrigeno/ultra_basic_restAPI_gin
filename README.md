## Run restAPI
```
- sudo systemctl start redis-server
- go run main.go
```
## try the server
GET
```
- curl http://localhost:8080/items
- curl http://localhost:8080/items/<id>
- curl http://localhost:8080/items/search?name=Item
```
POST
```
- curl -X POST http://localhost:8080/items -H "Content-Type: application/json" -d '{"name": "New Item"}'
```
the new file ID is set automatically

PUT
```
- curl -X PUT http://localhost:8080/items/3 -H "Content-Type: application/json" -d '{"name": "Updated Item"}'
```
DELETE
```
- curl -X DELETE http://localhost:8080/items/<id>
```

## Docs
to generate
```
- go install github.com/swaggo/swag/cmd/swag@latest
```
after put the swag code in your controllers and then
```
- swag init
```

after search in your browser
```
http://localhost:8080/swagger/index.html
```

## test

```
go test ./...
```

## Docker

```
- docker build -t my-gin-api .
-  docker run -d --name my-app-container -p 8080:8080 --network="host" my-gin-api
```
or

```
- docker compose up
```
