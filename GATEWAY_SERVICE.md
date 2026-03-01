# 🌐 Gateway Service

## 📋 Описание

**Gateway Service** — центральный шлюз (API Gateway) системы STALKnet, который маршрутизирует все входящие запросы к соответствующим микросервисам.

Сервис отвечает за:
- ✅ Маршрутизацию запросов к микросервисам
- ✅ Раздачу статического контента (веб-клиент)
- ✅ CORS (Cross-Origin Resource Sharing)
- ✅ Логирование запросов
- ✅ JWT аутентификацию
- ✅ Reverse proxy для микросервисов
- ✅ WebSocket проксирование

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP/WS      ┌──────────────────┐
│ Web Client  │ ───────────────► │ Gateway Service  │
│  (Browser)  │  Port: 8080      │ Port: 8080       │
└─────────────┘                  └────────┬─────────┘
                                          │
         ┌────────────────────────────────┼────────────────────────────────┐
         │                                │                                │
         ▼                                ▼                                ▼
┌──────────────────┐            ┌──────────────────┐            ┌──────────────────┐
│   Auth Service   │            │   User Service   │            │   Chat Service   │
│   Port: 8081     │            │   Port: 8082     │            │   Port: 8083     │
└──────────────────┘            └──────────────────┘            └──────────────────┘
         │                                │                                │
         ▼                                ▼                                ▼
┌──────────────────┐            ┌──────────────────┐
│   Task Service   │            │ Notification     │
│   Port: 8084     │            │ Service :8085    │
└──────────────────┘            └──────────────────┘
```

### Поток данных

#### Статический контент:
1. Клиент запрашивает `/` или `/app.js`
2. Gateway возвращает файлы из embedded FS
3. Запрещает кэширование (no-cache)

#### API запросы:
1. Клиент отправляет запрос на `/api/auth/*`
2. Gateway проксирует запрос на Auth Service
3. Возвращает ответ клиенту

#### WebSocket:
1. Клиент подключается к `/ws/chat`
2. Gateway проксирует соединение на Chat Service
3. Chat Service обрабатывает WebSocket

---

## 🔧 API Endpoints

### 🌐 Статический контент

#### GET /

**Получение HTML страницы веб-клиента**

**Response (200 OK):**
```
Content-Type: text/html; charset=utf-8
Cache-Control: no-store, no-cache, must-revalidate, max-age=0
Pragma: no-cache
Expires: 0

<!DOCTYPE html>
<html lang="ru">
...
</html>
```

---

#### GET /app.js

**Получение JavaScript файла веб-клиента**

**Response (200 OK):**
```
Content-Type: application/javascript
Cache-Control: no-store, no-cache, must-revalidate, max-age=0
Pragma: no-cache
Expires: 0

// STALKnet Web Client JavaScript
...
```

---

### 🔐 Auth Service Proxy

Все запросы на `/api/auth/*` проксируются на Auth Service (порт 8081).

#### POST /api/auth/register

**Регистрация нового пользователя**

Проксируется на: `POST http://auth:8081/api/auth/register`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### POST /api/auth/login

**Вход в систему**

Проксируется на: `POST http://auth:8081/api/auth/login`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### POST /api/auth/logout

**Выход из системы**

Проксируется на: `POST http://auth:8081/api/auth/logout`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### POST /api/auth/refresh

**Обновление access токена**

Проксируется на: `POST http://auth:8081/api/auth/refresh`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### POST /api/auth/validate

**Проверка валидности токена**

Проксируется на: `POST http://auth:8081/api/auth/validate`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### POST /api/auth/check-username

**Проверка существования пользователя**

Проксируется на: `POST http://auth:8081/api/auth/check-username`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### GET /api/auth/session

**Получение информации о сессии**

Проксируется на: `GET http://auth:8081/api/auth/session`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

#### PUT /api/auth/update-username

**Смена имени пользователя**

Проксируется на: `PUT http://auth:8081/api/auth/update-username`

См. [AUTH_SERVICE.md](AUTH_SERVICE.md)

---

### 👤 User Service Proxy

Все запросы на `/api/user/*` проксируются на User Service (порт 8082).

**Требуется JWT аутентификация!**

#### GET /api/user/profile/:id

**Получение профиля пользователя**

Проксируется на: `GET http://user:8082/api/user/profile/:id`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

#### GET /api/user/profile/me

**Получение своего профиля**

Проксируется на: `GET http://user:8082/api/user/profile/me`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

#### PUT /api/user/profile/me

**Обновление своего профиля**

Проксируется на: `PUT http://user:8082/api/user/profile/me`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

#### GET /api/user/status

**Получение статуса**

Проксируется на: `GET http://user:8082/api/user/status`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

#### PUT /api/user/status

**Установка статуса**

Проксируется на: `PUT http://user:8082/api/user/status`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

#### GET /api/user/online

**Получение онлайн-пользователей**

Проксируется на: `GET http://user:8082/api/user/online`

См. [USER_SERVICE.md](USER_SERVICE.md)

---

### 💬 Chat Service Proxy

Все запросы на `/api/chat/*` проксируются на Chat Service (порт 8083).

**Требуется JWT аутентификация!**

#### GET /api/chat/rooms

**Получение списка комнат**

Проксируется на: `GET http://chat:8083/api/chat/rooms`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

---

#### POST /api/chat/rooms

**Создание комнаты**

Проксируется на: `POST http://chat:8083/api/chat/rooms`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

---

#### GET /api/chat/rooms/:id/messages

**Получение сообщений комнаты**

Проксируется на: `GET http://chat:8083/api/chat/rooms/:id/messages`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

---

#### POST /api/chat/rooms/:id/messages

**Отправка сообщения**

Проксируется на: `POST http://chat:8083/api/chat/rooms/:id/messages`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

---

#### GET /api/chat/rooms/:id/members

**Получение участников комнаты**

Проксируется на: `GET http://chat:8083/api/chat/rooms/:id/members`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

---

### 📋 Task Service Proxy

Все запросы на `/api/task/*` проксируются на Task Service (порт 8084).

**Требуется JWT аутентификация!**

#### GET /api/task

**Получение списка задач**

Проксируется на: `GET http://task:8084/api/task`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### POST /api/task

**Создание задачи**

Проксируется на: `POST http://task:8084/api/task`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### GET /api/task/:id

**Получение задачи по ID**

Проксируется на: `GET http://task:8084/api/task/:id`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### PUT /api/task/:id

**Обновление задачи**

Проксируется на: `PUT http://task:8084/api/task/:id`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### PUT /api/task/:id/complete

**Завершение задачи**

Проксируется на: `PUT http://task:8084/api/task/:id/complete`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### PUT /api/task/:id/confirm

**Подтверждение задачи**

Проксируется на: `PUT http://task:8084/api/task/:id/confirm`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

#### DELETE /api/task/:id

**Удаление задачи**

Проксируется на: `DELETE http://task:8084/api/task/:id`

См. [TASK_SERVICE.md](TASK_SERVICE.md)

---

### 🔔 Notification Service Proxy

Все запросы на `/api/notification/*` проксируются на Notification Service (порт 8085).

**Требуется JWT аутентификация!**

#### GET /api/notification/unread

**Получение непрочитанных уведомлений**

Проксируется на: `GET http://notification:8085/api/notification/unread`

См. [NOTIFICATION_SERVICE.md](NOTIFICATION_SERVICE.md)

---

#### PUT /api/notification/unread/:id/read

**Отметка уведомления как прочитанного**

Проксируется на: `PUT http://notification:8085/api/notification/unread/:id/read`

См. [NOTIFICATION_SERVICE.md](NOTIFICATION_SERVICE.md)

---

#### PUT /api/notification/read-all

**Отметка всех уведомлений как прочитанные**

Проксируется на: `PUT http://notification:8085/api/notification/read-all`

См. [NOTIFICATION_SERVICE.md](NOTIFICATION_SERVICE.md)

---

### 🔌 WebSocket

#### GET /ws/chat

**WebSocket подключение к чату**

Проксируется на: `WS http://chat:8083/ws/chat`

См. [CHAT_SERVICE.md](CHAT_SERVICE.md)

**Query Parameters:**
- `room_id` — ID комнаты
- `user_id` — ID пользователя
- `username` — Имя пользователя

**Пример:**
```javascript
const ws = new WebSocket(
  `ws://localhost:8080/ws/chat?room_id=1&user_id=5&username=BG`
);
```

---

#### GET /ws/notification

**WebSocket подключение к уведомлениям**

Проксируется на: `WS http://notification:8085/ws/notification`

См. [NOTIFICATION_SERVICE.md](NOTIFICATION_SERVICE.md)

**Query Parameters:**
- `user_id` — ID пользователя

**Пример:**
```javascript
const ws = new WebSocket(
  `ws://localhost:8080/ws/notification?user_id=5`
);
```

---

### 📚 Static Content API

#### GET /api/content/:key

**Получение статического контента из базы данных**

Проксируется на: `GET http://auth:8081/api/content/:key`

**Query Parameters:**
- `auth_state` — статус авторизации (0=Guest, 4=Authorized)

**Пример:**
```bash
# Справка для гостя
curl "http://localhost:8080/api/content/help_guest?auth_state=0"

# Справка для авторизованного
curl "http://localhost:8080/api/content/help_authorized?auth_state=4"
```

**Response (200 OK):**
```json
{
  "key": "help_guest",
  "type": "text",
  "title": "Базовые команды",
  "content": "╭────────────────────────────────────────────╮\n│ Доступные команды:\n│ /help - Показать эту справку\n│ ..."
}
```

---

### ❤️ Health Check

#### GET /health

**Проверка здоровья Gateway**

**Response (200 OK):**
```json
{"status": "ok"}
```

---

## 🔧 Middleware

### CORS

Добавляет заголовки для поддержки Cross-Origin запросов:

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        
        c.Next()
    }
}
```

---

### Logging

Логирует все входящие запросы:

```
[200] POST /api/auth/login 15.234ms
[401] GET /api/user/profile/me 2.145ms
[200] GET /ws/chat?room_id=1 0.523ms
```

---

### JWTAuth

Проверяет JWT токен для защищённых endpoints:

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
            c.Abort()
            return
        }
        
        // Проверка JWT токена
        // ...
        
        c.Next()
    }
}
```

---

### Proxy

Создаёт middleware для reverse proxy:

```go
func Proxy(proxy *httputil.ReverseProxy) gin.HandlerFunc {
    return func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps gateway

# Логи
docker-compose logs -f gateway
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8080` |
| `AUTH_SERVICE_URL` | URL Auth Service | `http://localhost:8081` |
| `USER_SERVICE_URL` | URL User Service | `http://localhost:8082` |
| `CHAT_SERVICE_URL` | URL Chat Service | `http://localhost:8083` |
| `TASK_SERVICE_URL` | URL Task Service | `http://localhost:8084` |
| `NOTIFICATION_SERVICE_URL` | URL Notification Service | `http://localhost:8085` |
| `COMPLIANCE_SERVICE_URL` | URL Compliance Service | `http://localhost:8086` |
| `JWT_SECRET` | Секрет JWT токенов | `your-secret-key` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка логов

```bash
# Логи в реальном времени
docker-compose logs -f gateway

# Последние 100 строк
docker-compose logs --tail=100 gateway
```

### Статистика запросов

```bash
# Количество запросов по статусам
docker-compose logs gateway | grep -E "^\[([0-9]+)\]" | \
  awk -F'[][]' '{print $2}' | sort | uniq -c | sort -rn
```

---

## 📝 Примеры использования

### Получение веб-клиента

```bash
curl http://localhost:8080/
```

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"BG","password":"password123"}'
```

### Вход

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"BG","password":"password123"}'
```

### Получение профиля (с JWT)

```bash
curl "http://localhost:8080/api/user/profile/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### WebSocket подключение

```javascript
const ws = new WebSocket("ws://localhost:8080/ws/chat?room_id=1&user_id=5&username=BG");

ws.onopen = () => console.log("Connected");
ws.onmessage = (e) => console.log("Message:", e.data);
ws.send(JSON.stringify({ type: "message", content: "Привет!" }));
```

---

## 📊 Маршруты Gateway

| Метод | Путь | Прокси | Auth |
|-------|------|--------|------|
| GET | `/` | Static (index.html) | ❌ |
| GET | `/app.js` | Static (app.js) | ❌ |
| GET | `/health` | Local | ❌ |
| POST | `/api/auth/*` | Auth Service (8081) | ❌ |
| GET | `/api/auth/*` | Auth Service (8081) | ❌ |
| PUT | `/api/auth/*` | Auth Service (8081) | ❌ |
| GET | `/api/content/*` | Auth Service (8081) | ❌ |
| GET | `/api/user/*` | User Service (8082) | ✅ |
| PUT | `/api/user/*` | User Service (8082) | ✅ |
| GET | `/api/chat/*` | Chat Service (8083) | ✅ |
| POST | `/api/chat/*` | Chat Service (8083) | ✅ |
| GET | `/api/task/*` | Task Service (8084) | ✅ |
| POST | `/api/task/*` | Task Service (8084) | ✅ |
| PUT | `/api/task/*` | Task Service (8084) | ✅ |
| DELETE | `/api/task/*` | Task Service (8084) | ✅ |
| GET | `/api/notification/*` | Notification Service (8085) | ✅ |
| PUT | `/api/notification/*` | Notification Service (8085) | ✅ |
| GET | `/ws/chat` | Chat Service (8083) | ❌ |
| GET | `/ws/notification` | Notification Service (8085) | ❌ |

---

## ⚠️ Важные замечания

1. **Статический контент** — файлы встроены через `embed.FS`
2. **Кэширование** — запрещено для index.html и app.js
3. **CORS** — разрешены все origin (для разработки)
4. **JWT проверка** — требуется для `/api/user/*`, `/api/chat/*`, `/api/task/*`, `/api/notification/*`
5. **WebSocket** — проксируется без проверки токена (авторизация внутри сервиса)
6. **Reverse proxy** — используется `httputil.NewSingleHostReverseProxy`

---

## 📈 Производительность

### Рекомендации

1. **Keep-Alive** — включены для прокси соединений
2. **Gzip** — рекомендуется включить для сжатия
3. **Rate limiting** — рекомендуется добавить для защиты от DDoS
4. **Кэширование статики** — app.js можно кэшировать с версионированием

### Метрики

- **RPS** — запросов в секунду
- **Latency** — среднее время обработки запроса
- **Connections** — количество одновременных соединений
- **WebSocket** — количество активных WebSocket соединений

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs gateway`
2. Проверьте health: `curl http://localhost:8080/health`
3. Проверьте доступность сервисов: `curl http://auth:8081/health`

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** ✅ Работает
