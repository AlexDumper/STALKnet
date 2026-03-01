# 👤 User Service

## 📋 Описание

**User Service** — микросервис для управления профилями пользователей и их статусами в системе STALKnet.

Сервис отвечает за:
- ✅ Получение профиля пользователя
- ✅ Обновление профиля
- ✅ Управление статусом пользователя (online/offline)
- ✅ Получение списка онлайн-пользователей
- ✅ Кэширование данных в Redis

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP GET       ┌──────────────────┐
│ Web Client  │ ─────────────────► │ User Service     │
│  (Browser)  │  /api/user/*       │ Port: 8082       │
└─────────────┘                    └────────┬─────────┘
                                            │
                         ┌──────────────────┴──────────────────┐
                         ▼                                     ▼
                  ┌──────────────┐                     ┌──────────────┐
                  │  PostgreSQL  │                     │    Redis     │
                  │   (users)    │                     │   (status)   │
                  └──────────────┘                     └──────────────┘
```

### Поток данных

#### Получение профиля:
1. Клиент отправляет GET `/api/user/profile/:id`
2. User Service получает данные из PostgreSQL
3. Возвращает профиль пользователя

#### Обновление статуса:
1. Клиент отправляет PUT `/api/user/status`
2. User Service обновляет статус в PostgreSQL
3. Обновляет кэш в Redis

---

## 🔧 API Endpoints

### 👤 Профиль пользователя

#### GET /api/user/profile/:id

**Получение профиля пользователя по ID**

**Path Parameters:**
- `id` — ID пользователя

**Пример:**
```bash
curl "http://localhost:8082/api/user/profile/5"
```

**Response (200 OK):**
```json
{
  "id": 5,
  "username": "BG",
  "email": "bg@stalknet.com",
  "status": "online",
  "created_at": "2026-03-01T10:00:00Z",
  "last_seen": "2026-03-01T12:00:00Z"
}
```

---

#### GET /api/user/profile/me

**Получение профиля текущего пользователя**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "id": 5,
  "username": "BG",
  "email": "bg@stalknet.com",
  "status": "online",
  "created_at": "2026-03-01T10:00:00Z",
  "last_seen": "2026-03-01T12:00:00Z"
}
```

---

#### PUT /api/user/profile/me

**Обновление профиля текущего пользователя**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "email": "newemail@stalknet.com"
}
```

**Response (200 OK):**
```json
{
  "message": "Profile updated successfully",
  "user_id": 5,
  "updated_fields": ["email"]
}
```

---

### 📊 Статус пользователя

#### GET /api/user/status

**Получение статуса текущего пользователя**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "user_id": 5,
  "username": "BG",
  "status": "online",
  "last_seen": "2026-03-01T12:00:00Z"
}
```

---

#### PUT /api/user/status

**Установка статуса пользователя**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "status": "online"
}
```

**Возможные статусы:**
- `online` — пользователь активен
- `offline` — пользователь не в сети
- `away` — пользователь отошёл
- `busy` — пользователь занят

**Response (200 OK):**
```json
{
  "message": "Status updated successfully",
  "user_id": 5,
  "status": "online"
}
```

---

#### GET /api/user/online

**Получение списка онлайн-пользователей**

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": 5,
      "username": "BG",
      "status": "online",
      "last_seen": "2026-03-01T12:00:00Z"
    },
    {
      "id": 7,
      "username": "Stalker",
      "status": "online",
      "last_seen": "2026-03-01T12:05:00Z"
    }
  ],
  "total": 2
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
CREATE INDEX idx_users_last_seen ON users(last_seen);
```

---

## 🔄 Интеграция с другими сервисами

### Auth Service

- Auth Service обновляет статус пользователя при входе/выходе
- Смена имени пользователя синхронизируется через Auth Service

### Chat Service

- Chat Service использует имя пользователя для отображения
- Статус пользователя отображается в чате

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps user

# Логи
docker-compose logs -f user
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8082` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `stalknet` |
| `DB_PASSWORD` | Пароль БД | `stalknet_secret` |
| `DB_NAME` | Имя БД | `stalknet` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8082/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка количества пользователей

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT COUNT(*) as total_users FROM users;"
```

### Просмотр онлайн-пользователей

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT id, username, status, last_seen 
   FROM users 
   WHERE status = 'online' 
   ORDER BY last_seen DESC;"
```

---

## 📝 Примеры использования

### Получение профиля пользователя

```bash
curl "http://localhost:8082/api/user/profile/5"
```

### Получение своего профиля

```bash
curl "http://localhost:8082/api/user/profile/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Обновление профиля

```bash
curl -X PUT "http://localhost:8082/api/user/profile/me" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"email":"newemail@stalknet.com"}'
```

### Установка статуса

```bash
curl -X PUT "http://localhost:8082/api/user/status" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"status":"busy"}'
```

### Получение онлайн-пользователей

```bash
curl "http://localhost:8082/api/user/online"
```

---

## 📊 SQL запросы

### Получить всех пользователей

```sql
SELECT id, username, email, status, created_at, last_seen
FROM users
ORDER BY created_at DESC;
```

### Получить активных пользователей

```sql
SELECT id, username, status, last_seen
FROM users
WHERE status = 'online'
ORDER BY last_seen DESC;
```

### Пользователи зарегистрированные за сегодня

```sql
SELECT id, username, created_at
FROM users
WHERE DATE(created_at) = CURRENT_DATE
ORDER BY created_at DESC;
```

### Пользователи которые давно не заходили

```sql
SELECT id, username, last_seen
FROM users
WHERE last_seen < NOW() - INTERVAL '30 days'
ORDER BY last_seen ASC;
```

### Статистика пользователей по статусам

```sql
SELECT
  status,
  COUNT(*) as count,
  MIN(last_seen) as oldest_seen,
  MAX(last_seen) as newest_seen
FROM users
GROUP BY status;
```

---

## ⚠️ Важные замечания

1. **Статусы пользователей:**
   - `online` — пользователь активен
   - `offline` — пользователь не в сети
   - `away` — пользователь отошёл
   - `busy` — пользователь занят

2. **last_seen** — обновляется при активности пользователя

3. **username** — уникальное поле, используется для идентификации

4. **Кэширование** — Redis используется для кэширования статусов

5. **🟡 Статус разработки:** сервис в разработке — базовая структура создана, требуется реализация репозитория

---

## 📈 План развития

### Реализовать в будущих версиях:

1. **Репозиторий** — реализация работы с PostgreSQL
2. **Redis кэш** — кэширование профилей и статусов
3. **События статуса** — автоматическое обновление при подключении/отключении
4. **Расширенные профили** — аватары, биография, настройки
5. **Поиск пользователей** — поиск по имени
6. **История статусов** — логирование изменений статуса

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs user`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8082/health`

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** 🟡 В РАЗРАБОТКЕ
