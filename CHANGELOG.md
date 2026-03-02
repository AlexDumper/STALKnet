# 📝 Журнал изменений STALKnet

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
