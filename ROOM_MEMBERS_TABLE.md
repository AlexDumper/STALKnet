# 📊 Таблица room_members - Участники комнат

> ⚠️ **Этот файл устарел!** Актуальная документация: **[DATABASE.md](DATABASE.md#3-room_members-участники-комнат)**

## 📋 Описание

Таблица `room_members` хранит информацию о членстве пользователей в комнатах чата.

**⚠️ Статус:** 🟡 **В РАЗРАБОТКЕ** (таблица создана но функционал комнат не реализован)

---

## 📊 Структура таблицы

```sql
CREATE TABLE room_members (
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `room_id` | INTEGER | ID комнаты |
| `user_id` | INTEGER | ID пользователя |
| `joined_at` | TIMESTAMP | Дата и время вступления в комнату |

---

## 🔧 Индексы

```sql
-- Первичный ключ (room_id, user_id)
PRIMARY KEY (room_id, user_id)

-- Поиск пользователей в комнате
CREATE INDEX idx_room_members_room ON room_members(room_id);

-- Поиск комнат пользователя
CREATE INDEX idx_room_members_user ON room_members(user_id);
```

---

## 🔄 Операции с членством

| Операция | Описание |
|----------|----------|
| **JOIN** | Пользователь вступает в комнату |
| **LEAVE** | Пользователь покидает комнату |
| **KICK** | Администратор удаляет пользователя |

---

## 📝 Примеры SQL запросов

### Получить всех участников комнаты

```sql
SELECT 
    rm.room_id,
    rm.user_id,
    u.username,
    u.status,
    rm.joined_at
FROM room_members rm
JOIN users u ON rm.user_id = u.id
WHERE rm.room_id = 1
ORDER BY rm.joined_at DESC;
```

---

### Получить все комнаты пользователя

```sql
SELECT 
    rm.user_id,
    rm.room_id,
    r.name as room_name,
    r.is_private,
    rm.joined_at
FROM room_members rm
JOIN rooms r ON rm.room_id = r.id
WHERE rm.user_id = 5
ORDER BY rm.joined_at DESC;
```

---

### Проверить членство пользователя в комнате

```sql
SELECT EXISTS (
    SELECT 1 FROM room_members
    WHERE room_id = 1 AND user_id = 5
) as is_member;
```

---

### Количество участников в каждой комнате

```sql
SELECT 
    r.id,
    r.name,
    COUNT(rm.user_id) as member_count
FROM rooms r
LEFT JOIN room_members rm ON r.id = rm.room_id
GROUP BY r.id, r.name
ORDER BY member_count DESC;
```

---

### Пользователи в нескольких комнатах

```sql
SELECT 
    u.id,
    u.username,
    COUNT(rm.room_id) as room_count
FROM users u
JOIN room_members rm ON u.id = rm.user_id
GROUP BY u.id, u.username
HAVING COUNT(rm.room_id) > 1
ORDER BY room_count DESC;
```

---

### Новые участники за сегодня

```sql
SELECT 
    r.name as room_name,
    u.username,
    rm.joined_at
FROM room_members rm
JOIN rooms r ON rm.room_id = r.id
JOIN users u ON rm.user_id = u.id
WHERE DATE(rm.joined_at) = CURRENT_DATE
ORDER BY rm.joined_at DESC;
```

---

## 🔐 Безопасность

### Вступление в комнату

- **Публичные комнаты** — любой пользователь может вступить
- **Приватные комнаты** — требуется приглашение или подтверждение

### Выход из комнаты

- Пользователь может покинуть комнату в любой момент
- При удалении комнаты (`ON DELETE CASCADE`) все записи о членстве удаляются

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| `rooms` | `room_id` → `rooms.id` | Комната |
| `users` | `user_id` → `users.id` | Пользователь |
| `messages` | через `room_id` | Сообщения в комнате |

---

## ⚠️ Важные замечания

1. **PRIMARY KEY (room_id, user_id)** — пользователь не может быть в комнате дважды
2. **ON DELETE CASCADE** — при удалении комнаты/пользователя запись удаляется
3. **joined_at** — автоматически устанавливается при вступлении
4. **🟡 Функционал комнат находится в разработке** — таблица создана но API и UI не реализованы

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** 🟡 В РАЗРАБОТКЕ
