# 📊 База данных STALKnet

## 📋 Описание

Этот документ описывает структуру базы данных PostgreSQL для проекта STALKnet.

**Важно:** Все изменения в схеме базы данных должны вноситься **только через файл `init.sql`**.

---

## 🗂️ Файлы

| Файл | Назначение |
|------|----------|
| [`init.sql`](init.sql) | **Единственный источник истины** для схемы БД |
| [`create_user_events.sql`](create_user_events.sql) | Миграция для таблицы user_events (устарело) |
| [`create_user_sessions.sql`](create_user_sessions.sql) | Миграция для таблицы user_sessions (устарело) |
| [`create_cleanup_function.sql`](create_cleanup_function.sql) | Миграция для функции очистки (устарело) |

---

## ⚠️ Важное правило

> **ВСЕ новые таблицы, индексы, функции и представления должны создаваться ТОЛЬКО через файл `deploy/postgres/init.sql`**

### Почему?

1. **Единый источник истины** — `init.sql` содержит полную актуальную схему БД
2. **Автоматическое применение** — при создании контейнера PostgreSQL все таблицы создаются автоматически
3. **Идемпотентность** — используется `CREATE TABLE IF NOT EXISTS`, что позволяет применять файл многократно
4. **Версионирование** — все изменения схемы хранятся в одном файле в Git
5. **Продакшн-развёртывание** — на сервере применяется тот же файл, что и локально

### Как добавлять новые таблицы?

1. Откройте `deploy/postgres/init.sql`
2. Добавьте новый `CREATE TABLE` в конец файла (перед комментариями о примерах запросов)
3. Создайте необходимые индексы
4. Добавьте комментарии к таблице и колонкам
5. Закоммитьте изменения в Git
6. Примените на продакшене:
   ```bash
   docker exec -i stalknet-postgres psql -U stalknet -d stalknet < deploy/postgres/init.sql
   ```

### Чего НЕ делать?

❌ **НЕ создавайте отдельные SQL-файлы** для новых таблиц (типа `create_something.sql`)
❌ **НЕ применяйте изменения напрямую** через psql без добавления в init.sql
❌ **НЕ редактируйте базу данных вручную** без отражения изменений в init.sql

---

## 📊 Структура базы данных

### Таблицы (9)

| № | Таблица | Описание | ФЗ-374 |
|---|---------|----------|--------|
| 1 | `users` | Пользователи системы | ❌ |
| 2 | `rooms` | Комнаты чата | ❌ |
| 3 | `room_members` | Участники комнат | ❌ |
| 4 | `messages` | Сообщения (основная) | ❌ |
| 5 | `chat_messages` | История сообщений чата | ✅ |
| 6 | `tasks` | Задачи | ❌ |
| 7 | `static_content` | Статический контент (справка) | ❌ |
| 8 | `user_events` | События пользователей | ✅ |
| 9 | `user_sessions` | Сессии пользователей | ✅ |

### Представления (1)

| Название | Описание |
|----------|----------|
| `v_user_sessions_active` | Активные сессии пользователей (без logout_time) |

### Функции (3)

| Название | Описание |
|----------|----------|
| `fn_cleanup_old_chat_messages()` | Удаление сообщений старше 1 года |
| `fn_cleanup_old_user_sessions()` | Удаление сессий старше 1 года |
| `fn_cleanup_old_compliance_data()` | **Единая функция** для очистки всех данных Compliance |

---

## 📖 Описание таблиц

### 1. users (Пользователи)

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
- `idx_users_username` — поиск по имени пользователя
- `idx_users_status` — поиск по статусу

---

### 2. rooms (Комнаты)

```sql
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by INTEGER REFERENCES users(id),
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_rooms_name` — поиск по имени комнаты

---

### 3. room_members (Участники комнат)

```sql
CREATE TABLE room_members (
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);
```

---

### 4. messages (Сообщения — оперативная таблица)

**Назначение:** Хранение последних 50 сообщений для каждой комнаты для быстрой загрузки истории при подключении пользователя.

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_messages_room_id` — поиск по комнате
- `idx_messages_user_id` — поиск по пользователю
- `idx_messages_created_at` — поиск по времени
- `idx_messages_room_created` — поиск по комнате + времени (DESC, для быстрой загрузки последних сообщений)

**Важно:**
- Таблица хранит только последние 50 сообщений на комнату
- При сохранении нового сообщения старые автоматически удаляются (SQL-запрос с LIMIT 50)
- Данные дублируются в `chat_messages` для соблюдения ФЗ-374
- Используется для быстрой загрузки истории при подключении WebSocket
- **Запись производится через `ChatRepository.SaveMessage()`** — метод сохраняет в обе таблицы одновременно (транзакция)
- **Чтение выполняется через `ChatRepository.GetRecentMessages()`** — загружает последние 50 сообщений с JOIN к таблице users

---

### 5. chat_messages (История сообщений чата) ⚖️ ФЗ-374

```sql
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    username VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    client_port INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    message_type VARCHAR(20) DEFAULT 'message',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_chat_messages_room_id` — поиск по комнате
- `idx_chat_messages_user_id` — поиск по пользователю
- `idx_chat_messages_timestamp` — поиск по времени
- `idx_chat_messages_username` — поиск по имени
- `idx_chat_messages_room_timestamp` — поиск по комнате + времени (DESC)

**Срок хранения:** 1 год (автоматическая очистка)

---

### 6. tasks (Задачи)

```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    assignee_id INTEGER REFERENCES users(id),
    room_id INTEGER REFERENCES rooms(id),
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE
);
```

**Индексы:**
- `idx_tasks_creator_id` — поиск по создателю
- `idx_tasks_assignee_id` — поиск по исполнителю
- `idx_tasks_status` — поиск по статусу
- `idx_tasks_room_id` — поиск по комнате

**Статусы:** `open`, `in_progress`, `done`, `confirmed`

---

### 7. static_content (Статический контент)

```sql
CREATE TABLE static_content (
    id SERIAL PRIMARY KEY,
    content_key VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    min_auth_state INT DEFAULT 0,
    max_auth_state INT DEFAULT 4,
    language VARCHAR(10) DEFAULT 'ru',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_static_content_key` — поиск по ключу
- `idx_static_content_auth` — поиск по уровню доступа
- `idx_static_content_language` — поиск по языку

**Уровни доступа:**
- `0` — Guest (гость)
- `4` — Authorized (авторизованный)

---

### 8. user_events (События пользователей) ⚖️ ФЗ-374

```sql
CREATE TABLE user_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- CREATE, UPDATE
    user_id INTEGER,
    username VARCHAR(100) NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    client_port INTEGER NOT NULL,
    old_username VARCHAR(100),
    new_username VARCHAR(100),
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_user_events_event_type` — поиск по типу события
- `idx_user_events_user_id` — поиск по ID пользователя
- `idx_user_events_username` — поиск по имени
- `idx_user_events_timestamp` — поиск по времени
- `idx_user_events_ip` — поиск по IP
- `idx_user_events_event_timestamp` — поиск по типу + времени (DESC)

**Срок хранения:** **Бессрочно** (НЕ очищается!)

**Типы событий:**
- `CREATE` — регистрация нового пользователя
- `UPDATE` — смена имени пользователя

---

### 9. user_sessions (Сессии пользователей) ⚖️ ФЗ-374

```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- LOGIN, LOGOUT, DISCONNECT
    user_id INTEGER NOT NULL,
    username VARCHAR(100) NOT NULL,
    session_id VARCHAR(255),
    client_ip VARCHAR(45) NOT NULL,
    client_port INTEGER NOT NULL,
    user_agent TEXT,
    login_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    logout_time TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Индексы:**
- `idx_user_sessions_event_type` — поиск по типу события
- `idx_user_sessions_user_id` — поиск по ID пользователя
- `idx_user_sessions_username` — поиск по имени
- `idx_user_sessions_login_time` — поиск по времени входа
- `idx_user_sessions_session_id` — поиск по ID сессии
- `idx_user_sessions_active` — поиск активных сессий (WHERE logout_time IS NULL)

**Срок хранения:** 1 год (автоматическая очистка)

**Типы событий:**
- `LOGIN` — вход пользователя
- `LOGOUT` — выход пользователя
- `DISCONNECT` — разрыв соединения

---

## 🔧 Функции

### fn_cleanup_old_compliance_data()

**Назначение:** Очистка старых данных Compliance (ФЗ-374)

**Что очищает:**
- ✅ `chat_messages` — старше 1 года
- ✅ `user_sessions` — старше 1 года

**Что НЕ очищает:**
- ❌ `user_events` — хранится бессрочно

**Использование:**
```sql
SELECT fn_cleanup_old_compliance_data();
```

**Автоматизация (cron):**
```bash
# Ежедневная очистка в 3:00
0 3 * * * docker exec stalknet-postgres psql -U stalknet -d stalknet -c "SELECT fn_cleanup_old_compliance_data();"
```

---

## 📝 Примеры запросов

### Пользователи

```sql
-- Все пользователи
SELECT id, username, email, status, created_at FROM users ORDER BY created_at DESC;

-- Активные пользователи
SELECT id, username, status, last_seen FROM users WHERE status = 'online';

-- Пользователь по имени
SELECT * FROM users WHERE username = 'BG';
```

### Сообщения

```sql
-- Последние 50 сообщений из комнаты (быстрая загрузка из messages)
SELECT m.username, m.content, m.timestamp
FROM messages m
JOIN users u ON m.user_id = u.id
WHERE m.room_id = 1
ORDER BY m.created_at DESC
LIMIT 50;

-- Последние 50 сообщений из комнаты (из chat_messages, для ФЗ-374)
SELECT username, content, timestamp FROM chat_messages
WHERE room_id = 1 ORDER BY timestamp DESC LIMIT 50;

-- Сообщения пользователя
SELECT * FROM chat_messages WHERE user_id = 5 ORDER BY timestamp DESC;

-- Статистика по дням
SELECT DATE(timestamp) as date, COUNT(*) as messages_count
FROM chat_messages GROUP BY DATE(timestamp) ORDER BY date DESC;
```

---

## 🔄 Поток данных при подключении WebSocket

### 1. Подключение пользователя

```
GET /ws/chat?room_id=1&user_id=5&username=BG
```

### 2. Загрузка истории

```
ChatHandler.HandleWebSocket()
  → ChatRepository.GetRecentMessages(roomID=1, limit=50)
    → SELECT из messages JOIN users
    → Реверс порядка (старые → новые)
  → Отправка клиенту (с флагом from_history: true)
```

### 3. Сохранение нового сообщения

```
ChatHandler.readPump()
  → Получение сообщения от клиента
  → ChatRepository.SaveMessage()
    → BEGIN TRANSACTION
    → INSERT INTO messages (room_id, user_id, content)
    → INSERT INTO chat_messages (room_id, user_id, username, content, client_ip, ...)
    → DELETE FROM messages (оставить последние 50)
    → COMMIT
  → Отправка в Compliance Service (для ФЗ-374)
  → Broadcast всем клиентам в комнате
```

---

### События пользователей (Compliance)

```sql
-- Все события пользователя
SELECT * FROM user_events WHERE user_id = 5 ORDER BY timestamp DESC;

-- Все регистрации за период
SELECT username, client_ip, timestamp FROM user_events
WHERE event_type = 'CREATE' AND timestamp >= NOW() - INTERVAL '30 days'
ORDER BY timestamp DESC;

-- Все смены имён
SELECT username, old_username, new_username, client_ip, timestamp
FROM user_events WHERE event_type = 'UPDATE'
ORDER BY timestamp DESC;

-- Полная история пользователя
SELECT event_type, username, old_username, new_username, client_ip, timestamp
FROM user_events WHERE user_id = 5 ORDER BY timestamp DESC;
```

### Сессии пользователей (Compliance)

```sql
-- Все сессии пользователя
SELECT event_type, username, client_ip, login_time, logout_time, duration_seconds
FROM user_sessions WHERE user_id = 5 ORDER BY login_time DESC;

-- Активные сессии
SELECT * FROM v_user_sessions_active;

-- Статистика по сессиям за сегодня
SELECT event_type, COUNT(*) as count FROM user_sessions
WHERE DATE(login_time) = CURRENT_DATE GROUP BY event_type;

-- Средняя длительность сессии
SELECT AVG(duration_seconds) as avg_duration FROM user_sessions
WHERE logout_time IS NOT NULL;
```

---

## 🚀 Развёртывание

### Локально

```bash
# При первом запуске docker-compose up -d
# init.sql применяется автоматически через docker-entrypoint-initdb.d

# Применить изменения вручную
docker exec -i stalknet-postgres psql -U stalknet -d stalknet < deploy/postgres/init.sql
```

### На продакшене

```bash
# Применить изменения
ssh root@87.242.103.13 "docker exec -i stalknet-postgres psql -U stalknet -d stalknet" < deploy/postgres/init.sql

# Или по частям (для новых таблиц)
ssh root@87.242.103.13 "docker exec -i stalknet-postgres psql -U stalknet -d stalknet" < deploy/postgres/create_user_events.sql
ssh root@87.242.103.13 "docker exec -i stalknet-postgres psql -U stalknet -d stalknet" < deploy/postgres/create_user_sessions.sql
ssh root@87.242.103.13 "docker exec -i stalknet-postgres psql -U stalknet -d stalknet" < deploy/postgres/create_cleanup_function.sql
```

---

## 📊 ER-диаграмма

```
┌─────────────┐       ┌─────────────┐
│   users     │       │   rooms     │
├─────────────┤       ├─────────────┤
│ id (PK)     │       │ id (PK)     │
│ username    │       │ name        │
│ password    │       │ description │
│ email       │       │ created_by  │◄──┐
│ status      │       │ is_private  │   │
└──────┬──────┘       └──────┬──────┘   │
       │                     │          │
       │         ┌───────────┘          │
       │         │                      │
       │         ▼                      │
       │   ┌─────────────┐              │
       │   │room_members │              │
       │   ├─────────────┤              │
       │   │ room_id (FK)│──┘          │
       │   │ user_id (FK)│─────────────┘
       │   └─────────────┘
       │
       │   ┌─────────────┐       ┌─────────────┐
       └──►│  messages   │       │   tasks     │
           ├─────────────┤       ├─────────────┤
           │ id (PK)     │       │ id (PK)     │
           │ room_id (FK)│       │ creator_id  │──┐
           │ user_id (FK)│       │ assignee_id │──┼──► users.id
           │ content     │       │ room_id (FK)│──┘
           └─────────────┘       └─────────────┘

       ┌──────────────────────────────────────────┐
       │         COMPLIANCE (ФЗ-374)              │
       ├──────────────────────────────────────────┤
       │  ┌─────────────┐    ┌─────────────┐      │
       │  │chat_messages│    │user_events  │      │
       │  ├─────────────┤    ├─────────────┤      │
       │  │ id (PK)     │    │ id (PK)     │      │
       │  │ room_id     │    │ event_type  │      │
       │  │ user_id     │    │ user_id     │──┐   │
       │  │ username    │    │ username    │  │   │
       │  │ content     │    │ client_ip   │  │   │
       │  │ client_ip   │    │ metadata    │  │   │
       │  └─────────────┘    └─────────────┘  │   │
       │                                      │   │
       │  ┌─────────────┐                     │   │
       │  │user_sessions│◄────────────────────┘   │
       │  ├─────────────┤                         │
       │  │ id (PK)     │                         │
       │  │ event_type  │                         │
       │  │ user_id     │─────────────────────────┘
       │  │ username    │
       │  │ session_id  │
       │  │ login_time  │
       │  │ logout_time │
       │  └─────────────┘
       └──────────────────────────────────────────┘
```

---

## 🔐 Безопасность

### Резервное копирование

```bash
# Бэкап базы данных
docker exec stalknet-postgres pg_dump -U stalknet stalknet > backup_$(date +%Y%m%d).sql

# Восстановление
cat backup_20260302.sql | docker exec -i stalknet-postgres psql -U stalknet -d stalknet
```

### Мониторинг

```sql
-- Размер таблиц
SELECT
    relname AS table_name,
    pg_size_pretty(pg_total_relation_size(relid)) AS total_size
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC;

-- Количество записей
SELECT
    'users' as table_name, COUNT(*) as row_count FROM users
UNION ALL SELECT 'chat_messages', COUNT(*) FROM chat_messages
UNION ALL SELECT 'user_events', COUNT(*) FROM user_events
UNION ALL SELECT 'user_sessions', COUNT(*) FROM user_sessions;
```

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи PostgreSQL:
   ```bash
   docker logs stalknet-postgres
   ```

2. Проверьте подключение:
   ```bash
   docker exec stalknet-postgres psql -U stalknet -d stalknet -c "SELECT 1;"
   ```

3. Проверьте наличие таблиц:
   ```bash
   docker exec stalknet-postgres psql -U stalknet -d stalknet -c "\dt"
   ```

4. Проверьте функции:
   ```bash
   docker exec stalknet-postgres psql -U stalknet -d stalknet -c "\df fn_cleanup*"
   ```

---

**Дата обновления:** 2026-03-02  
**Версия:** v0.1.13  
**Статус:** ✅ Актуально

---

## 📊 История изменений

### v0.1.13 (2026-03-02)

**Новые возможности:**
- ✅ Загрузка последних 50 сообщений при подключении WebSocket
- ✅ Двухтабличная архитектура хранения сообщений (messages + chat_messages)
- ✅ Автоматическая очистка старых сообщений из messages (>50)
- ✅ Быстрый SELECT из messages с JOIN к users

**Изменения:**
- Обновлён `ChatRepository.SaveMessage()` — запись в обе таблицы
- Добавлен `ChatRepository.GetRecentMessages()` — загрузка истории
- Обновлён `websocket.go` — загрузка при подключении и запись новых сообщений
