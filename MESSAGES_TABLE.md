# 📊 Таблица messages - Сообщения в комнатах

## 📋 Описание

Таблица `messages` хранит историю сообщений отправленных в комнатах чата.

**Примечание:** Это таблица для постоянной истории сообщений. Для real-time чата используется таблица `chat_messages`.

---

## 📊 Структура таблицы

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор сообщения |
| `room_id` | INTEGER | ID комнаты где отправлено сообщение |
| `user_id` | INTEGER | ID пользователя отправившего сообщение |
| `content` | TEXT | Текст сообщения |
| `created_at` | TIMESTAMP | Дата и время отправки |

---

## 🔧 Индексы

```sql
-- Поиск сообщений в комнате
CREATE INDEX idx_messages_room_id ON messages(room_id);

-- Поиск сообщений пользователя
CREATE INDEX idx_messages_user_id ON messages(user_id);

-- Поиск по времени
CREATE INDEX idx_messages_created_at ON messages(created_at);
```

---

## 🔄 Типы сообщений

| Тип | Описание |
|-----|----------|
| **Текстовое** | Обычное текстовое сообщение |
| **Системное** | Системное уведомление |
| **Задача** | Уведомление о задаче |

---

## 📝 Примеры SQL запросов

### Получить последние сообщения комнаты

```sql
SELECT 
    m.id,
    m.room_id,
    m.user_id,
    u.username,
    m.content,
    m.created_at
FROM messages m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.room_id = 1
ORDER BY m.created_at DESC
LIMIT 50;
```

---

### Получить сообщения пользователя

```sql
SELECT 
    m.id,
    r.name as room_name,
    m.content,
    m.created_at
FROM messages m
JOIN rooms r ON m.room_id = r.id
WHERE m.user_id = 5
ORDER BY m.created_at DESC
LIMIT 50;
```

---

### Количество сообщений в комнате

```sql
SELECT 
    r.id,
    r.name,
    COUNT(m.id) as message_count
FROM rooms r
LEFT JOIN messages m ON r.id = m.room_id
GROUP BY r.id, r.name
ORDER BY message_count DESC;
```

---

### Сообщения за сегодня

```sql
SELECT 
    u.username,
    r.name as room_name,
    m.content,
    m.created_at
FROM messages m
JOIN users u ON m.user_id = u.id
JOIN rooms r ON m.room_id = r.id
WHERE DATE(m.created_at) = CURRENT_DATE
ORDER BY m.created_at DESC;
```

---

### Топ пользователей по количеству сообщений

```sql
SELECT 
    u.id,
    u.username,
    COUNT(m.id) as message_count
FROM users u
LEFT JOIN messages m ON u.id = m.user_id
GROUP BY u.id, u.username
ORDER BY message_count DESC
LIMIT 10;
```

---

### Статистика сообщений по дням

```sql
SELECT 
    DATE(m.created_at) as date,
    COUNT(*) as message_count
FROM messages m
WHERE m.created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(m.created_at)
ORDER BY date DESC;
```

---

## 🔐 Безопасность

### Удаление сообщений

- При удалении комнаты сообщения удаляются (`ON DELETE CASCADE`)
- При удалении пользователя сообщения сохраняются (`ON DELETE SET NULL`)

### Модерация

- Администраторы могут удалять любые сообщения
- Пользователи могут удалять только свои сообщения

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| `rooms` | `room_id` → `rooms.id` | Комната |
| `users` | `user_id` → `users.id` | Пользователь |
| `chat_messages` | — | Real-time сообщения чата |

---

## ⚠️ Важные замечания

1. **content** — может содержать Markdown разметку
2. **user_id** — при удалении пользователя становится NULL
3. **created_at** — автоматически устанавливается при создании
4. **room_id** — при удалении комнаты сообщения удаляются

---

## 🔍 Отличия от chat_messages

| Характеристика | messages | chat_messages |
|---------------|----------|---------------|
| Назначение | Постоянная история | Real-time чат |
| Хранение метаданных | Базовые | IP, порт, тип |
| Срок хранения | Бессрочно | 1 год |
| Индексы | 3 индекса | 5 индексов |

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** ✅ Работает
