# 📫 Личные сообщения (Private Messages)

## 📋 Описание

Система личных сообщений позволяет отправлять сообщения конкретному пользователю, которые видны **только отправителю и получателю**.

**Офлайн-сообщения:** Если получатель офлайн, сообщение сохраняется и будет доставлено при его следующем подключении (хранение 3 суток).

---

## 🎯 Команда

```
/private <имя> <текст>
```

**Пример:**
```
/private BG Привет, как дела?
```

---

## 📡 Формат отображения

```
[11:55] [Alex] private [BG] Привет, как дела?
```

**Элементы:**
- `[11:55]` — время отправки
- `[Alex]` — имя отправителя (красный, жирный)
- `private` — метка типа сообщения (красный, курсив)
- `[BG]` — имя получателя (красный, жирный)
- `Привет, как дела?` — текст сообщения (красный)

**Цвет:** Все элементы красные (`#cf4e4e`), фон обычный.

---

## 🗄️ База данных

### Таблица chat_messages

**Новая колонка:**
```sql
contacts JSONB DEFAULT NULL
```

**Структура contacts:**
```json
[
  {"id": 5, "name": "Alex"},
  {"id": 3, "name": "BG"}
]
```

| Элемент | Описание |
|---------|----------|
| `contacts[0]` | **Отправитель** (sender) |
| `contacts[1]` | **Получатель** (recipient) |

**Пример записи:**
```sql
INSERT INTO chat_messages 
    (room_id, user_id, username, content, message_type, contacts)
VALUES 
    (1, 5, 'Alex', 'Привет!', 'private', 
     '[{"id": 5, "name": "Alex"}, {"id": 3, "name": "BG"}]');
```

### Индексы

```sql
CREATE INDEX idx_chat_messages_contacts ON chat_messages USING GIN(contacts);
CREATE INDEX idx_chat_messages_message_type ON chat_messages(message_type);
```

---

## 🔌 API Endpoints

### Auth Service (:8081)

#### GET /api/users/search?username=<query>

Поиск пользователя по имени (частичное совпадение).

**Параметры:**
- `username` (query) — имя для поиска

**Ответ:**
```json
{
  "count": 1,
  "users": [
    {
      "id": 5,
      "username": "BG",
      "email": "",
      "status": "online",
      "created_at": "2026-02-28T12:09:01Z"
    }
  ]
}
```

#### GET /api/users/:id

Информация о пользователе по ID.

**Параметры:**
- `id` (path) — ID пользователя

**Ответ:**
```json
{
  "id": 5,
  "username": "BG",
  "status": "online"
}
```

---

## 💬 WebSocket Protocol

### Клиент → Сервер

**Запрос:**
```json
{
  "type": "private_message",
  "recipient_username": "BG",
  "content": "Привет!"
}
```

### Сервер → Клиент (получатель)

**Сообщение:**
```json
{
  "type": "private_message",
  "sender_id": 5,
  "sender_username": "Alex",
  "recipient_id": 3,
  "recipient_username": "BG",
  "content": "Привет!",
  "message_type": "private",
  "contacts": [
    {"id": 5, "name": "Alex"},
    {"id": 3, "name": "BG"}
  ],
  "timestamp": "2026-03-07T11:55:00Z"
}
```

### Сервер → Клиент (отправитель)

**Подтверждение:**
```json
{
  "type": "private_message_sent",
  "recipient_username": "BG",
  "content": "Привет!",
  "timestamp": "2026-03-07T11:55:00Z"
}
```

---

## 🔄 Поток данных

```
┌─────────────┐
│ Отправитель │  /private BG Привет!
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────┐
│         Chat Service (WebSocket)                │
│  1. Проверка JWT (авторизован?)                │
│  2. Поиск BG → user_id=3 (Auth API)            │
│  3. Проверка: получатель существует            │
│  4. Создание contacts: [отправитель, получатель]│
│  5. message_type = "private"                   │
└──────────────┬──────────────────────────────────┘
               │
       ┌───────┴───────┐
       ▼               ▼
┌─────────────┐ ┌─────────────┐
│  PostgreSQL │ │  Broadcast  │
│  (БД)       │ │  (фильтр)   │
│             │ │             │
│ INSERT INTO │ │ Видят:      │
│ chat_messages│ │ - Alex (5)  │
│ contacts =  │ │ - BG (3)    │
│ [{5}, {3}]  │ │             │
└─────────────┘ └─────────────┘
```

---

## 🎨 Frontend

### Команда /private

**Обработчик:**
```javascript
function handlePrivateMessage(args) {
    // Проверка авторизации
    if (authState < AuthState.Authorized) {
        addMessage("❌ Личные сообщения доступны только авторизованным пользователям", "system");
        return;
    }

    // Проверка: не самому себе
    if (recipientUsername === username) {
        addMessage("❌ Нельзя отправить сообщение самому себе", "system");
        return;
    }

    // Отправка через WebSocket
    ws.send(JSON.stringify({
        type: "private_message",
        recipient_username: recipientUsername,
        content: messageContent
    }));
}
```

### Клик по имени

**Функциональность:**
- Клик по имени в чате → вставка `/private <имя> ` в поле ввода
- Работает для обычных и личных сообщений
- Доступно только авторизованным пользователям

**Обработчики:**
```javascript
// Обычные сообщения
document.querySelectorAll('.message:not(.private):not(.system) .username')

// Личные сообщения - отправитель
document.querySelectorAll('.message.private .sender-name')

// Личные сообщения - получатель
document.querySelectorAll('.message.private .recipient-name')
```

### CSS стили

```css
.message.private {
    padding: 8px 12px;
    margin: 4px 0;
    border-radius: 4px;
}
.message.private .sender-name,
.message.private .recipient-name,
.message.private .private-label,
.message.private .content {
    color: #cf4e4e;  /* Красный */
}
.message.private .sender-name,
.message.private .recipient-name {
    font-weight: bold;
}
.message.private .private-label {
    font-style: italic;
}
```

---

## 🔒 Безопасность

### Проверки

| Проверка | Описание |
|----------|----------|
| **Авторизация** | Только для авторизованных пользователей |
| **Существует ли получатель** | Поиск через Auth Service API |
| **Не самому себе** | Проверка: `recipientUsername !== username` |
| **Фильтрация broadcast** | Видят только участники `contacts` |
| **Сохранение в БД** | Для соблюдения ФЗ-374 |
| **Логирование** | Отправка в Compliance Service |

### Ошибки

| Ошибка | Сообщение |
|--------|-----------|
| Не авторизован | `❌ Личные сообщения доступны только авторизованным пользователям` |
| Самому себе | `❌ Нельзя отправить сообщение самому себе` |
| Пользователь не найден | `❌ User '<имя>' not found` |
| Нет подключения | `❌ Нет подключения к чату` |

---

## 📊 SQL запросы

### Офлайн-сообщения

**Таблица `private_messages_offline`:**

```sql
-- Получить все непрочитанные сообщения пользователя
SELECT 
    pm.id,
    pm.sender_username,
    pm.content,
    pm.created_at,
    pm.expires_at
FROM private_messages_offline pm
WHERE pm.recipient_id = 5 AND pm.is_read = FALSE
ORDER BY pm.created_at ASC;

-- Пометить все сообщения как прочитанные
UPDATE private_messages_offline
SET is_read = TRUE
WHERE recipient_id = 5 AND is_read = FALSE;

-- Удаление просроченных сообщений (автоматически)
SELECT fn_cleanup_old_private_offline();
```

**Срок хранения:** 3 суток (автоматическое удаление через `fn_cleanup_old_private_offline()`)

---

### Приватные сообщения в chat_messages

### Получить все личные сообщения пользователя

```sql
SELECT 
    cm.id,
    cm.username as sender,
    cm.contacts->>0->>'name' as from_name,
    cm.contacts->>1->>'name' as to_name,
    cm.content,
    cm.timestamp
FROM chat_messages cm
WHERE cm.message_type = 'private'
  AND (cm.contacts->>0->>'id')::int = 5  -- Пользователь ID=5
   OR (cm.contacts->>1->>'id')::int = 5
ORDER BY cm.timestamp DESC;
```

### Статистика по типам сообщений

```sql
SELECT 
    message_type, 
    COUNT(*) as count 
FROM chat_messages 
GROUP BY message_type;
```

### Личные сообщения между двумя пользователями

```sql
SELECT 
    cm.id,
    cm.username as sender,
    cm.content,
    cm.timestamp
FROM chat_messages cm
WHERE cm.message_type = 'private'
  AND cm.contacts->>0->>'id' = '5'  -- Отправитель
  AND cm.contacts->>1->>'id' = '3'  -- Получатель
ORDER BY cm.timestamp DESC;
```

---

## 📁 Файлы

| Файл | Описание |
|------|----------|
| `deploy/postgres/init.sql` | Миграция БД (contacts JSONB) |
| `deploy/postgres/update_help_private.sql` | Обновление справки |
| `services/auth/handlers/auth.go` | Поиск пользователей |
| `services/auth/repository/repository.go` | `SearchUsersByUsername()` |
| `services/chat/handlers/websocket.go` | Обработка `private_message` |
| `services/chat/handlers/handlers.go` | `findUserByUsername()` |
| `services/chat/repository/repository.go` | `SavePrivateMessage()` |
| `services/chat/hub/hub.go` | `BroadcastPrivate()` |
| `client/web/app.js` | Команда `/private`, клик по имени |
| `client/web/index.html` | CSS стили |
| `docker-compose.yml` | `AUTH_SERVICE_URL` для chat |

---

## 🧪 Тестирование

### Проверка команды

1. Авторизуйтесь: `/auth`
2. Отправьте личное сообщение: `/private BG Привет!`
3. Проверьте отображение: `[Alex] private [BG] Привет!`

### Проверка клика

1. Кликните по имени пользователя в чате
2. Проверьте вставку: `/private <имя> `

### Проверка видимости

1. Войдите под пользователем **Alex**
2. Отправьте: `/private BG Текст`
3. Войдите под пользователем **BG** — должен видеть сообщение
4. Войдите под пользователем **Other** — **не должен** видеть сообщение

### Проверка БД

```bash
docker exec stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT id, username, message_type, contacts, content FROM chat_messages WHERE message_type = 'private' ORDER BY timestamp DESC LIMIT 5;"
```

---

## 📚 Справка

### Обновление /help

```sql
UPDATE static_content 
SET content = '───
• /private <имя> <текст> - Личное сообщение (видно только получателю)
───'
WHERE content_key = 'help_authorized';
```

---

**Версия:** v0.1.14  
**Дата:** 2026-03-07  
**Статус:** ✅ Реализовано
