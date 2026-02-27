# STALKnet

Децентрализованный чат-сервис с системой задач в консольном интерфейсе.

## Архитектура

```
┌─────────────┐
│   Client    │ ──HTTP/WS──>  ┌──────────────┐
│  (Terminal) │               │ API Gateway  │
└─────────────┘               └──────┬───────┘
                                     │
         ┌──────────┬────────────────┼────────────────┬──────────┐
         ▼          ▼                ▼                ▼          ▼
   ┌─────────┐ ┌─────────┐   ┌─────────────┐  ┌───────────┐ ┌────────────┐
   │  Auth   │ │  User   │   │    Chat     │  │   Task    │ │Notification│
   │ Service │ │ Service │   │   Service   │  │  Service  │ │  Service   │
   └────┬────┘ └────┬────┘   └──────┬──────┘  └─────┬─────┘ └─────┬──────┘
        │           │               │               │             │
        └───────────┴───────────────┴───────────────┴─────────────┘
                                    │
                          ┌─────────┴─────────┐
                          ▼                   ▼
                    ┌───────────┐       ┌───────────┐
                    │PostgreSQL │       │   Redis   │
                    └───────────┘       └───────────┘
```

## Сервисы

| Сервис | Порт | Описание |
|--------|------|----------|
| Gateway | 8080 | API Gateway, роутинг, аутентификация |
| Auth | 8081 | Регистрация, логин, JWT |
| User | 8082 | Профили пользователей, статусы |
| Chat | 8083 | WebSocket, комнаты, сообщения |
| Task | 8084 | Управление задачами |
| Notification | 8085 | Уведомления |

## Быстрый старт

```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

## Структура проекта

```
STALKnet/
├── gateway/              # API Gateway
├── services/
│   ├── auth/            # Auth Service
│   ├── user/            # User Service
│   ├── chat/            # Chat Service
│   ├── task/            # Task Service
│   └── notification/    # Notification Service
├── client/              # Консольный клиент
├── pkg/                 # Общие библиотеки
└── deploy/              # Конфигурация БД
```

## API Endpoints

### Auth Service
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Вход

### Chat Service
- `WS /ws/chat` - WebSocket соединение для чата
- `GET /api/chat/rooms` - Список комнат
- `POST /api/chat/rooms` - Создать комнату

### Task Service
- `GET /api/tasks` - Список задач
- `POST /api/tasks` - Создать задачу
- `PUT /api/tasks/:id/complete` - Выполнить задачу
- `PUT /api/tasks/:id/confirm` - Подтвердить выполнение

## Технологии

- **Backend:** Go
- **База данных:** PostgreSQL 15
- **Кэш:** Redis 7
- **WebSocket:** gorilla/websocket
- **Контейнеризация:** Docker, Docker Compose
