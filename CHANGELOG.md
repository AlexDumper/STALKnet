# 📝 Журнал изменений STALKnet

## [v0.1.17] - 2026-03-07

### 🐛 Исправление поиска пользователей для приватных сообщений

#### Проблема
При отправке приватных сообщений (`/private <имя> <текст>`) Chat Service не мог найти пользователей и возвращал ошибку:
```
❌ User 'username' not found
```

#### Причина
Переменная окружения `AUTH_SERVICE_URL` не была настроена в `docker-compose.prod.yml` для Chat Service. 
Код использовал значение по умолчанию `http://localhost:8081`, но внутри контейнера `localhost` — это сам контейнер чата, 
а не хост-машина или другие контейнеры.

#### Решение
Добавлена переменная окружения `AUTH_SERVICE_URL=http://auth:8081` в конфигурацию Chat Service.

**Изменённые файлы:**
- `docker-compose.prod.yml` — добавлена переменная `AUTH_SERVICE_URL` для сервиса `chat`
- `scripts/update-prod.sh` — обновлён скрипт для автоматического удаления старых контейнеров перед перезапуском

**Файлы:**
- `services/chat/handlers/handlers.go` — `NewChatHandler()` использует `os.Getenv("AUTH_SERVICE_URL")`
- `services/chat/main.go` — передача `authURL` в `SetupRouter()` (требуется обновление)

---

### 🚀 GitHub CI/CD

#### Новые файлы для автоматического развёртывания
- `.github/workflows/deploy.yml` — workflow для автоматического деплоя при пуше в `main`
- `.github/ENV_SETUP.md` — инструкция по настройке GitHub Secrets
- `scripts/update-prod.sh` — bash-скрипт для обновления на сервере
- `scripts/update-prod-remote.ps1` — PowerShell-скрипт для удалённого обновления
- `GITHUB_DEPLOY.md` — полное руководство по развёртыванию через GitHub

#### Обновлённые файлы
- `README.md` — добавлен раздел о GitHub CI/CD
- `docker-compose.prod.yml` — добавлена переменная `AUTH_SERVICE_URL` для Chat Service

---

## [v0.1.16] - 2026-03-07

### 🔐 Увеличение срока действия сессии

#### Access токен
- **Было:** 15 минут
- **Стало:** 3 дня (72 часа)

#### Сессия в Redis
- **Было:** 15 минут
- **Стало:** 3 дня (72 часа)

#### Refresh токен
- **Осталось:** 7 дней

**Причина:** Более стабильная работа автологина без частых обновлений токена.

**Файлы:**
- `services/auth/handlers/auth.go` — `generateTokens()`, `Login()`, `Refresh()`

---

### 🎨 UI улучшения

#### Иконка офлайн-сообщений
- **Было:** 📬 (почтовый ящик)
- **Стало:** ✉️ (конверт)

**Файлы:**
- `client/web/app.js` — `loadOfflinePrivateMessages()`

---

## [v0.1.15] - 2026-03-07

### 📬 Офлайн-приватные сообщения

#### Архитектура
Реализована система сохранения приватных сообщений для офлайн-получателей с автоматическим удалением через 3 суток.

**Поток данных:**
1. Отправитель отправляет `/private BG Текст`
2. Chat Service проверяет: BG онлайн?
3. Если **онлайн** → отправка через WebSocket Broadcast
4. Если **офлайн** → сохранение в `private_messages_offline`
5. При подключении BG → загрузка непрочитанных сообщений
6. Через 3 суток → автоматическое удаление

---

### 🗄️ База данных

#### Новая таблица: private_messages_offline

**Назначение:** Временное хранение приватных сообщений для офлайн-получателей.

```sql
CREATE TABLE IF NOT EXISTS private_messages_offline (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES users(id),
    sender_username VARCHAR(100) NOT NULL,
    recipient_id INTEGER NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + INTERVAL '3 days')
);
```

**Индексы:**
```sql
CREATE INDEX idx_private_offline_recipient ON private_messages_offline(recipient_id);
CREATE INDEX idx_private_offline_expires ON private_messages_offline(expires_at);
CREATE INDEX idx_private_offline_unread ON private_messages_offline(recipient_id, is_read) WHERE is_read = FALSE;
```

**Функция очистки:**
```sql
CREATE OR REPLACE FUNCTION fn_cleanup_old_private_offline()
RETURNS void AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO deleted_count
    FROM private_messages_offline
    WHERE expires_at < NOW();

    DELETE FROM private_messages_offline
    WHERE expires_at < NOW();

    RAISE NOTICE 'Очистка старых приватных сообщений: удалено % записей', deleted_count;
END;
$$ LANGUAGE plpgsql;
```

**Автоматизация очистки (cron):**
```bash
# Ежедневная очистка в 3:00
0 3 * * * docker exec stalknet-postgres psql -U stalknet -d stalknet -c "SELECT fn_cleanup_old_private_offline();"
```

---

### 🔌 Chat Service (:8083)

#### Hub: Отслеживание онлайн-статуса

**Новые поля:**
```go
type Hub struct {
    Clients    map[int]map[*Client]bool
    UserOnline map[int]bool  // userID -> online статус
    // ...
}
```

**Методы:**
- `IsUserOnline(userID int) bool` — проверка онлайн-статуса
- `GetOnlineUsers() []int` — список всех онлайн-пользователей
- `SetUserOnline(userID, online bool)` — установка статуса

**Логика:**
- При регистрации клиента → `UserOnline[userID] = true`
- При отключении → проверка: есть ли другие клиенты этого пользователя
- Если нет → `delete(UserOnline, userID)`

---

#### Repository: Офлайн-сообщения

**Модель:**
```go
type OfflinePrivateMessage struct {
    ID             int
    SenderID       int
    SenderUsername string
    RecipientID    int
    Content        string
    IsRead         bool
    CreatedAt      time.Time
    ExpiresAt      time.Time  // +3 суток
}
```

**Методы:**
- `SaveOfflinePrivateMessage(ctx, msg)` — сохранение офлайн-сообщения
- `GetUnreadOfflineMessages(ctx, recipientID)` — загрузка непрочитанных
- `MarkOfflineMessagesAsRead(ctx, recipientID)` — пометка как прочитанные

---

#### WebSocket: Обработка private_message

**Обновлённая логика:**

```go
if msgType == "private_message" {
    recipientID, recipientName := h.findUserByUsername(...)
    
    // Проверяем: получатель онлайн?
    isOnline := h.hub.IsUserOnline(recipientID)
    
    if isOnline {
        // Отправка через WebSocket
        h.hub.BroadcastPrivate(...)
    } else {
        // Сохранение в БД для офлайн-получателя
        offlineMsg := &repository.OfflinePrivateMessage{...}
        go h.repo.SaveOfflinePrivateMessage(ctx, offlineMsg)
    }
    
    // Сохранение в chat_messages (ФЗ-374)
    // Подтверждение отправителю
}
```

---

#### API Endpoints

**GET /api/chat/offline-messages**
- Получение непрочитанных офлайн-сообщений
- Требуется JWT токен
- Ответ: `{"messages": [...], "count": 5}`

**POST /api/chat/offline-messages/read**
- Пометка всех сообщений как прочитанных
- Требуется JWT токен
- Ответ: `{"message": "Messages marked as read"}`

**JWT Middleware:**
```go
func JWTMiddleware() gin.HandlerFunc {
    // Проверка JWT токена
    // Извлечение user_id из claims
    // Установка c.Set("user_id", userID)
}
```

---

### 🎨 Frontend

#### Загрузка офлайн-сообщений

**При подключении WebSocket:**
```javascript
ws.onopen = function() {
    wsConnected = true;
    
    // Загружаем непрочитанные офлайн-сообщения
    if (authState === AuthState.Authorized && accessToken) {
        loadOfflinePrivateMessages();
    }
};
```

**Функции:**

```javascript
async function loadOfflinePrivateMessages() {
    const resp = await fetch(API_BASE + "/api/chat/offline-messages", {
        headers: { "Authorization": "Bearer " + accessToken }
    });
    
    const data = await resp.json();
    if (data.messages && data.messages.length > 0) {
        addMessage("📬 Получено N новых приватных сообщений:", "system");
        
        data.messages.forEach(msg => {
            const formattedText = `[${msg.sender_username}] private [${username}] ${msg.content}`;
            addMessage(formattedText, "private", msg.sender_username, {...});
        });
        
        markOfflineMessagesAsRead();
    }
}

async function markOfflineMessagesAsRead() {
    await fetch(API_BASE + "/api/chat/offline-messages/read", {
        method: "POST",
        headers: { "Authorization": "Bearer " + accessToken }
    });
}
```

---

### 🔒 Безопасность

| Проверка | Описание |
|----------|----------|
| **JWT токен** | Доступ к API только с авторизацией |
| **Проверка user_id** | Извлечение из токена, не из запроса |
| **Срок хранения** | 3 суток (автоматическое удаление) |
| **Сохранение для ФЗ-374** | Все сообщения в `chat_messages` |
| **Доступ к сообщениям** | Только свои (по `recipient_id` из токена) |

---

### 📁 Файлы

| Файл | Изменения |
|------|-----------|
| `deploy/postgres/create_private_messages_offline.sql` | Таблица + функция очистки |
| `services/chat/hub/hub.go` | `UserOnline` карта, `IsUserOnline()` |
| `services/chat/repository/repository.go` | `OfflinePrivateMessage`, методы |
| `services/chat/handlers/handlers.go` | API endpoints, `JWTMiddleware()` |
| `services/chat/handlers/websocket.go` | Проверка онлайн при отправке |
| `client/web/app.js` | `loadOfflinePrivateMessages()`, `markOfflineMessagesAsRead()` |
| `services/chat/go.mod` | `github.com/golang-jwt/jwt/v5` |

---

### 📊 SQL запросы

**Получить непрочитанные сообщения:**
```sql
SELECT * FROM private_messages_offline
WHERE recipient_id = 5 AND is_read = FALSE
ORDER BY created_at ASC;
```

**Статистика по пользователю:**
```sql
SELECT 
    recipient_id,
    COUNT(*) as unread_count,
    MIN(created_at) as oldest_message
FROM private_messages_offline
WHERE is_read = FALSE
GROUP BY recipient_id;
```

**Удаление просроченных:**
```sql
SELECT fn_cleanup_old_private_offline();
```

---

### 📦 Версия

- **Версия:** v0.1.15
- **Дата:** 2026-03-07
- **Статус:** ✅ Released

---

## [v0.1.14] - 2026-03-07

### 🎯 Личные сообщения (Private Messages)

#### Архитектура
Реализована система личных сообщений с фильтрацией видимости и сохранением в базу данных.

**Формат команды:**
```
/private <имя> <текст>
```

**Формат отображения:**
```
[11:55] [Alex] private [BG] Привет, как дела?
```

---

### 🗄️ База данных

#### Изменения в chat_messages

**Новая колонка:**
```sql
ALTER TABLE chat_messages 
ADD COLUMN contacts JSONB DEFAULT NULL;
```

**Структура contacts:**
```json
[
  {"id": 5, "name": "Alex"},
  {"id": 3, "name": "BG"}
]
```

Где:
- Первый элемент — **отправитель** (sender)
- Второй элемент — **получатель** (recipient)

**Индексы:**
```sql
CREATE INDEX idx_chat_messages_contacts ON chat_messages USING GIN(contacts);
CREATE INDEX idx_chat_messages_message_type ON chat_messages(message_type);
```

**Типы сообщений:**
- `message` — обычное сообщение
- `system` — системное сообщение
- `task` — уведомление о задаче
- `private` — личное сообщение

---

### 🔌 Auth Service (:8081)

#### Новые endpoint'ы

**GET /api/users/search?username=<query>**
- Поиск пользователя по имени (частичное совпадение)
- Возвращает до 10 результатов
- Пример: `/api/users/search?username=BG`

**GET /api/users/:id**
- Информация о пользователе по ID
- Пример: `/api/users/5`

**Файлы:**
- `services/auth/handlers/auth.go` — функции `SearchUsers()`, `GetUserByID()`
- `services/auth/handlers/handlers.go` — маршруты `/api/users/*`
- `services/auth/repository/repository.go` — метод `SearchUsersByUsername()`

---

### 💬 Chat Service (:8083)

#### Обработка личных сообщений

**WebSocket сообщение (клиент → сервер):**
```json
{
  "type": "private_message",
  "recipient_username": "BG",
  "content": "Привет!"
}
```

**Логика обработки:**
1. Проверка авторизации (JWT токен)
2. Поиск получателя через Auth Service API
3. Проверка: получатель существует
4. Создание contacts: `[отправитель, получатель]`
5. Сохранение в `chat_messages` с `message_type='private'`
6. Broadcast с фильтрацией (видят только contacts)

**Файлы:**
- `services/chat/handlers/websocket.go`:
  - `readPump()` — обработка `private_message`
  - `sendPrivateMessageToCompliance()` — отправка в Compliance
  - `sendError()` — отправка ошибки клиенту
  - `sendPrivateMessageSent()` — подтверждение отправителю

- `services/chat/handlers/handlers.go`:
  - `findUserByUsername()` — поиск через Auth Service
  - `NewChatHandler()` — инициализация `authBaseURL`

- `services/chat/repository/repository.go`:
  - `Contact` — модель контакта
  - `ChatMessage.Contacts` — поле для contacts
  - `SavePrivateMessage()` — сохранение личного сообщения
  - `GetMessagesByRoom()` — загрузка с contacts

- `services/chat/hub/hub.go`:
  - `Contact` — модель контакта
  - `BroadcastPrivate()` — отправка только contacts

---

### 🎨 Frontend

#### Команда /private

**Обработка команды:**
```javascript
function handlePrivateMessage(args) {
    // Проверка авторизации
    // Проверка: не самому себе
    // Отправка через WebSocket
    ws.send(JSON.stringify({
        type: "private_message",
        recipient_username: recipientUsername,
        content: messageContent
    }));
}
```

**Файлы:**
- `client/web/app.js`:
  - `handlePrivateMessage()` — обработка команды
  - `preparePrivateMessage()` — подготовка по клику
  - `attachUsernameClickHandlers()` — обработчики клика
  - `addMessage()` — отображение с классом `private`
  - `ws.onmessage` — обработка входящих

#### Отображение

**CSS стили:**
```css
.message.private {
    padding: 8px 12px;
    margin: 4px 0;
}
.message.private .sender-name {
    color: #cf4e4e;  /* Красный текст */
    font-weight: bold;
}
.message.private .recipient-name {
    color: #cf4e4e;  /* Красный текст */
    font-weight: bold;
}
.message.private .private-label {
    color: #cf4e4e;
    font-style: italic;
}
.message.private .content {
    color: #cf4e4e;  /* Красный текст */
}
```

**Формат:**
```
[11:55] [Alex] private [BG] Привет!
       ↑        ↑       ↑        ↑
    время  отправитель  |    получатель  текст
                      метка
```

Все элементы красные (`#cf4e4e`), фон обычный.

#### Клик по имени

**Функциональность:**
- Клик по имени в обычном сообщении → `/private <имя> `
- Клик по имени отправителя в личном → `/private <имя> `
- Клик по имени получателя в личном → `/private <имя> `

**Обработчики:**
- `.message .username` — обычные сообщения
- `.message.private .sender-name` — отправитель
- `.message.private .recipient-name` — получатель

**Hover эффект:**
- Подсветка: `#ffffff`
- Тень: `0 0 5px rgba(255, 255, 255, 0.5)`

---

### 📚 Обновление справки

**static_content:**
```sql
UPDATE static_content 
SET content = '───
• /help - Эта справка
• /private <имя> <текст> - Личное сообщение (видно только получателю)
───'
WHERE content_key = 'help_authorized';
```

**Файлы:**
- `deploy/postgres/init.sql` — начальная справка
- `deploy/postgres/update_help_private.sql` — миграция

---

### 🔒 Безопасность

**Проверки:**
- ✅ Только для авторизованных пользователей
- ✅ Нельзя отправить самому себе
- ✅ Фильтрация broadcast (видят только contacts)
- ✅ Сохранение в БД для ФЗ-374
- ✅ Логирование в Compliance Service

**Поток данных:**
```
Отправитель → Chat Service → Проверка JWT
                          → Поиск получателя (Auth API)
                          → Сохранение (БД)
                          → Broadcast (фильтрация)
                          → Compliance (ФЗ-374)
```

---

### 📦 Версия

- **Версия:** v0.1.14
- **Дата:** 2026-03-07
- **Статус:** ✅ Released

---

## [v0.1.13] - 2026-03-02

### 🎨 Дизайн и UI

#### Цвет системных сообщений
- Изменён цвет системных сообщений с `#cccccc` (светло-серый) на `#00ff00` (ярко-зелёный)
- Файлы: `client/web/index.html`, `gateway/web/index.html`

#### Подсветка собственных сообщений
- Добавлен CSS класс `.message.own` для сообщений текущего пользователя
- Цвет: `#cf9c4e` (золотисто-оранжевый, как имя пользователя)
- Стиль: жирный (`font-weight: bold`)
- Логика: JavaScript автоматически добавляет класс `own` если `msgUsername === username`
- Файлы: `client/web/index.html`, `client/web/app.js`, `gateway/web/index.html`, `gateway/web/app.js`

---

### 🗄️ База данных

#### Двухтабличная архитектура сообщений

**Таблица `messages` (оперативная):**
- Хранит последние 50 сообщений на комнату
- Быстрая загрузка истории при подключении
- Автоматическая очистка старых сообщений (>50)
- Индекс: `idx_messages_room_created` (room_id, created_at DESC)

**Таблица `chat_messages` (архивная, ФЗ-374):**
- Полный архив всех сообщений
- Хранение метаданных (IP, порт)
- Срок хранения: 1 год

**Метод сохранения:**
```go
ChatRepository.SaveMessage() {
    BEGIN TRANSACTION
    → INSERT INTO messages (room_id, user_id, content)
    → INSERT INTO chat_messages (room_id, user_id, username, content, client_ip, ...)
    → DELETE FROM messages (оставить последние 50)
    → COMMIT
}
```

---

### 🔄 WebSocket

#### Загрузка истории при подключении

**Поток данных:**
1. Пользователь подключается: `/ws/chat?room_id=1&user_id=5&username=BG`
2. ChatHandler загружает последние 50 сообщений из `messages`
3. Отправляет клиенту с флагом `from_history: true`
4. Отправляет системное сообщение о подключении

**Код:**
```go
messages, err := h.repo.GetRecentMessages(ctx, roomID, 50)
for _, msg := range messages {
    conn.WriteMessage(websocket.TextMessage, jsonData)
}
```

#### Сохранение новых сообщений

**readPump обновлён:**
```go
if msg.Type == "message" {
    chatMsg := &repository.ChatMessage{...}
    go h.repo.SaveMessage(ctx, chatMsg)  // В обе таблицы
    go sendToComplianceService(...)      // Для ФЗ-374
}
```

---

### 📚 Документация

#### Новые файлы
- **DATABASE.md** - полная документация по базе данных
  - Описание всех 9 таблиц
  - Индексы и комментарии
  - Примеры SQL-запросов
  - Поток данных при подключении WebSocket
  - История изменений v0.1.13

- **SERVER_ACCESS.md** - доступ к серверу
  - SSH-подключение
  - PostgreSQL подключение
  - Управление сервисами
  - Диагностика проблем

#### Обновлённые файлы
- **README.md**
  - Добавлен раздел "Подключение к серверу"
  - Ссылки на DATABASE.md, SERVER_ACCESS.md
  - Удалены ссылки на устаревшие файлы

- **DEPLOYMENT_CLOUD_RU.md**
  - Добавлены SSH-команды для подключения
  - Таблица диагностики проблем SSH
  - Проверка доступных ключей

- **DATABASE.md** (обновления)
  - Раздел о таблице messages
  - Поток данных WebSocket
  - История изменений v0.1.13

#### Удалённые файлы (устаревшие)
- `USERS_TABLE.md` → заменён на DATABASE.md
- `ROOMS_TABLE.md` → заменён на DATABASE.md
- `TASKS_TABLE.md` → заменён на DATABASE.md
- `USER_EVENTS.md` → заменён на DATABASE.md
- `USER_SESSIONS.md` → заменён на DATABASE.md
- `CHAT_MESSAGES_STORAGE.md` → заменён на DATABASE.md

---

### 📦 Версия

- **Версия:** v0.1.13
- **Дата:** 2026-03-02
- **Статус:** ✅ Released

---

## [v0.1.12] - 2026-03-01

### Изменения
- Обновлены файлы web-клиента
- Исправлены стили сообщений

---

## [v0.1.11] - 2026-02-28

### Изменения
- Обновлена документация сервисов
- Исправлены ошибки в AUTH_SERVICE.md, CHAT_SERVICE.md

---

## 📊 Статистика изменений v0.1.13

**Всего коммитов:** 10

| Тип | Количество |
|-----|------------|
| feat | 2 |
| docs | 5 |
| style | 2 |
| chore | 1 |

**Изменённые файлы:**
- `services/chat/handlers/websocket.go`
- `services/chat/handlers/handlers.go`
- `services/chat/repository/repository.go`
- `client/web/index.html`
- `client/web/app.js`
- `gateway/web/index.html`
- `gateway/web/app.js`
- `deploy/postgres/init.sql`
- `DATABASE.md`
- `README.md`
- `DEPLOYMENT_CLOUD_RU.md`
- `SERVER_ACCESS.md` (новый)

**Удалённые файлы:** 6 устаревших документов

**Добавленные файлы:** 2 новых документа

---

## 🔗 Ссылки

- **GitHub:** https://github.com/AlexDumper/STALKnet
- **Production:** http://87.242.103.13:8080
- **Документация:** DATABASE.md, SERVER_ACCESS.md
