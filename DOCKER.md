# Docker Guide для Tribute Chatbot

## 🐳 Обзор

Проект поддерживает полную контейнеризацию с помощью Docker и Docker Compose. Доступны два режима:
- **Production** - оптимизированный образ для продакшена
- **Development** - режим разработки с hot reload

## 📋 Требования

- Docker 20.10+
- Docker Compose 2.0+
- Минимум 2GB RAM для разработки

## 🚀 Быстрый старт

### Production режим

```bash
# Собрать и запустить
make docker-all

# Или пошагово:
make docker-build
make docker-compose-up
```

### Development режим

```bash
# Запустить с дополнительными сервисами
make docker-dev

# Просмотр логов
make docker-dev-logs
```

## 📁 Файлы Docker

### Dockerfile
- **Многоэтапная сборка** для оптимизации размера
- **Alpine Linux** для минимального размера образа
- **Security best practices** - непривилегированный пользователь
- **Health checks** для мониторинга
- **Оптимизированные флаги сборки** для производительности

### docker-compose.yml (Production)
- Основной сервис бота
- Настройка переменных окружения
- Лимиты ресурсов
- Структурированное логирование
- Health checks
- Сетевые настройки

### docker-compose.dev.yml (Development)
- Режим разработки с hot reload
- Дополнительные сервисы:
  - Redis для кэширования
  - PostgreSQL для данных
  - pgAdmin для управления БД
- Монтирование исходного кода
- Увеличенные лимиты ресурсов

## 🛠 Команды Makefile

### Основные команды
```bash
make docker-build          # Собрать Docker образ
make docker-run            # Запустить контейнер
make docker-compose-up     # Запустить с docker-compose
make docker-compose-down   # Остановить docker-compose
make docker-compose-logs   # Показать логи
```

### Development команды
```bash
make docker-dev            # Запустить режим разработки
make docker-dev-down       # Остановить режим разработки
make docker-dev-logs       # Показать логи разработки
```

### Утилиты
```bash
make docker-stop           # Остановить все контейнеры
make docker-logs           # Показать логи контейнера
make docker-clean          # Очистить Docker
make docker-all            # Полный цикл: очистка → сборка → запуск
```

## 🔧 Конфигурация

### Переменные окружения

Создайте файл `.env` или используйте `config.env`:

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
LOG_LEVEL=info
PORT=8080
```

### Портфолио портов

| Сервис | Порт | Описание |
|--------|------|----------|
| Bot | 8080 | Основное приложение |
| Redis | 6379 | Кэширование (dev) |
| PostgreSQL | 5432 | База данных (dev) |
| pgAdmin | 5050 | Управление БД (dev) |

## 📊 Мониторинг

### Health Checks
- **Production**: проверка процесса каждые 30 секунд
- **Development**: проверка Go процесса каждые 30 секунд

### Логирование
- **Production**: JSON формат, ротация 10MB, 3 файла
- **Development**: JSON формат, ротация 20MB, 5 файлов

## 🔒 Безопасность

### Production образ
- Непривилегированный пользователь (UID 1001)
- Минимальный базовый образ (Alpine 3.18)
- Отсутствие отладочной информации в бинарнике
- Read-only файловая система где возможно

### Development образ
- Монтирование исходного кода для разработки
- Дополнительные инструменты отладки
- Увеличенные права для hot reload

## 🚀 Развертывание

### Локальное развертывание
```bash
# Production
docker-compose up -d

# Development
docker-compose -f docker-compose.dev.yml up -d
```

### Облачное развертывание
```bash
# Сборка для продакшена
docker build -t tribute-chatbot:latest .

# Push в registry
docker tag tribute-chatbot:latest your-registry/tribute-chatbot:latest
docker push your-registry/tribute-chatbot:latest
```

## 🐛 Отладка

### Просмотр логов
```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f tribute-chatbot

# Режим разработки
docker-compose -f docker-compose.dev.yml logs -f
```

### Вход в контейнер
```bash
# Production
docker exec -it tribute-chatbot sh

# Development
docker exec -it tribute-chatbot-dev sh
```

### Проверка состояния
```bash
# Статус контейнеров
docker-compose ps

# Использование ресурсов
docker stats
```

## 📈 Производительность

### Оптимизации образа
- Многоэтапная сборка
- Минимальный базовый образ
- Оптимизированные флаги компиляции
- Отсутствие отладочной информации

### Лимиты ресурсов
- **Production**: 512MB RAM, 0.5 CPU
- **Development**: 1GB RAM, 1.0 CPU

## 🔄 CI/CD

### GitHub Actions пример
```yaml
name: Build and Deploy
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t tribute-chatbot .
      - name: Push to registry
        run: |
          docker tag tribute-chatbot:latest ${{ secrets.REGISTRY }}/tribute-chatbot:latest
          docker push ${{ secrets.REGISTRY }}/tribute-chatbot:latest
```

## 🆘 Устранение неполадок

### Частые проблемы

1. **Контейнер не запускается**
   ```bash
   docker-compose logs tribute-chatbot
   ```

2. **Проблемы с токеном**
   ```bash
   # Проверьте переменную окружения
   docker-compose config
   ```

3. **Нехватка памяти**
   ```bash
   # Увеличьте лимиты в docker-compose.yml
   docker-compose down
   docker-compose up -d
   ```

### Полезные команды
```bash
# Очистка всего Docker
make docker-clean

# Пересборка без кэша
docker-compose build --no-cache

# Проверка конфигурации
docker-compose config
``` 