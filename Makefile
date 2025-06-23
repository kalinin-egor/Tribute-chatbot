.PHONY: build run test clean deps lint help docker-build docker-run docker-dev docker-stop docker-logs docker-clean

# Переменные
BINARY_NAME=tribute-chatbot
BUILD_DIR=build
MAIN_FILE=main.go
DOCKER_IMAGE=tribute-chatbot

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

deps: ## Установить зависимости
	@echo "$(GREEN)Устанавливаем зависимости...$(NC)"
	go mod tidy
	go mod download

build: deps ## Собрать приложение
	@echo "$(GREEN)Собираем приложение...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Приложение собрано: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

run: ## Запустить приложение
	@echo "$(GREEN)Запускаем приложение...$(NC)"
	go run $(MAIN_FILE)

dev: ## Запустить в режиме разработки
	@echo "$(GREEN)Запускаем в режиме разработки...$(NC)"
	LOG_LEVEL=debug go run $(MAIN_FILE)

test: ## Запустить тесты
	@echo "$(GREEN)Запускаем тесты...$(NC)"
	go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запускаем тесты с покрытием...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

lint: ## Проверить код линтером
	@echo "$(GREEN)Проверяем код...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint не установлен. Устанавливаем...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

clean: ## Очистить сборки
	@echo "$(GREEN)Очищаем сборки...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: build ## Установить приложение
	@echo "$(GREEN)Устанавливаем приложение...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Приложение установлено в /usr/local/bin/$(BINARY_NAME)$(NC)"

# Docker команды
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Собираем Docker образ...$(NC)"
	docker build -t $(DOCKER_IMAGE):latest .

docker-run: ## Запустить в Docker
	@echo "$(GREEN)Запускаем в Docker...$(NC)"
	docker run --env-file config.env -p 8080:8080 --name $(BINARY_NAME) $(DOCKER_IMAGE):latest

docker-compose-up: ## Запустить с docker-compose
	@echo "$(GREEN)Запускаем с docker-compose...$(NC)"
	docker-compose up -d

docker-compose-down: ## Остановить docker-compose
	@echo "$(GREEN)Останавливаем docker-compose...$(NC)"
	docker-compose down

docker-compose-logs: ## Показать логи docker-compose
	@echo "$(GREEN)Показываем логи docker-compose...$(NC)"
	docker-compose logs -f

docker-dev: ## Запустить в режиме разработки с docker-compose
	@echo "$(GREEN)Запускаем в режиме разработки с docker-compose...$(NC)"
	docker-compose -f docker-compose.dev.yml up -d

docker-dev-down: ## Остановить режим разработки
	@echo "$(GREEN)Останавливаем режим разработки...$(NC)"
	docker-compose -f docker-compose.dev.yml down

docker-dev-logs: ## Показать логи режима разработки
	@echo "$(GREEN)Показываем логи режима разработки...$(NC)"
	docker-compose -f docker-compose.dev.yml logs -f

docker-stop: ## Остановить все контейнеры
	@echo "$(GREEN)Останавливаем все контейнеры...$(NC)"
	docker stop $(BINARY_NAME) 2>/dev/null || true
	docker-compose down 2>/dev/null || true
	docker-compose -f docker-compose.dev.yml down 2>/dev/null || true

docker-logs: ## Показать логи контейнера
	@echo "$(GREEN)Показываем логи контейнера...$(NC)"
	docker logs -f $(BINARY_NAME)

docker-clean: ## Очистить Docker образы и контейнеры
	@echo "$(GREEN)Очищаем Docker...$(NC)"
	docker stop $(BINARY_NAME) 2>/dev/null || true
	docker rm $(BINARY_NAME) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):latest 2>/dev/null || true
	docker system prune -f

fmt: ## Форматировать код
	@echo "$(GREEN)Форматируем код...$(NC)"
	go fmt ./...

vet: ## Проверить код с go vet
	@echo "$(GREEN)Проверяем код с go vet...$(NC)"
	go vet ./...

all: clean deps lint test build ## Выполнить все проверки и сборку
	@echo "$(GREEN)Все готово!$(NC)"

docker-all: docker-clean docker-build docker-compose-up ## Собрать и запустить в Docker
	@echo "$(GREEN)Docker приложение готово!$(NC)" 