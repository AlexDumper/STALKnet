# 📊 Таблица tasks - Задачи

## 📋 Описание

Таблица `tasks` хранит информацию о задачах в системе STALKnet.

**⚠️ Статус:** 🟡 **В РАЗРАБОТКЕ** (таблица создана но функционал задач не реализован)

---

## 📊 Структура таблицы

```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    assignee_id INTEGER REFERENCES users(id),
    room_id INTEGER REFERENCES rooms(id),
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор задачи |
| `title` | VARCHAR(200) | Заголовок задачи |
| `description` | TEXT | Описание задачи (опционально) |
| `creator_id` | INTEGER | ID пользователя создавшего задачу |
| `assignee_id` | INTEGER | ID пользователя назначенного на задачу |
| `room_id` | INTEGER | ID комнаты к которой относится задача |
| `status` | VARCHAR(20) | Статус задачи |
| `created_at` | TIMESTAMP | Дата и время создания |
| `completed_at` | TIMESTAMP | Дата и время выполнения |
| `confirmed_at` | TIMESTAMP | Дата и время подтверждения |

---

## 🔧 Индексы

```sql
-- Поиск по создателю
CREATE INDEX idx_tasks_creator_id ON tasks(creator_id);

-- Поиск по исполнителю
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);

-- Поиск по статусу
CREATE INDEX idx_tasks_status ON tasks(status);

-- Поиск по комнате
CREATE INDEX idx_tasks_room_id ON tasks(room_id);
```

---

## 🔄 Статусы задач

| Статус | Описание |
|--------|----------|
| `open` | Задача открыта |
| `in_progress` | Задача в работе |
| `done` | Задача выполнена |
| `confirmed` | Задача подтверждена заказчиком |

---

## 📝 Примеры SQL запросов

### Получить все задачи

```sql
SELECT 
    t.id,
    t.title,
    t.description,
    t.status,
    c.username as creator,
    a.username as assignee,
    r.name as room_name,
    t.created_at,
    t.completed_at
FROM tasks t
LEFT JOIN users c ON t.creator_id = c.id
LEFT JOIN users a ON t.assignee_id = a.id
LEFT JOIN rooms r ON t.room_id = r.id
ORDER BY t.created_at DESC;
```

---

### Получить открытые задачи

```sql
SELECT id, title, description, creator_id, assignee_id, created_at
FROM tasks
WHERE status = 'open'
ORDER BY created_at DESC;
```

---

### Получить задачи пользователя

```sql
-- Задачи созданные пользователем
SELECT id, title, status, created_at
FROM tasks
WHERE creator_id = 5
ORDER BY created_at DESC;

-- Задачи назначенные пользователю
SELECT id, title, status, created_at
FROM tasks
WHERE assignee_id = 5
ORDER BY created_at DESC;
```

---

### Статистика задач по статусам

```sql
SELECT 
    status,
    COUNT(*) as count
FROM tasks
GROUP BY status;
```

---

### Задачи в конкретной комнате

```sql
SELECT 
    t.id,
    t.title,
    t.status,
    a.username as assignee,
    t.created_at
FROM tasks t
LEFT JOIN users a ON t.assignee_id = a.id
WHERE t.room_id = 1
ORDER BY t.created_at DESC;
```

---

### Просроченные задачи (выполняются дольше 7 дней)

```sql
SELECT 
    id,
    title,
    assignee_id,
    created_at,
    NOW() - created_at as duration
FROM tasks
WHERE status = 'in_progress'
  AND created_at < NOW() - INTERVAL '7 days'
ORDER BY created_at;
```

---

### Выполненные задачи за сегодня

```sql
SELECT 
    t.id,
    t.title,
    a.username as assignee,
    t.completed_at
FROM tasks t
LEFT JOIN users a ON t.assignee_id = a.id
WHERE t.status = 'done'
  AND DATE(t.completed_at) = CURRENT_DATE
ORDER BY t.completed_at DESC;
```

---

## 🔐 Безопасность

### Создание задач

- Любой авторизованный пользователь может создать задачу
- Задача может быть привязана к комнате

### Назначение исполнителей

- Создатель задачи может назначить исполнителя
- Исполнитель может отказаться от задачи

### Выполнение задач

- Исполнитель может отметить задачу как выполненную
- Создатель должен подтвердить выполнение

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| `users` | `creator_id` → `users.id` | Создатель задачи |
| `users` | `assignee_id` → `users.id` | Исполнитель задачи |
| `rooms` | `room_id` → `rooms.id` | Комната задачи |
| `chat_messages` | — | Уведомления о задачах |

---

## ⚠️ Важные замечания

1. **title** — обязательное поле, максимум 200 символов
2. **status** — по умолчанию `open`
3. **completed_at** — устанавливается при выполнении задачи
4. **confirmed_at** — устанавливается при подтверждении заказчиком
5. **assignee_id** — может быть NULL (задача без исполнителя)
6. **🟡 Функционал задач находится в разработке** — таблица создана но API и UI не реализованы

---

## 📈 Жизненный цикл задачи

```
1. CREATE (open)
   ↓
2. ASSIGN (in_progress)
   ↓
3. COMPLETE (done)
   ↓
4. CONFIRM (confirmed)
```

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** 🟡 В РАЗРАБОТКЕ
