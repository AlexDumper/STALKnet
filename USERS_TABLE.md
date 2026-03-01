# 📊 Таблица users - Пользователи

## 📋 Описание

Таблица `users` хранит информацию обо всех пользователях системы STALKnet.

---

## 📊 Структура таблицы

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор пользователя |
| `username` | VARCHAR(50) | Уникальное имя пользователя (логин) |
| `password_hash` | VARCHAR(255) | Хэш пароля (bcrypt) |
| `email` | VARCHAR(100) | Адрес электронной почты (опционально) |
| `status` | VARCHAR(20) | Статус пользователя: `offline`, `online` |
| `created_at` | TIMESTAMP | Дата и время регистрации |
| `last_seen` | TIMESTAMP | Последнее время активности |

---

## 🔧 Индексы

```sql
-- Поиск по имени пользователя
CREATE INDEX idx_users_username ON users(username);

-- Поиск по статусу
CREATE INDEX idx_users_status ON users(status);
```

---

## 🔄 Статусы пользователей

| Статус | Описание |
|--------|----------|
| `offline` | Пользователь не в сети |
| `online` | Пользователь активен в системе |

---

## 📝 Примеры SQL запросов

### Получить всех пользователей

```sql
SELECT id, username, email, status, created_at, last_seen
FROM users
ORDER BY created_at DESC;
```

---

### Найти пользователя по имени

```sql
SELECT id, username, email, status
FROM users
WHERE username = 'BG';
```

---

### Получить активных пользователей

```sql
SELECT id, username, status, last_seen
FROM users
WHERE status = 'online';
```

---

### Статистика пользователей

```sql
SELECT 
    status,
    COUNT(*) as count
FROM users
GROUP BY status;
```

---

### Пользователи зарегистрированные за сегодня

```sql
SELECT id, username, created_at
FROM users
WHERE DATE(created_at) = CURRENT_DATE
ORDER BY created_at DESC;
```

---

## 🔐 Безопасность

### Хранение паролей

Пароли хранятся в хэшированном виде с использованием алгоритма **bcrypt**.

```go
// Генерация хэша
hashedPassword, _ := bcrypt.GenerateFromPassword(
    []byte(password), 
    bcrypt.DefaultCost
)

// Проверка пароля
err := bcrypt.CompareHashAndPassword(
    []byte(hashedPassword), 
    []byte(password)
)
```

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| `user_events` | `user_id` → `users.id` | События пользователя |
| `user_sessions` | `user_id` → `users.id` | Сессии пользователя |
| `chat_messages` | `user_id` → `users.id` | Сообщения в чате |
| `messages` | `user_id` → `users.id` | Сообщения в комнатах |
| `tasks` | `creator_id` → `users.id` | Созданные задачи |
| `tasks` | `assignee_id` → `users.id` | Назначенные задачи |
| `rooms` | `created_by` → `users.id` | Созданные комнаты |
| `room_members` | `user_id` → `users.id` | Членство в комнатах |

---

## ⚠️ Важные замечания

1. **username** — уникальное поле, нельзя создать двух пользователей с одинаковым именем
2. **password_hash** — никогда не храните пароли в открытом виде
3. **status** — обновляется при входе/выходе из системы
4. **last_seen** — обновляется при активности пользователя

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** ✅ Работает
