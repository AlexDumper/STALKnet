# 💬 Chat Service

## 📋 Описание

**Chat Service** — микросервис для real-time обмена сообщениями между пользователями через WebSocket.

Сервис отвечает за:
- ✅ WebSocket соединения с клиентами
- ✅ Трансляцию сообщений в комнаты
- ✅ Управление комнатами чата
- ✅ Сохранение сообщений в базу данных (через Compliance Service)
- ✅ Отслеживание участников комнат
- ✅ Системные уведомления о событиях

---

## 🏗️ Архитектура

```
┌─────────────┐    WebSocket     ┌──────────────────┐
│ Web Client  │ ◄──────────────► │ Chat Service     │
│  (Browser)  │ ws://:8083/ws    │ Port: 8083       │
└─────────────┘                  └────────┬─────────┘
                                          │
                                   ┌──────┴──────┐
                                   ▼             ▼
                            ┌──────────┐  ┌──────────────┐
                            │  Hub     │  │ ChatRepository│
                            │(in-memory)│ │   (PostgreSQL) │
                            └──────────┘  └──────────────┘
                                   │               │
                                   ▼               ▼
                            ┌──────────┐  ┌──────────────┐
                            │  Room 1  │  │chat_messages │
                            │ general  │  │   table      │
                            └──────────┘  └──────────────┘
                                          │
                                          ▼
                                   ┌──────────────────┐
                                   │ Compliance       │
                                   │ Service :8086    │
                                   └──────────────────┘
```

### Компоненты

#### Hub (Концентратор)
- Управляет WebSocket соединениями
- Хранит активных клиентов в памяти
- Рассылает сообщения по комнатам

#### Room (Комната)
- Логическая группировка клиентов
- Каждый клиент подключён к одной комнате
- Сообщения транслируются только в пределах комнаты

#### Repository (Репозиторий)
- Сохранение сообщений в PostgreSQL
- Получение истории сообщений

---

## 🔧 API Endpoints

### 💬 WebSocket

#### GET /ws/chat

**WebSocket подключение к чату**

**Query Parameters:**
- `room_id` — ID комнаты для подключения
- `user_id` — ID пользователя
- `username` — Имя пользователя для отображения

**Пример подключения:**
```javascript
const ws = new WebSocket(
  `ws://localhost:8083/ws/chat?room_id=1&user_id=5&username=BG`
);
```

**Формат сообщений (client → server):**
```json
{
  "type": "message",
  "content": "Привет, сталкер!"
}
```

**Формат сообщений (server → client):**
```json
{
  "type": "message",
  "content": "Привет, сталкер!",
  "username": "BG",
  "user_id": 5,
  "room_id": 1,
  "timestamp": "2026-03-01T12:00:00Z"
}
```

**Типы сообщений:**
- `message` — обычное сообщение пользователя
- `system` — системное сообщение (вход/выход)
- `task` — уведомление о задаче

---

### 📚 Комнаты

#### GET /api/chat/rooms

**Получение списка комнат**

**Response (200 OK):**
```json
{
  "rooms": [
    {
      "id": 1,
      "name": "general",
      "description": "Общая комната"
    },
    {
      "id": 2,
      "name": "tasks",
      "description": "Задачи"
    }
  ]
}
```

---

#### POST /api/chat/rooms

**Создание новой комнаты**

**Request:**
```json
{
  "name": "new-room",
  "description": "Описание комнаты",
  "is_private": false
}
```

**Response (201 Created):**
```json
{
  "message": "Room created",
  "name": "new-room"
}
```

---

#### GET /api/chat/rooms/:id/messages

**Получение сообщений комнаты**

**Query Parameters:**
- `limit` (опционально) — количество сообщений (по умолчанию 50)
- `offset` (опционально) — смещение (по умолчанию 0)

**Пример:**
```bash
curl "http://localhost:8083/api/chat/rooms/1/messages?limit=50&offset=0"
```

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": 1,
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

#### POST /api/chat/rooms/:id/messages

**Отправка сообщения в комнату**

**Request:**
```json
{
  "content": "Привет, сталкер!"
}
```

**Response (200 OK):**
```json
{
  "message": "Message sent",
  "content": "Привет, сталкер!"
}
```

---

#### GET /api/chat/rooms/:id/members

**Получение участников комнаты**

**Response (200 OK):**
```json
{
  "members": [
    {
      "user_id": 5,
      "username": "BG"
    }
  ],
  "room_id": 1
}
```

---

#### POST /api/chat/rooms/:id/join

**Присоединение к комнате**

**Response (200 OK):**
```json
{
  "message": "Joined room",
  "room_id": 1
}
```

---

#### POST /api/chat/rooms/:id/leave

**Покидание комнаты**

**Response (200 OK):**
```json
{
  "message": "Left room",
  "room_id": 1
}
```

---

## 🗄️ Таблицы базы данных

### 1. Таблица `chat_messages` (сообщения чата)

```sql
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    username VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    client_ip VARCHAR(45) NOT NULL,
    client_port INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE,
    message_type VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE
);
```

**Индексы:**
```sql
CREATE INDEX idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_chat_messages_timestamp ON chat_messages(timestamp);
CREATE INDEX idx_chat_messages_username ON chat_messages(username);
CREATE INDEX idx_chat_messages_room_timestamp 
  ON chat_messages(room_id, timestamp DESC);
```

---

### 2. Таблица `rooms` (комнаты)

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

---

### 3. Таблица `room_members` (участники комнат)

```sql
CREATE TABLE room_members (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id),
    user_id INTEGER REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_id, user_id)
);
```

---

## 🔄 Как работает

### 1. Подключение клиента

```
1. Клиент отправляет WebSocket запрос:
   GET /ws/chat?room_id=1&user_id=5&username=BG

2. Chat Service создаёт клиента:
   client := &hub.Client{
       Hub:      h.hub,
       Conn:     conn,
       UserID:   5,
       Username: "BG",
       RoomID:   1,
       Send:     make(chan []byte, 256),
   }

3. Клиент регистрируется в Hub:
   client.Hub.Register <- client

4. Hub добавляет клиента в комнату:
   hub.rooms[1][client] = true

5. Отправляется системное сообщение:
   "BG присоединился к чату"
```

---

### 2. Отправка сообщения

```
1. Клиент отправляет JSON:
   {
     "type": "message",
     "content": "Привет, сталкер!"
   }

2. readPump получает сообщение:
   json.Unmarshal(message, &msg)

3. Сообщение сохраняется в Compliance Service:
   POST http://compliance:8086/api/compliance/messages
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

4. Сообщение рассылается клиентам в комнате:
   h.hub.Broadcast(1, 5, "BG", "Привет, сталкер!", "message")
```

---

### 3. Отключение клиента

```
1. WebSocket соединение разорвано

2. readPump обнаруживает ошибку:
   _, message, err := client.Conn.ReadMessage()
   if err != nil { break }

3. Отправляется событие DISCONNECT:
   POST http://compliance:8086/api/compliance/sessions
   {
     "event_type": "DISCONNECT",
     "user_id": 5,
     "username": "BG",
     "client_ip": "192.168.1.100",
     "client_port": 54321
   }

4. Клиент удаляется из Hub:
   client.Hub.Unregister <- client

5. Отправляется системное сообщение:
   "BG покинул чат"
```

---

## 🔐 Безопасность

### Извлечение IP адреса

```go
func getClientIPAndPort(r *http.Request) (string, int) {
    // X-Forwarded-For (для reverse proxy)
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0]), 0
    }

    // X-Real-IP
    xri := r.Header.Get("X-Real-IP")
    if xri != "" {
        return xri, 0
    }

    // RemoteAddr
    host, portStr, _ := net.SplitHostPort(r.RemoteAddr)
    port, _ := strconv.Atoi(portStr)
    return host, port
}
```

### Требования к данным

- **Username:** обязателен для подключения
- **Room ID:** должен быть валидным integer
- **User ID:** должен быть валидным integer

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps chat

# Логи
docker-compose logs -f chat
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8083` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `stalknet` |
| `DB_PASSWORD` | Пароль БД | `stalknet_secret` |
| `DB_NAME` | Имя БД | `stalknet` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `COMPLIANCE_SERVICE_URL` | URL Compliance Service | `http://localhost:8086` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8083/health
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
  "SELECT username, content, timestamp, client_ip
   FROM chat_messages
   ORDER BY timestamp DESC
   LIMIT 10;"
```

---

## 📝 Примеры использования

### WebSocket подключение (JavaScript)

```javascript
const ws = new WebSocket(
  `ws://localhost:8083/ws/chat?room_id=1&user_id=5&username=BG`
);

ws.onopen = () => {
  console.log('Connected to chat');
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(`${msg.username}: ${msg.content}`);
};

ws.send(JSON.stringify({
  type: "message",
  content: "Привет, сталкер!"
}));
```

### Получение истории сообщений

```bash
curl "http://localhost:8083/api/chat/rooms/1/messages?limit=50&offset=0"
```

### Отправка сообщения через API

```bash
curl -X POST http://localhost:8083/api/chat/rooms/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Привет, сталкер!"}'
```

---

## 📊 SQL запросы

### Последние 50 сообщений из комнаты 1

```sql
SELECT username, content, timestamp, client_ip
FROM chat_messages
WHERE room_id = 1
ORDER BY timestamp DESC
LIMIT 50;
```

### Сообщения от пользователя

```sql
SELECT * FROM chat_messages
WHERE user_id = 5
ORDER BY timestamp DESC;
```

### Сообщения за последний час

```sql
SELECT * FROM chat_messages
WHERE timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY timestamp;
```

### Количество сообщений по комнатам

```sql
SELECT room_id, COUNT(*) as message_count
FROM chat_messages
GROUP BY room_id;
```

### Статистика по дням

```sql
SELECT
  DATE(timestamp) as date,
  COUNT(*) as messages_count
FROM chat_messages
GROUP BY DATE(timestamp)
ORDER BY date DESC;
```

---

## 🗑️ Очистка старых сообщений

### Автоматическая функция

```sql
-- Удалить сообщения старше 1 года
SELECT fn_cleanup_old_chat_messages();
```

### Вручную

```sql
DELETE FROM chat_messages
WHERE timestamp < NOW() - INTERVAL '1 year';
```

### Через Compliance API

```bash
curl -X DELETE http://localhost:8086/api/compliance/cleanup
```

**Ответ:**
```json
{
  "message": "Old chat messages cleaned up",
  "deleted_count": 1500,
  "retention_days": 365
}
```

---

## ⚠️ Важные замечания

1. **Сохраняются только сообщения типа `message`** — системные и task сообщения не сохраняются
2. **IP извлекается из заголовков** — X-Forwarded-For, X-Real-IP, RemoteAddr
3. **При ошибке БД сообщение всё равно рассылается** — логирование не блокирует работу чата
4. **Таймаут записи в БД:** 5 секунд
5. **Срок хранения сообщений:** 1 год (автоматическая очистка)
6. **Hub хранится в памяти** — при перезапуске сервиса все соединения сбрасываются

---

## 📈 Производительность

### Рекомендации

1. **Асинхронная отправка** — сообщения отправляются в Compliance асинхронно
2. **Пул подключений** — PostgreSQL настроен на 25 одновременных подключений
3. **Индексы** — все необходимые индексы созданы
4. **Периодическая очистка** — удаляйте сообщения старше 1 года

### Метрики

- **RPS** — запросов в секунду (отправка сообщений)
- **Latency** — время сохранения сообщения (< 100ms)
- **Storage** — объем занимаемых данных (~1GB на 1M сообщений)
- **Connections** — количество одновременных WebSocket соединений

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs chat`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8083/health`
4. Проверьте WebSocket: используйте инструмент типа wscat

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** ✅ Работает
**Соответствие:** ФЗ-374 от 06.07.2016
