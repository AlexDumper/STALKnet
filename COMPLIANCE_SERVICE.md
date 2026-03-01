# 🛡️ Compliance Service

## 📋 Описание

**Compliance Service** — микросервис для обеспечения соблюдения **Федерального закона № 374-ФЗ от 06.07.2016** "О противодействии терроризму".

Сервис отвечает за **сохранение всех сообщений пользователей** в базе данных с указанием:
- Времени отправки (timestamp)
- Имени пользователя
- IP адреса и порта подключения
- Текста сообщения
- ID комнаты где было отправлено сообщение

---

## 🎯 Назначение

### Соблюдение ФЗ-374

Согласно Федеральному закону № 374-ФЗ от 06.07.2016:

> **Организаторы распространения информации** обязаны хранить на территории Российской Федерации:
> - **Сообщения пользователей** — в течение **1 года**
> - **Метаданные** (IP, время, идентификаторы) — в течение **1 года**

**Compliance Service обеспечивает:**
- ✅ Сохранение всех сообщений пользователей
- ✅ Хранение метаданных (IP, порт, timestamp)
- ✅ Автоматическую очистку сообщений старше 1 года
- ✅ Возможность предоставления данных по запросу уполномоченных органов

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP POST      ┌──────────────────┐
│ Chat Service│ ─────────────────► │ Compliance       │
│  (WebSocket)│  /api/compliance/  │ Service :8086    │
└─────────────┘   messages          └────────┬─────────┘
                                             │
                                             ▼
                                      ┌──────────────┐
                                      │  PostgreSQL  │
                                      │chat_messages │
                                      └──────────────┘
```

### Поток данных

1. Пользователь отправляет сообщение через WebSocket
2. Chat Service получает сообщение
3. Chat Service **асинхронно** отправляет сообщение в Compliance Service
4. Compliance Service сохраняет сообщение в PostgreSQL
5. Сообщение рассылается другим пользователям через WebSocket

---

## 📊 Таблица базы данных

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

### Индексы

```sql
idx_chat_messages_room_id         -- Поиск по комнате
idx_chat_messages_user_id         -- Поиск по пользователю
idx_chat_messages_timestamp       -- Поиск по времени
idx_chat_messages_username        -- Поиск по имени
idx_chat_messages_room_timestamp  -- Поиск по комнате + времени (DESC)
```

---

## 🔧 API Endpoints

### POST /api/compliance/messages

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

### GET /api/compliance/rooms/:id/messages

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

### GET /api/compliance/users/:id/messages

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

### DELETE /api/compliance/cleanup

**Удалить сообщения старше 1 года**

**Response:**
```json
{
  "message": "Old messages cleaned up",
  "deleted_count": 1500,
  "retention_days": 365
}
```

---

### GET /api/compliance/stats

**Получить статистику**

**Response:**
```json
{
  "total_messages": 15000,
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

### Просмотр последних сообщений

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT username, content, client_ip, timestamp 
   FROM chat_messages 
   ORDER BY timestamp DESC 
   LIMIT 10;"
```

---

## 🗑️ Автоматическая очистка

### Вручную

```bash
curl -X DELETE http://localhost:8086/api/compliance/cleanup
```

### По расписанию (cron)

```bash
# /etc/cron.daily/compliance-cleanup
0 3 * * * curl -X DELETE http://localhost:8086/api/compliance/cleanup
```

### SQL функция

```sql
-- Удалить сообщения старше 1 года
DELETE FROM chat_messages
WHERE timestamp < NOW() - INTERVAL '1 year';
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

### 1. Получить все сообщения пользователя

```bash
curl "http://localhost:8086/api/compliance/users/5/messages?limit=100"
```

### 2. Найти сообщения по IP

```sql
SELECT username, content, timestamp
FROM chat_messages
WHERE client_ip = '192.168.1.100'
ORDER BY timestamp DESC;
```

### 3. Статистика по дням

```sql
SELECT 
  DATE(timestamp) as date,
  COUNT(*) as messages_count
FROM chat_messages
GROUP BY DATE(timestamp)
ORDER BY date DESC;
```

### 4. Экспорт данных за период

```sql
COPY (
  SELECT username, content, client_ip, timestamp
  FROM chat_messages
  WHERE timestamp BETWEEN '2026-01-01' AND '2026-01-31'
  ORDER BY timestamp
) TO '/tmp/compliance_export_jan_2026.csv' WITH CSV HEADER;
```

---

## ⚠️ Важные замечания

1. **Не удаляйте Compliance Service** — это нарушит требования ФЗ-374
2. **Регулярно делайте бэкапы** — данные должны храниться 1 год
3. **Мониторьте объем** — таблица может расти быстро
4. **Настройте мониторинг** — отслеживайте ошибки сохранения

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs compliance`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8086/health`

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.8  
**Статус:** ✅ Работает  
**Соответствие:** ФЗ-374 от 06.07.2016
