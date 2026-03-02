# 📦 Сохранение сообщений чата в базу данных

> ⚠️ **Этот файл устарел!** Актуальная документация: **[DATABASE.md](DATABASE.md#5-chat_messages-история-сообщений-чата-374-фз)**

## 📋 Описание

Реализовано сохранение всех сообщений пользователей из общего чата в таблицу `chat_messages` базы данных PostgreSQL.

## 🎯 Цель

- **Соблюдение ФЗ-374** - хранение сообщений в течение 1 года
- **Аудит действий** - возможность просмотра истории переписки
- **Отслеживание IP** - фиксация адреса и порта подключения

## 📊 Структура таблицы

```sql
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER NOT NULL,           -- Комната где отправлено сообщение
    user_id INTEGER NOT NULL,           -- ID пользователя
    username VARCHAR(100) NOT NULL,     -- Имя пользователя
    content TEXT NOT NULL,              -- Текст сообщения
    client_ip VARCHAR(45) NOT NULL,     -- IP адрес (IPv4/IPv6)
    client_port INTEGER NOT NULL,       -- Порт подключения
    timestamp TIMESTAMP WITH TIME ZONE, -- Время отправки
    message_type VARCHAR(20),           -- Тип: message, system, task
    created_at TIMESTAMP WITH TIME ZONE
);
```

### Индексы

```sql
idx_chat_messages_room_id          -- Поиск по комнате
idx_chat_messages_user_id          -- Поиск по пользователю
idx_chat_messages_timestamp        -- Поиск по времени
idx_chat_messages_username         -- Поиск по имени
idx_chat_messages_room_timestamp   -- Поиск по комнате + времени (DESC)
```

## 🔧 Архитектура

```
┌─────────────┐     WebSocket      ┌──────────────┐
│ Web Client  │ ◄────────────────► │  Chat Hub    │
│  Browser    │  ws://:8083/ws     │  (in-memory) │
└─────────────┘                    └──────────────┘
                                          │
                                   ┌──────┴──────┐
                                   ▼             ▼
                            ┌──────────┐  ┌──────────────┐
                            │  Room 1  │  │ ChatRepository│
                            └──────────┘  │   (PostgreSQL) │
                                          └──────────────┘
                                                   │
                                                   ▼
                                          ┌──────────────┐
                                          │ chat_messages│
                                          │   table      │
                                          └──────────────┘
```

## 📁 Изменённые файлы

### 1. База данных

**Файлы:**
- `deploy/postgres/init.sql` - добавлена таблица `chat_messages`
- `deploy/postgres/create_chat_messages.sql` - скрипт создания таблицы

### 2. Chat Service

**Файлы:**
- `services/chat/repository/repository.go` - репозиторий для работы с сообщениями
- `services/chat/handlers/handlers.go` - обновлён роутер и обработчики
- `services/chat/handlers/websocket.go` - сохранение сообщений при получении
- `services/chat/go.mod` - добавлена зависимость `github.com/lib/pq`

## 🔄 Как работает

### 1. Пользователь отправляет сообщение

```javascript
ws.send(JSON.stringify({
    type: "message",
    content: "Привет, сталкер!"
}));
```

### 2. WebSocket readPump получает сообщение

```go
func (h *ChatHandler) readPump(client *hub.Client, clientIP string, clientPort int) {
    for {
        _, message, err := client.Conn.ReadMessage()
        // ...
        
        var msg struct {
            Type    string `json:"type"`
            Content string `json:"content"`
        }
        json.Unmarshal(message, &msg)
        
        // Сохраняем в БД
        if h.repo != nil && msg.Type == "message" {
            chatMsg := &repository.ChatMessage{
                RoomID:      client.RoomID,
                UserID:      client.UserID,
                Username:    client.Username,
                Content:     msg.Content,
                ClientIP:    clientIP,
                ClientPort:  clientPort,
                MessageType: msg.Type,
                Timestamp:   time.Now(),
            }
            h.repo.SaveMessage(ctx, chatMsg)
        }
        
        // Рассылаем клиентам
        h.hub.Broadcast(...)
    }
}
```

### 3. Извлечение IP и порта

```go
func getClientIPAndPort(r *http.Request) (string, int) {
    // X-Forwarded-For для reverse proxy
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

## 📊 Примеры запросов

### Получить сообщения комнаты

```bash
GET /api/chat/rooms/1/messages?limit=50&offset=0
```

**Ответ:**
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

### SQL запросы

```sql
-- Последние 50 сообщений из комнаты 1
SELECT username, content, timestamp, client_ip
FROM chat_messages
WHERE room_id = 1
ORDER BY timestamp DESC
LIMIT 50;

-- Сообщения от пользователя
SELECT * FROM chat_messages
WHERE user_id = 5
ORDER BY timestamp DESC;

-- Сообщения за последний час
SELECT * FROM chat_messages
WHERE timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY timestamp;

-- Количество сообщений по комнатам
SELECT room_id, COUNT(*) as message_count
FROM chat_messages
GROUP BY room_id;
```

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

## 🔍 Мониторинг

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

## 📈 Производительность

### Индексы

- `idx_chat_messages_room_timestamp` - для быстрого поиска по комнате и времени
- `idx_chat_messages_user_id` - для поиска по пользователю
- `idx_chat_messages_timestamp` - для поиска по времени

### Пагинация

API поддерживает пагинацию:
- `limit` - количество сообщений (по умолчанию 50)
- `offset` - смещение (по умолчанию 0)

## 🔐 Безопасность

### Хранение данных

- **Срок хранения:** 1 год (согласно ФЗ-374)
- **Шифрование:** Не реализовано (данные в БД)
- **Доступ:** Только через Chat Service API

### Логи

- Сохранение IP и порта для аудита
- Тип сообщения (message, system, task)
- Временная метка с часовым поясом

## 🧪 Тестирование

### 1. Отправить сообщение

```bash
# Через веб-клиент или WebSocket
ws://localhost:8083/ws/chat?room_id=1&user_id=1&username=test
```

### 2. Проверить в БД

```sql
SELECT * FROM chat_messages 
WHERE username = 'test' 
ORDER BY timestamp DESC 
LIMIT 1;
```

### 3. Проверить через API

```bash
curl "http://localhost:8083/api/chat/rooms/1/messages?limit=1"
```

## 📝 Примечания

1. **Сохраняются только сообщения типа `message`** - системные и task сообщения не сохраняются
2. **IP извлекается из заголовков** - X-Forwarded-For, X-Real-IP, RemoteAddr
3. **При ошибке БД сообщение всё равно рассылается** - логирование не блокирует работу чата
4. **Таймаут записи в БД:** 5 секунд

---

**Дата реализации:** 2026-03-01  
**Версия:** v0.1.8  
**Статус:** ✅ Реализовано
