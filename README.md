# Go Examples L0


## инструкция

```
# Сборка и запуск всех сервисов
docker-compose up --build

# После запуска PostgreSQL
migrate -path=./migrations -database="postgres://wbexaml0db:wbexam@localhost/wbexaml0db?sslmode=disable" up

# Основной сервис (HTTP API + Kafka Consumer)
go run cmd/wb-examples-l0/main.go

# Producer для тестовых данных
go run cmd/producer/main.go
```