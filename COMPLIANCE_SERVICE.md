# 🛡️ Compliance Service

## 📋 Описание

**Compliance Service** — микросервис для обеспечения соблюдения **Федерального закона № 374-ФЗ от 06.07.2016** "О противодействии терроризму".

Сервис отвечает за **сохранение всех данных пользователей** в базе данных:
- ✅ Сообщения чата (текст, IP, порт, timestamp)
- ✅ События пользователей (регистрация, смена имени, IP, порт)
- ✅ Сессии пользователей (LOGIN, LOGOUT, DISCONNECT, IP, порт, длительность)

---

## 🎯 Назначение

### Соблюдение ФЗ-374

Согласно Федеральному закону № 374-ФЗ от 06.07.2016:

> **Организаторы распространения информации** обязаны хранить на территории Российской Федерации:
> - **Сообщения пользователей** — в течение **1 года**
> - **Метаданные** (IP, время, идентификаторы) — в течение **1 года**
> - **События регистрации и смены имени** — в течение **1 года**
> - **Сессии пользователей** (вход/выход) — в течение **1 года**

**Compliance Service обеспечивает:**
- ✅ Сохранение всех сообщений пользователей
- ✅ Сохранение событий пользователей (CREATE, UPDATE)
- ✅ Сохранение сессий пользователей (LOGIN, LOGOUT, DISCONNECT)
- ✅ Хранение метаданных (IP, порт, timestamp, длительность сессии)
- ✅ Автоматическую очистку данных старше 1 года (кроме user_events)
- ✅ Возможность предоставления данных по запросу уполномоченных органов

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP POST      ┌──────────────────┐
│ Chat Service│ ─────────────────► │ Compliance       │
│  (WebSocket)│  /api/compliance/  │ Service :8086    │
└─────────────┘   messages          └────────┬─────────┘
                                             │
                          ┌──────────────────┴──────────────────┐
                          ▼                                     ▼
                   ┌──────────────┐                     ┌──────────────┐
                   │  PostgreSQL  │                     │  PostgreSQL  │
                   │chat_messages │                     │ user_events  │
                   └──────────────┘                     └──────────────┘
                                             
┌─────────────┐     HTTP POST      ┌──────────────────┐
│ Auth Service│ ─────────────────► │ Compliance       │
│  (Login)    │  /api/compliance/  │ Service :8086    │
└─────────────┘   sessions          └────────┬─────────┘
                                             │
                          ┌──────────────────┴──────────────────┐
                          ▼                                     ▼
                   ┌──────────────┐                     ┌──────────────┐
                   │  PostgreSQL  │                     │  PostgreSQL  │
                   │user_sessions │                     │ user_events  │
                   └──────────────┘                     └──────────────┘
```

### Поток данных

#### Сообщения чата:
1. Пользователь отправляет сообщение через WebSocket
2. Chat Service получает сообщение
3. Chat Service **асинхронно** отправляет сообщение в Compliance Service
4. Compliance Service сохраняет сообщение в `chat_messages`
5. Сообщение рассылается другим пользователям через WebSocket

#### События пользователей:
1. Пользователь регистрируется / меняет имя
2. Auth Service создаёт / обновляет пользователя
3. Auth Service **асинхронно** отправляет событие в Compliance Service
4. Compliance Service сохраняет событие в `user_events`

#### Сессии пользователей:
1. **LOGIN**: Пользователь входит в систему → Auth Service отправляет в Compliance
2. **LOGOUT**: Пользователь выходит → Auth Service отправляет в Compliance
3. **DISCONNECT**: WebSocket разрыв → Chat Service отправляет в Compliance
4. Compliance Service сохраняет сессию в `user_sessions`

---

## 📊 Таблицы базы данных

### 1. Таблица `chat_messages` (сообщения чата)

```sql
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,           -- Комната
    user_id INTEGER NOT NULL,           -- ID пользователя
    username VARCHAR(100) NOT NULL,     -- Имя пользователя
    content TEXT NOT NULL,              -- Текст сообщения
    client_ip VARCHAR(45) NOT NULL,     -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,       -- Порт подключения
    timestamp TIMESTAMP WITH TIME ZONE, -- Время отправки
    message_type VARCHAR(20),           -- Тип: message, system, task
    created_at TIMESTAMP WITH TIME ZONE -- Время создания записи
);
```

### 2. Таблица `user_events` (события пользователей)

```sql
CREATE TABLE user_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,    -- CREATE, UPDATE
    user_id INTEGER,                    -- ID пользователя
    username VARCHAR(100) NOT NULL,     -- Имя пользователя
    client_ip VARCHAR(45) NOT NULL,     -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,       -- Порт подключения
    old_username VARCHAR(100),          -- Старое имя (для UPDATE)
    new_username VARCHAR(100),          -- Новое имя (для UPDATE)
    metadata JSONB,                     -- Дополнительные данные
    timestamp TIMESTAMP WITH TIME ZONE, -- Время события
    created_at TIMESTAMP WITH TIME ZONE -- Время создания записи
);
```

### 3. Таблица `user_sessions` (сессии пользователей)

```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,    -- LOGIN, LOGOUT, DISCONNECT
    user_id INTEGER NOT NULL,           -- ID пользователя
    username VARCHAR(100) NOT NULL,     -- Имя пользователя
    session_id VARCHAR(255),            -- ID сессии (JWT token ID)
    client_ip VARCHAR(45) NOT NULL,     -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,       -- Порт подключения
    user_agent TEXT,                    -- User agent (опционально)
    login_time TIMESTAMP WITH TIME ZONE,-- Время входа
    logout_time TIMESTAMP WITH TIME ZONE,-- Время выхода
    duration_seconds INTEGER            -- Длительность сессии в секундах
);
```

**Срок хранения:**
- `chat_messages` — 1 год (автоматическая очистка)
- `user_events` — **бессрочно** (НЕ очищается)
- `user_sessions` — 1 год (автоматическая очистка)

### Индексы

**chat_messages:**
```sql
idx_chat_messages_room_id         -- Поиск по комнате
idx_chat_messages_user_id         -- Поиск по пользователю
idx_chat_messages_timestamp       -- Поиск по времени
idx_chat_messages_username        -- Поиск по имени
idx_chat_messages_room_timestamp  -- Поиск по комнате + времени (DESC)
```

**user_events:**
```sql
idx_user_events_event_type        -- Поиск по типу события
idx_user_events_user_id           -- Поиск по ID пользователя
idx_user_events_username          -- Поиск по имени
idx_user_events_timestamp         -- Поиск по времени
idx_user_events_ip                -- Поиск по IP адресу
idx_user_events_event_timestamp   -- Поиск по типу + времени (DESC)
```

---

## 🔧 API Endpoints

### 📝 Сообщения чата

#### POST /api/compliance/messages

**Сохранить сообщение**

**Request:**
```json
{
  "room_id": 1,
  "user_id": 5,
  "username": "BG",
  "content": "Привет, сталкер!",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "message_type": "message",
  "timestamp": "2026-03-01T12:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "message": "Message saved successfully",
  "message_id": 123
}
```

---

#### GET /api/compliance/rooms/:id/messages

**Получить сообщения комнаты**

**Query Parameters:**
- `limit` (опционально) — количество сообщений (по умолчанию 50)
- `offset` (опционально) — смещение (по умолчанию 0)

**Response:**
```json
{
  "messages": [
    {
      "id": 123,
      "room_id": 1,
      "user_id": 5,
      "username": "BG",
      "content": "Привет, сталкер!",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "timestamp": "2026-03-01T12:00:00Z",
      "message_type": "message"
    }
  ],
  "room_id": 1,
  "total": 1
}
```

---

#### GET /api/compliance/users/:id/messages

**Получить сообщения пользователя**

**Query Parameters:**
- `limit` (опционально) — количество сообщений (по умолчанию 50)

**Response:**
```json
{
  "messages": [...],
  "user_id": 5,
  "total": 10
}
```

---

### 👤 События пользователей

#### POST /api/compliance/user-events

**Сохранить событие пользователя**

**Типы событий:**
- `CREATE` — регистрация нового пользователя
- `UPDATE` — смена имени пользователя

**Request (CREATE):**
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

**Request (UPDATE):**
```json
{
  "event_type": "UPDATE",
  "user_id": 5,
  "username": "NewBG",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "old_username": "BG",
  "new_username": "NewBG",
  "timestamp": "2026-03-01T12:30:00Z"
}
```

**Response (201 Created):**
```json
{
  "message": "User event saved successfully",
  "event_id": 1
}
```

---

### 🔐 Сессии пользователей

#### POST /api/compliance/sessions

**Сохранить сессию пользователя (LOGIN)**

**Request:**
```json
{
  "event_type": "LOGIN",
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123def456",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "user_agent": "Mozilla/5.0...",
  "login_time": "2026-03-01T12:00:00Z"
}
```

**Response (201 Created):**
```json
{
  "message": "Session saved successfully",
  "session_id": 1
}
```

**Примечание:** Для LOGOUT и DISCONNECT используется тот же endpoint с `event_type: "LOGOUT"` или `event_type: "DISCONNECT"`.

---

#### GET /api/compliance/sessions

**Получить все сессии**

**Query Parameters:**
- `event_type` (опционально) — фильтр по типу (LOGIN, LOGOUT, DISCONNECT)
- `limit` (опционально) — количество (по умолчанию 50)
- `offset` (опционально) — смещение

**Request:**
```bash
curl "http://localhost:8086/api/compliance/sessions?event_type=LOGIN&limit=10"
```

**Response:**
```json
{
  "sessions": [
    {
      "id": 1,
      "event_type": "LOGIN",
      "user_id": 5,
      "username": "BG",
      "session_id": "BG_abc123def456",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "login_time": "2026-03-01T12:00:00Z",
      "logout_time": null,
      "duration_seconds": null
    }
  ],
  "total": 1
}
```

---

#### GET /api/compliance/sessions/active

**Получить активные сессии** (без logout_time)

**Response:**
```json
{
  "sessions": [...],
  "total": 5
}
```

---

#### GET /api/compliance/sessions/user/:userId

**Получить сессии конкретного пользователя**

**Request:**
```bash
curl "http://localhost:8086/api/compliance/sessions/user/5"
```

**Response:**
```json
{
  "sessions": [
    {
      "id": 10,
      "event_type": "LOGIN",
      "user_id": 5,
      "username": "BG",
      "session_id": "BG_abc123",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "login_time": "2026-03-01T14:00:00Z",
      "logout_time": "2026-03-01T15:00:00Z",
      "duration_seconds": 3600
    }
  ],
  "user_id": 5,
  "total": 1
}
```

---

#### PUT /api/compliance/sessions/:id/logout

**Обновить сессию при LOGOUT** (устанавливает logout_time и duration_seconds)

**Request:**
```bash
curl -X PUT "http://localhost:8086/api/compliance/sessions/10/logout"
```

**Response:**
```json
{
  "message": "Logout recorded successfully"
}
```

---

#### GET /api/compliance/user-events

**Получить все события пользователей**

**Query Parameters:**
- `event_type` (опционально) — фильтр по типу (CREATE или UPDATE)
- `limit` (опционально) — количество (по умолчанию 50)
- `offset` (опционально) — смещение

**Request:**
```bash
curl "http://localhost:8086/api/compliance/user-events?event_type=CREATE&limit=10"
```

**Response:**
```json
{
  "events": [
    {
      "id": 1,
      "event_type": "CREATE",
      "user_id": 5,
      "username": "BG",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "timestamp": "2026-03-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

---

#### GET /api/compliance/user-events/:username

**Получить события по имени пользователя**

**Request:**
```bash
curl "http://localhost:8086/api/compliance/user-events/BG"
```

**Response:**
```json
{
  "events": [
    {
      "id": 2,
      "event_type": "UPDATE",
      "user_id": 5,
      "username": "NewBG",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "old_username": "BG",
      "new_username": "NewBG",
      "timestamp": "2026-03-01T12:30:00Z"
    },
    {
      "id": 1,
      "event_type": "CREATE",
      "user_id": 5,
      "username": "BG",
      "client_ip": "192.168.1.100",
      "client_port": 54321,
      "timestamp": "2026-03-01T12:00:00Z"
    }
  ],
  "username": "BG",
  "total": 2
}
```

---

## 🗑️ Обслуживание

### Очистка данных

**Важно:** Данные в таблицах обрабатываются по-разному:

| Таблица | Срок хранения | Очистка |
|---------|--------------|---------|
| `chat_messages` | 1 год | ✅ Автоматическая очистка |
| `user_events` | **Бессрочно** | ❌ **НЕ очищается** |

**Обоснование:**
- **chat_messages** — сообщения чата хранятся 1 год согласно ФЗ-374
- **user_events** — сведения о регистрации и смене имён накапливаются **бессрочно** для идентификации пользователей и предоставления уполномоченным органам

---

#### DELETE /api/compliance/cleanup

**Удалить сообщения чата старше 1 года**

⚠️ **Примечание:** `user_events` НЕ очищается!

**Response:**
```json
{
  "message": "Old chat messages cleaned up (user_events preserved)",
  "deleted_count": 1500,
  "retention_days": 365,
  "user_events": "preserved indefinitely"
}
```

---

### 📊 Статистика

#### GET /api/compliance/stats

**Получить статистику**

**Response:**
```json
{
  "total_messages": 15000,
  "total_user_events": 500,
  "retention_days": 365,
  "compliance": "ФЗ-374 от 06.07.2016"
}
```

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps compliance

# Логи
docker-compose logs -f compliance
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8086` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `stalknet` |
| `DB_PASSWORD` | Пароль БД | `stalknet_secret` |
| `DB_NAME` | Имя БД | `stalknet` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8086/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка количества сообщений

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT COUNT(*) as total_messages FROM chat_messages;"
```

### Проверка количества событий пользователей

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, COUNT(*) as count FROM user_events GROUP BY event_type;"
```

### Просмотр последних сообщений

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT username, content, client_ip, timestamp
   FROM chat_messages
   ORDER BY timestamp DESC
   LIMIT 10;"
```

### Просмотр последних событий пользователей

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, username, client_ip, timestamp
   FROM user_events
   ORDER BY timestamp DESC
   LIMIT 10;"
```

---

## 🗑️ Автоматическая очистка

### Вручную

```bash
# Очищает ТОЛЬКО chat_messages (user_events сохраняется!)
curl -X DELETE http://localhost:8086/api/compliance/cleanup
```

### По расписанию (cron)

```bash
# /etc/cron.daily/compliance-cleanup
# Очищает ТОЛЬКО сообщения чата старше 1 года
# user_events НЕ очищается - данные накапливаются бессрочно!
0 3 * * * curl -X DELETE http://localhost:8086/api/compliance/cleanup
```

### SQL

```sql
-- ⚠️ Очищает ТОЛЬКО chat_messages (user_events НЕ трогать!)
DELETE FROM chat_messages
WHERE timestamp < NOW() - INTERVAL '1 year';

-- ❌ НЕ ВЫПОЛНЯТЬ для user_events!
-- Данные о регистрации и смене имён должны накапливаться бессрочно!
```

---

## 🔐 Безопасность

### Требования к доступу

1. **Внутренняя сеть** — доступ только из внутренней сети
2. **Аутентификация** — требуется JWT токен для API
3. **Логирование** — все запросы логируются

### Ограничение доступа

```yaml
# docker-compose.yml
services:
  compliance:
    networks:
      - stalknet-network  # Только внутренняя сеть
    # Не открывать порт наружу!
```

---

## 📈 Производительность

### Рекомендации

1. **Асинхронная отправка** — Chat Service отправляет сообщения асинхронно (не блокирует WebSocket)
2. **Пул подключений** — PostgreSQL настроен на 25 одновременных подключений
3. **Индексы** — все необходимые индексы созданы
4. **Периодическая очистка** — удаляйте сообщения старше 1 года

### Метрики

- **RPS** — запросов в секунду (отправка сообщений)
- **Latency** — время сохранения сообщения (< 100ms)
- **Storage** — объем занимаемых данных (~1GB на 1M сообщений)

---

## 📝 Примеры использования

### Сообщения чата

#### 1. Получить все сообщения пользователя

```bash
curl "http://localhost:8086/api/compliance/users/5/messages?limit=100"
```

#### 2. Найти сообщения по IP

```sql
SELECT username, content, timestamp
FROM chat_messages
WHERE client_ip = '192.168.1.100'
ORDER BY timestamp DESC;
```

#### 3. Статистика по дням

```sql
SELECT
  DATE(timestamp) as date,
  COUNT(*) as messages_count
FROM chat_messages
GROUP BY DATE(timestamp)
ORDER BY date DESC;
```

#### 4. Экспорт данных за период

```sql
COPY (
  SELECT username, content, client_ip, timestamp
  FROM chat_messages
  WHERE timestamp BETWEEN '2026-01-01' AND '2026-01-31'
  ORDER BY timestamp
) TO '/tmp/compliance_export_jan_2026.csv' WITH CSV HEADER;
```

---

### События пользователей

#### 1. Получить все регистрации за последний месяц

```sql
SELECT username, client_ip, timestamp
FROM user_events
WHERE event_type = 'CREATE'
  AND timestamp >= NOW() - INTERVAL '30 days'
ORDER BY timestamp DESC;
```

#### 2. Получить все смены имён пользователя

```sql
SELECT old_username, new_username, client_ip, timestamp
FROM user_events
WHERE event_type = 'UPDATE'
  AND username = 'BG'
ORDER BY timestamp DESC;
```

#### 3. Получить полную историю пользователя

```sql
SELECT
  event_type,
  username,
  old_username,
  new_username,
  client_ip,
  timestamp
FROM user_events
WHERE user_id = 5
ORDER BY timestamp DESC;
```

#### 4. Найти все события с IP адреса

```sql
SELECT username, event_type, timestamp
FROM user_events
WHERE client_ip = '192.168.1.100'
ORDER BY timestamp DESC;
```

#### 5. Статистика событий по типам

```sql
SELECT
  event_type,
  COUNT(*) as count,
  DATE(MIN(timestamp)) as first_event,
  DATE(MAX(timestamp)) as last_event
FROM user_events
GROUP BY event_type;
```

---

## ⚠️ Важные замечания

1. **Не удаляйте Compliance Service** — это нарушит требования ФЗ-374
2. **Регулярно делайте бэкапы** — данные должны храниться согласно политике
3. **Мониторьте объем** — таблицы могут расти быстро
4. **Настройте мониторинг** — отслеживайте ошибки сохранения
5. **Два типа данных:**
   - `chat_messages` — сообщения чата (хранение 1 год, автоматическая очистка)
   - `user_events` — события пользователей (хранение **бессрочно**, НЕ очищается)
6. **Асинхронная отправка** — события отправляются асинхронно и не блокируют работу других сервисов
7. **⚠️ user_events НЕ подлежит очистке!** — данные о регистрации и смене имён накапливаются бессрочно

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs compliance`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8086/health`

---

**Дата обновления:** 2026-03-01
**Версия:** v0.1.9
**Статус:** ✅ Работает
**Соответствие:** ФЗ-374 от 06.07.2016
