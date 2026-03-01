# 📊 User Events - События пользователей

## 📋 Описание

Compliance Service теперь сохраняет не только сообщения чата, но и **события пользователей** для соблюдения требований **Федерального закона № 374-ФЗ от 06.07.2016**.

## 🎯 Типы событий

| Событие | Код | Описание |
|---------|-----|----------|
| **Create** | `CREATE` | Регистрация нового пользователя |
| **Update** | `UPDATE` | Изменение имени пользователя |

---

## 📊 Структура таблицы

```sql
CREATE TABLE user_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- CREATE, UPDATE
    user_id INTEGER,                        -- ID пользователя
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,           -- Порт подключения
    old_username VARCHAR(100),              -- Старое имя (для UPDATE)
    new_username VARCHAR(100),              -- Новое имя (для UPDATE)
    metadata JSONB,                         -- Дополнительные данные
    timestamp TIMESTAMP WITH TIME ZONE,     -- Время события
    created_at TIMESTAMP WITH TIME ZONE     -- Время создания записи
);
```

### Индексы

```sql
idx_user_events_event_type       -- Поиск по типу события
idx_user_events_user_id          -- Поиск по ID пользователя
idx_user_events_username         -- Поиск по имени
idx_user_events_timestamp        -- Поиск по времени
idx_user_events_ip               -- Поиск по IP адресу
idx_user_events_event_timestamp  -- Поиск по типу + времени
```

---

## 🔧 API Endpoints

### POST /api/compliance/user-events

**Сохранить событие пользователя**

**Request:**
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

**Для UPDATE:**
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

### GET /api/compliance/user-events

**Получить все события пользователей**

**Query Parameters:**
- `event_type` (опционально) — фильтр по типу (CREATE или UPDATE)
- `limit` (опционально) — количество (по умолчанию 50)
- `offset` (опционально) — смещение

**Пример:**
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

### GET /api/compliance/user-events/:username

**Получить события по имени пользователя**

**Пример:**
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

## 🔄 Поток данных

### Регистрация пользователя (CREATE)

```
1. Пользователь отправляет POST /api/auth/register
   {
     "username": "BG",
     "password": "password123"
   }

2. Auth Service создаёт пользователя в БД

3. Auth Service отправляет в Compliance Service:
   {
     "event_type": "CREATE",
     "user_id": 5,
     "username": "BG",
     "client_ip": "192.168.1.100",
     "client_port": 54321,
     "timestamp": "2026-03-01T12:00:00Z"
   }

4. Compliance Service сохраняет в user_events
```

### Смена имени (UPDATE)

```
1. Пользователь отправляет PUT /api/auth/update-username
   {
     "user_id": 5,
     "new_username": "NewBG"
   }

2. Auth Service проверяет доступность имени

3. Auth Service обновляет имя в БД

4. Auth Service отправляет в Compliance Service:
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

5. Compliance Service сохраняет в user_events
```

---

## 📝 Примеры SQL запросов

### Получить все регистрации за последний месяц

```sql
SELECT username, client_ip, timestamp
FROM user_events
WHERE event_type = 'CREATE'
  AND timestamp >= NOW() - INTERVAL '30 days'
ORDER BY timestamp DESC;
```

### Получить все смены имён пользователя

```sql
SELECT old_username, new_username, client_ip, timestamp
FROM user_events
WHERE event_type = 'UPDATE'
  AND username = 'BG'
ORDER BY timestamp DESC;
```

### Получить полную историю пользователя

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

### Статистика по событиям

```sql
SELECT 
  event_type,
  COUNT(*) as count,
  DATE(MIN(timestamp)) as first_event,
  DATE(MAX(timestamp)) as last_event
FROM user_events
GROUP BY event_type;
```

### Найти все события с IP адреса

```sql
SELECT username, event_type, timestamp
FROM user_events
WHERE client_ip = '192.168.1.100'
ORDER BY timestamp DESC;
```

---

## 🔍 Мониторинг

### Проверка количества событий

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, COUNT(*) as count FROM user_events GROUP BY event_type;"
```

### Просмотр последних событий

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, username, client_ip, timestamp FROM user_events ORDER BY timestamp DESC LIMIT 10;"
```

---

## ⚠️ Важные замечания

1. **Асинхронная отправка** — Auth Service отправляет события асинхронно (не блокирует ответ пользователю)
2. **Гарантия доставки** — при ошибке отправки событие логируется но не прерывает регистрацию
3. **Срок хранения** — 1 год согласно ФЗ-374
4. **IP адрес** — определяется из X-Forwarded-For, X-Real-IP или RemoteAddr

---

## 📊 Интеграция с другими сервисами

### Chat Service
- Сохраняет сообщения в `chat_messages`
- Содержит IP и порт отправителя

### Auth Service
- Сохраняет события регистрации (`CREATE`)
- Сохраняет события смены имени (`UPDATE`)
- Содержит IP и порт клиента

### Compliance Service
- Единый сервис для всех данных ФЗ-374
- API для получения статистики и истории
- Автоматическая очистка данных старше 1 года

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.9  
**Статус:** ✅ Работает  
**Соответствие:** ФЗ-374 от 06.07.2016
