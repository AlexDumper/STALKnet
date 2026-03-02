# 📊 Таблица rooms - Комнаты чата

> ⚠️ **Этот файл устарел!** Актуальная документация: **[DATABASE.md](DATABASE.md#2-rooms-комнаты)**

## 📋 Описание

Таблица `rooms` хранит информацию о комнатах для чата в которых пользователи могут обмениваться сообщениями.

**⚠️ Статус:** 🟡 **В РАЗРАБОТКЕ** (таблица создана но функционал комнат не реализован)

---

## 📊 Структура таблицы

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

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор комнаты |
| `name` | VARCHAR(100) | Название комнаты |
| `description` | TEXT | Описание комнаты (опционально) |
| `created_by` | INTEGER | ID пользователя создавшего комнату |
| `is_private` | BOOLEAN | Флаг приватности (false = публичная) |
| `created_at` | TIMESTAMP | Дата и время создания |

---

## 🔧 Индексы

```sql
-- Поиск по названию комнаты
CREATE INDEX idx_rooms_name ON rooms(name);
```

---

## 🔄 Типы комнат

| Тип | Описание |
|-----|----------|
| **Публичная** (`is_private = false`) | Доступна всем пользователям |
| **Приватная** (`is_private = true`) | Доступ только для участников |

---

## 📝 Примеры SQL запросов

### Получить все комнаты

```sql
SELECT 
    r.id,
    r.name,
    r.description,
    u.username as created_by_username,
    r.is_private,
    r.created_at
FROM rooms r
LEFT JOIN users u ON r.created_by = u.id
ORDER BY r.created_at DESC;
```

---

### Получить публичные комнаты

```sql
SELECT id, name, description, created_at
FROM rooms
WHERE is_private = false
ORDER BY name;
```

---

### Получить комнаты созданные пользователем

```sql
SELECT id, name, description, created_at
FROM rooms
WHERE created_by = 5
ORDER BY created_at DESC;
```

---

### Количество комнат по типу

```sql
SELECT 
    is_private,
    COUNT(*) as count
FROM rooms
GROUP BY is_private;
```

---

### Получить комнату с количеством участников

```sql
SELECT 
    r.id,
    r.name,
    r.description,
    COUNT(rm.user_id) as member_count
FROM rooms r
LEFT JOIN room_members rm ON r.id = rm.room_id
WHERE r.id = 1
GROUP BY r.id, r.name, r.description;
```

---

## 🔐 Безопасность

### Создание комнат

- Любой авторизованный пользователь может создать комнату
- Приватные комнаты требуют подтверждения участников

### Доступ к комнатам

- **Публичные комнаты** — доступны всем пользователям
- **Приватные комнаты** — только для участников таблицы `room_members`

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| `users` | `created_by` → `users.id` | Создатель комнаты |
| `room_members` | `room_id` → `rooms.id` | Участники комнаты |
| `messages` | `room_id` → `rooms.id` | Сообщения в комнате |
| `chat_messages` | `room_id` → `rooms.id` | История чата |
| `tasks` | `room_id` → `rooms.id` | Задачи комнаты |

---

## ⚠️ Важные замечания

1. **name** — название комнаты должно быть уникальным
2. **is_private** — по умолчанию `false` (публичная комната)
3. **created_by** — при удалении пользователя комнаты не удаляются (ON DELETE SET NULL)
4. **description** — может содержать Markdown разметку
5. **🟡 Функционал комнат находится в разработке** — таблица создана но API и UI не реализованы

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** 🟡 В РАЗРАБОТКЕ
