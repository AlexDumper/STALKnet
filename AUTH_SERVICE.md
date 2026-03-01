# 🔐 Auth Service

## 📋 Описание

**Auth Service** — микросервис аутентификации и авторизации пользователей в системе STALKnet.

Сервис отвечает за:
- ✅ Регистрацию новых пользователей
- ✅ Аутентификацию (login/logout)
- ✅ Выдачу JWT токенов (access + refresh)
- ✅ Управление сессиями пользователей
- ✅ Проверку имени пользователя на занятость
- ✅ Смену имени пользователя
- ✅ Интеграцию с Compliance Service (ФЗ-374)

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP POST      ┌──────────────────┐
│ Web Client  │ ─────────────────► │ Auth Service     │
│  (Browser)  │  /api/auth/*       │ Port: 8081       │
└─────────────┘                    └────────┬─────────┘
                                            │
                         ┌──────────────────┴──────────────────┐
                         ▼                                     ▼
                  ┌──────────────┐                     ┌──────────────┐
                  │  PostgreSQL  │                     │    Redis     │
                  │   (users)    │                     │  (sessions)  │
                  └──────────────┘                     └──────────────┘
                                            │
                                            ▼
                                   ┌──────────────────┐
                                   │ Compliance       │
                                   │ Service :8086    │
                                   └──────────────────┘
```

### Поток данных

#### Регистрация:
1. Пользователь отправляет POST `/api/auth/register`
2. Auth Service проверяет существование пользователя
3. Хэширует пароль (bcrypt)
4. Создаёт пользователя в PostgreSQL
5. Отправляет событие `CREATE` в Compliance Service

#### Вход (Login):
1. Пользователь отправляет POST `/api/auth/login`
2. Auth Service проверяет учётные данные
3. Генерирует JWT access + refresh токены
4. Сохраняет сессию в Redis
5. Отправляет событие `LOGIN` в Compliance Service

#### Выход (Logout):
1. Пользователь отправляет POST `/api/auth/logout`
2. Auth Service удаляет сессию из Redis
3. Отправляет событие `LOGOUT` в Compliance Service

---

## 🔧 API Endpoints

### 📝 Регистрация и аутентификация

#### POST /api/auth/register

**Регистрация нового пользователя**

**Request:**
```json
{
  "username": "BG",
  "password": "password123",
  "email": "bg@stalknet.com"
}
```

**Требования к учётным данным:**
- **Username:** 2-50 символов
- **Password:** 6-100 символов
- **Email:** опционально

**Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user_id": 5,
  "username": "BG"
}
```

**Возможные ошибки:**
- `400 Bad Request` — некорректные данные
- `409 Conflict` — имя занято
- `500 Internal Server Error` — ошибка БД

---

#### POST /api/auth/login

**Вход в систему (получение JWT токенов)**

**Request:**
```json
{
  "username": "BG",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123def456..."
}
```

**Параметры токенов:**
- **Access token:** 15 минут
- **Refresh token:** 7 дней
- **Session ID:** уникальный идентификатор сессии

---

#### POST /api/auth/logout

**Выход из системы (завершение сессии)**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

#### POST /api/auth/refresh

**Обновление access токена**

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123..."
}
```

---

#### POST /api/auth/validate

**Проверка валидности токена**

**Request:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200 OK):**
```json
{
  "valid": true,
  "user_id": 5,
  "username": "BG"
}
```

---

#### POST /api/auth/check-username

**Проверка существования пользователя**

**Request:**
```json
{
  "username": "BG"
}
```

**Response (200 OK):**

Пользователь существует:
```json
{
  "exists": true,
  "username": "BG"
}
```

Пользователь не найден:
```json
{
  "exists": false
}
```

---

#### GET /api/auth/session

**Получение информации о текущей сессии**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "session_id": "BG_abc123def456",
  "user_id": 5,
  "username": "BG",
  "expires_at": "2026-03-01T12:15:00Z"
}
```

---

#### PUT /api/auth/update-username

**Смена имени пользователя**

**Request:**
```json
{
  "user_id": 5,
  "new_username": "NewBG"
}
```

**Response (200 OK):**
```json
{
  "message": "Username updated successfully",
  "user_id": 5,
  "old_username": "BG",
  "new_username": "NewBG"
}
```

**Возможные ошибки:**
- `400 Bad Request` — некорректные данные
- `404 Not Found` — пользователь не найден
- `409 Conflict` — новое имя занято

---

### 📚 Статический контент

#### GET /api/content/:key

**Получение статического контента (справка и т.д.)**

**Query Parameters:**
- `auth_state` — статус авторизации (0=Guest, 4=Authorized)

**Пример:**
```bash
# Справка для гостя
curl "http://localhost:8081/api/content/help_guest?auth_state=0"

# Справка для авторизованного
curl "http://localhost:8081/api/content/help_authorized?auth_state=4"
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

## 🗄️ Таблицы базы данных

### 1. Таблица `users` (пользователи)

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE
);
```

**Индексы:**
```sql
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
```

---

### 2. Таблица `static_content` (статический контент)

```sql
CREATE TABLE static_content (
    id SERIAL PRIMARY KEY,
    content_key VARCHAR(100) NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    title VARCHAR(255),
    content TEXT NOT NULL,
    min_auth_state INT DEFAULT 0,
    max_auth_state INT DEFAULT 4,
    language VARCHAR(10) DEFAULT 'ru',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Уровни доступа:**
- `0` — Guest (гость)
- `4` — Authorized (авторизованный)

---

### 3. Redis (сессии)

**Формат ключей:**
- `session:<token>` — данные сессии
- `refresh:<refresh_token>` — связь refresh → access token
- `user_sessions:<user_id>` — список сессий пользователя

**TTL:**
- Access token: 15 минут
- Refresh token: 7 дней

---

## 🔐 Безопасность

### Хранение паролей

Пароли хэшируются алгоритмом **bcrypt**:

```go
// Генерация хэша
hashedPassword, _ := bcrypt.GenerateFromPassword(
    []byte(password),
    bcrypt.DefaultCost
)

// Проверка пароля
err := bcrypt.CompareHashAndPassword(
    []byte(hashedPassword),
    []byte(password)
)
```

### JWT токены

**Access token (15 минут):**
```json
{
  "user_id": 5,
  "username": "BG",
  "exp": 1709294100,
  "iat": 1709293200
}
```

**Refresh token (7 дней):**
```json
{
  "user_id": 5,
  "username": "BG",
  "exp": 1709898000,
  "iat": 1709293200,
  "type": "refresh"
}
```

### Сессии

- Сессии хранятся в Redis
- Access token действителен 15 минут
- Refresh token действителен 7 дней
- При logout сессия удаляется из Redis

---

## 🔄 Интеграция с Compliance Service

Auth Service асинхронно отправляет события в Compliance Service:

### События пользователей

#### CREATE (регистрация)
```json
{
  "event_type": "CREATE",
  "user_id": 5,
  "username": "BG",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "timestamp": "2026-03-01T12:00:00Z"
}
```

#### UPDATE (смена имени)
```json
{
  "event_type": "UPDATE",
  "user_id": 5,
  "username": "NewBG",
  "old_username": "BG",
  "new_username": "NewBG",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "timestamp": "2026-03-01T12:30:00Z"
}
```

### События сессий

#### LOGIN (вход)
```json
{
  "event_type": "LOGIN",
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123def456",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "login_time": "2026-03-01T12:00:00Z"
}
```

#### LOGOUT (выход)
```json
{
  "event_type": "LOGOUT",
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123def456",
  "client_ip": "192.168.1.100",
  "client_port": 54321
}
```

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps auth

# Логи
docker-compose logs -f auth
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8081` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `stalknet` |
| `DB_PASSWORD` | Пароль БД | `stalknet_secret` |
| `DB_NAME` | Имя БД | `stalknet` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `JWT_SECRET` | Секрет JWT токенов | `your-secret-key` |
| `COMPLIANCE_SERVICE_URL` | URL Compliance Service | `http://localhost:8086` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8081/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка подключений

```bash
# PostgreSQL
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT COUNT(*) FROM users;"

# Redis
docker exec stalknet-redis redis-cli KEYS "session:*"
```

---

## 📝 Примеры использования

### Регистрация нового пользователя

```bash
curl -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"BG","password":"password123"}'
```

### Вход

```bash
curl -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"BG","password":"password123"}'
```

### Проверка токена

```bash
curl -X POST http://localhost:8081/api/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"token":"eyJhbGciOiJIUzI1NiIs..."}'
```

### Выход

```bash
curl -X POST http://localhost:8081/api/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Проверка имени

```bash
curl -X POST http://localhost:8081/api/auth/check-username \
  -H "Content-Type: application/json" \
  -d '{"username":"BG"}'
```

### Смена имени

```bash
curl -X PUT http://localhost:8081/api/auth/update-username \
  -H "Content-Type: application/json" \
  -d '{"user_id":5,"new_username":"NewBG"}'
```

---

## 📊 SQL запросы

### Получить всех пользователей

```sql
SELECT id, username, email, status, created_at, last_seen
FROM users
ORDER BY created_at DESC;
```

### Найти пользователя по имени

```sql
SELECT id, username, email, status
FROM users
WHERE username = 'BG';
```

### Получить активных пользователей

```sql
SELECT id, username, status, last_seen
FROM users
WHERE status = 'online';
```

### Пользователи зарегистрированные за сегодня

```sql
SELECT id, username, created_at
FROM users
WHERE DATE(created_at) = CURRENT_DATE
ORDER BY created_at DESC;
```

---

## ⚠️ Важные замечания

1. **JWT_SECRET** — обязательно измените в production
2. **Пароли** — никогда не храните в открытом виде
3. **Сессии** — хранятся в Redis с TTL
4. **Compliance** — события отправляются асинхронно
5. **Срок хранения токенов:**
   - Access token: 15 минут
   - Refresh token: 7 дней

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs auth`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте Redis: `docker exec stalknet-redis redis-cli ping`
4. Проверьте health: `curl http://localhost:8081/health`

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** ✅ Работает
