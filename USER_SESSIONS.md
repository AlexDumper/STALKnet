# 📊 User Sessions - Сессии пользователей

> ⚠️ **Этот файл устарел!** Актуальная документация: **[DATABASE.md](DATABASE.md#9-user_sessions-сессии-пользователей-374-фз)**

## 📋 Описание

**User Sessions** — таблица в базе данных PostgreSQL для хранения сессий пользователей в рамках соблюдения **Федерального закона № 374-ФЗ от 06.07.2016**.

Таблица хранит информацию о:
- ✅ Входе пользователя в систему (LOGIN)
- ✅ Выходе пользователя из системы (LOGOUT)
- ✅ Разрыве соединения (DISCONNECT)
- ✅ Длительности сессии
- ✅ IP адресе и порте клиента

---

## 🎯 Назначение

### Соблюдение ФЗ-374

Согласно Федеральному закону № 374-ФЗ от 06.07.2016:

> **Организаторы распространения информации** обязаны хранить:
> - **Сессии пользователей** (вход/выход) — в течение **1 года**
> - **Метаданные** (IP, время, идентификаторы) — в течение **1 года**

**Таблица `user_sessions` обеспечивает:**
- ✅ Сохранение всех сессий пользователей
- ✅ Хранение метаданных (IP, порт, timestamp, длительность)
- ✅ Автоматическую очистку сессий старше 1 года
- ✅ Возможность предоставления данных по запросу уполномоченных органов

---

## 📊 Структура таблицы

```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL,        -- LOGIN, LOGOUT, DISCONNECT
    user_id INTEGER NOT NULL,               -- ID пользователя
    username VARCHAR(100) NOT NULL,         -- Имя пользователя
    session_id VARCHAR(255),                -- ID сессии (JWT token ID)
    client_ip VARCHAR(45) NOT NULL,         -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,           -- Порт подключения
    user_agent TEXT,                        -- User agent (опционально)
    login_time TIMESTAMP WITH TIME ZONE,    -- Время входа
    logout_time TIMESTAMP WITH TIME ZONE,   -- Время выхода
    duration_seconds INTEGER,               -- Длительность сессии в секундах
    created_at TIMESTAMP WITH TIME ZONE
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор записи |
| `event_type` | VARCHAR(20) | Тип события: `LOGIN`, `LOGOUT`, `DISCONNECT` |
| `user_id` | INTEGER | ID пользователя в системе |
| `username` | VARCHAR(100) | Имя пользователя на момент события |
| `session_id` | VARCHAR(255) | Уникальный ID сессии (JWT token ID) |
| `client_ip` | VARCHAR(45) | IP адрес клиента (IPv4 или IPv6) |
| `client_port` | INTEGER | Порт подключения клиента |
| `user_agent` | TEXT | User agent браузера/приложения (опционально) |
| `login_time` | TIMESTAMP | Время входа в систему |
| `logout_time` | TIMESTAMP | Время выхода из системы (NULL для активных сессий) |
| `duration_seconds` | INTEGER | Длительность сессии в секундах (вычисляется автоматически) |
| `created_at` | TIMESTAMP | Время создания записи |

---

## 🔧 Индексы

```sql
-- Поиск по типу события
CREATE INDEX idx_user_sessions_event_type ON user_sessions(event_type);

-- Поиск по ID пользователя
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);

-- Поиск по имени пользователя
CREATE INDEX idx_user_sessions_username ON user_sessions(username);

-- Поиск по времени входа
CREATE INDEX idx_user_sessions_login_time ON user_sessions(login_time);

-- Поиск по ID сессии
CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);

-- Поиск активных сессий (без logout_time)
CREATE INDEX idx_user_sessions_active ON user_sessions(user_id, logout_time) 
    WHERE logout_time IS NULL;
```

---

## 🔄 Типы событий

### 1. **LOGIN** — Вход пользователя

**Когда создаётся:**
- Пользователь успешно прошёл аутентификацию (`/api/auth/login`)
- Auth Service отправляет событие в Compliance Service

**Данные:**
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

---

### 2. **LOGOUT** — Выход пользователя

**Когда создаётся:**
- Пользователь явно вышел из системы (`/api/auth/logout`)
- Auth Service обновляет сессию (устанавливает `logout_time`)

**Данные:**
```json
{
  "event_type": "LOGOUT",
  "user_id": 5,
  "username": "BG",
  "session_id": "BG_abc123def456",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "login_time": "2026-03-01T12:00:00Z",
  "logout_time": "2026-03-01T13:00:00Z",
  "duration_seconds": 3600
}
```

---

### 3. **DISCONNECT** — Разрыв соединения

**Когда создаётся:**
- WebSocket соединение разорвано (клиент закрыл браузер, пропал интернет)
- Chat Service отправляет событие в Compliance Service

**Данные:**
```json
{
  "event_type": "DISCONNECT",
  "user_id": 5,
  "username": "BG",
  "client_ip": "192.168.1.100",
  "client_port": 54321,
  "login_time": "2026-03-01T12:00:00Z"
}
```

---

## 📈 Поток данных

### LOGIN (вход)

```
1. Пользователь отправляет POST /api/auth/login
   {
     "username": "BG",
     "password": "password123"
   }

2. Auth Service проверяет учётные данные

3. Auth Service генерирует JWT токены

4. Auth Service отправляет в Compliance Service:
   {
     "event_type": "LOGIN",
     "user_id": 5,
     "username": "BG",
     "session_id": "BG_abc123",
     "client_ip": "192.168.1.100",
     "client_port": 54321,
     "login_time": "2026-03-01T12:00:00Z"
   }

5. Compliance Service сохраняет в user_sessions
```

---

### LOGOUT (выход)

```
1. Пользователь отправляет POST /api/auth/logout
   Authorization: Bearer <access_token>

2. Auth Service получает сессию перед удалением

3. Auth Service удаляет сессию из Redis

4. Auth Service отправляет в Compliance Service:
   {
     "event_type": "LOGOUT",
     "user_id": 5,
     "username": "BG",
     "session_id": "BG_abc123",
     "client_ip": "192.168.1.100",
     "client_port": 54321
   }

5. Compliance Service обновляет сессию:
   - Устанавливает logout_time = NOW()
   - Вычисляет duration_seconds = logout_time - login_time
```

---

### DISCONNECT (разрыв)

```
1. Пользователь закрывает браузер / пропадает интернет

2. Chat Service обнаруживает разрыв WebSocket

3. Chat Service отправляет в Compliance Service:
   {
     "event_type": "DISCONNECT",
     "user_id": 5,
     "username": "BG",
     "client_ip": "192.168.1.100",
     "client_port": 54321
   }

4. Compliance Service сохраняет в user_sessions
```

---

## 🔍 Примеры SQL запросов

### Получить все сессии пользователя

```sql
SELECT 
    event_type,
    username,
    client_ip,
    login_time,
    logout_time,
    duration_seconds
FROM user_sessions
WHERE user_id = 5
ORDER BY login_time DESC;
```

---

### Получить активные сессии (без logout_time)

```sql
SELECT * FROM v_user_sessions_active;

-- Или напрямую:
SELECT 
    user_id,
    username,
    session_id,
    client_ip,
    login_time
FROM user_sessions
WHERE logout_time IS NULL
ORDER BY login_time DESC;
```

---

### Получить все LOGIN события за сегодня

```sql
SELECT 
    username,
    client_ip,
    login_time
FROM user_sessions
WHERE event_type = 'LOGIN'
  AND DATE(login_time) = CURRENT_DATE
ORDER BY login_time DESC;
```

---

### Статистика по сессиям за период

```sql
SELECT 
    event_type,
    COUNT(*) as count,
    DATE(MIN(login_time)) as first_event,
    DATE(MAX(login_time)) as last_event
FROM user_sessions
WHERE login_time >= NOW() - INTERVAL '30 days'
GROUP BY event_type;
```

---

### Средняя длительность сессий

```sql
SELECT 
    username,
    COUNT(*) as sessions_count,
    AVG(duration_seconds) as avg_duration_seconds,
    TO_CHAR(
        (AVG(duration_seconds) || ' seconds')::interval,
        'HH24:MI:SS'
    ) as avg_duration_formatted
FROM user_sessions
WHERE logout_time IS NOT NULL
GROUP BY username
ORDER BY sessions_count DESC;
```

---

### Найти сессии с конкретного IP

```sql
SELECT 
    username,
    event_type,
    login_time,
    logout_time
FROM user_sessions
WHERE client_ip = '192.168.1.100'
ORDER BY login_time DESC;
```

---

### Получить сессии с длительностью более 1 часа

```sql
SELECT 
    username,
    login_time,
    logout_time,
    duration_seconds,
    TO_CHAR(
        (duration_seconds || ' seconds')::interval,
        'HH24:MI:SS'
    ) as duration_formatted
FROM user_sessions
WHERE duration_seconds > 3600
ORDER BY duration_seconds DESC;
```

---

## 🗑️ Автоматическая очистка

### SQL функция

```sql
-- Удалить сессии старше 1 года
SELECT fn_cleanup_old_user_sessions();
```

**Что делает функция:**
1. Обновляет `duration_seconds` для сессий без `logout_time`
2. Удаляет сессии старше 1 года
3. Выводит количество оставшихся записей

### По расписанию (cron)

```bash
# /etc/cron.daily/compliance-cleanup
0 3 * * * docker exec stalknet-postgres psql -U stalknet -d stalknet -c "SELECT fn_cleanup_old_user_sessions();"
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
    # Не открывать порт наружу для записи!
```

---

## 📊 Мониторинг

### Проверка количества сессий

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, COUNT(*) as count FROM user_sessions GROUP BY event_type;"
```

### Просмотр последних сессий

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT event_type, username, client_ip, login_time, logout_time 
   FROM user_sessions 
   ORDER BY login_time DESC 
   LIMIT 10;"
```

### Активные сессии

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT username, session_id, client_ip, login_time 
   FROM v_user_sessions_active;"
```

---

## 📝 Примеры использования через API

### 1. Получить все сессии

```bash
curl "http://localhost:8086/api/compliance/sessions?limit=50"
```

**Ответ:**
```json
{
  "sessions": [
    {
      "id": 1,
      "event_type": "LOGIN",
      "user_id": 5,
      "username": "BG",
      "session_id": "BG_abc123",
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

### 2. Получить активные сессии

```bash
curl "http://localhost:8086/api/compliance/sessions/active"
```

**Ответ:**
```json
{
  "sessions": [...],
  "total": 5
}
```

---

### 3. Получить сессии пользователя

```bash
curl "http://localhost:8086/api/compliance/sessions/user/5"
```

**Ответ:**
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
      "login_time": "2026-03-01T12:00:00Z",
      "logout_time": "2026-03-01T13:00:00Z",
      "duration_seconds": 3600
    }
  ],
  "user_id": 5,
  "total": 1
}
```

---

### 4. Обновить сессию при LOGOUT

```bash
curl -X PUT "http://localhost:8086/api/compliance/sessions/10/logout"
```

**Ответ:**
```json
{
  "message": "Logout recorded successfully"
}
```

---

## ⚠️ Важные замечания

1. **Срок хранения** — 1 год (автоматическая очистка)
2. **Не удаляйте таблицу** — это нарушит требования ФЗ-374
3. **Регулярно делайте бэкапы** — данные должны храниться 1 год
4. **Мониторьте объем** — таблица может расти быстро
5. **duration_seconds** вычисляется автоматически при LOGOUT
6. **Активные сессии** имеют `logout_time = NULL`

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs compliance`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8086/health`

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.10  
**Статус:** ✅ Работает  
**Соответствие:** ФЗ-374 от 06.07.2016
