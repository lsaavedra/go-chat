# go-chat

Project to create a simple-browser chat application with a minimal frontend and pure backend.

### Diagram architecture

![Overview]([https://user-images.githubusercontent.com/9138602/184556862-0091bffb-9cc7-4eb3-837c-cbc0b97d5186.jpg))

##### Run Project
1. Use docker-compose command to start services (Postgres, Redis and RabbitMQ):
```
# docker-compose up
```
2. Next you can start the server with:
```
SERVER_HOST=localhost \
SERVER_PORT=8080 \
QUEUE_URL=amqp://guest:guest@localhost:5672/ \
CACHE_URL=localhost:6379 \
DB_HOST=localhost \
DB_PORT=7004 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=postgres \
DB_SCHEMA=chatroom \
CACHE_URL=localhost:6379 \
go run cmd/main.go
```
3. Create two random users with this curl command:
```
curl --request POST \
  --url http://localhost:8080/api/v1/users \
  --header 'Content-Type: application/json' \
  --data '{
	"first_name": "any_first_name",
	"last_name": "any_last_name",
	"email": "any_email",
	"nick_name": "any_nickname",
	"password": "any_pwd"
}'
```

4. Open a web browser and use this url http://localhost:8080/login (use email and pwd for the users created in above bullet)

