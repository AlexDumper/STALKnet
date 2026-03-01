# 📋 Task Service

## 📋 Описание

**Task Service** — микросервис для управления задачами в системе STALKnet.

Сервис отвечает за:
- ✅ Создание новых задач
- ✅ Назначение исполнителей
- ✅ Изменение статуса задач
- ✅ Отслеживание выполнения
- ✅ Привязка задач к комнатам
- ✅ Получение статистики по задачам

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP POST      ┌──────────────────┐
│ Web Client  │ ─────────────────► │ Task Service     │
│  (Browser)  │  /api/task/*       │ Port: 8084       │
└─────────────┘                    └────────┬─────────┘
                                            │
                                            ▼
                                   ┌──────────────┐
                                   │  PostgreSQL  │
                                   │   (tasks)    │
                                   └──────────────┘
```

### Поток данных

#### Создание задачи:
1. Пользователь отправляет POST `/api/task`
2. Task Service создаёт задачу в PostgreSQL
3. Отправляет уведомление в Notification Service
4. Возвращает созданную задачу

#### Выполнение задачи:
1. Пользователь отправляет PUT `/api/task/:id/complete`
2. Task Service обновляет статус на `done`
3. Устанавливает `completed_at`
4. Отправляет уведомление

---

## 🔧 API Endpoints

### 📝 Задачи

#### GET /api/task

**Получение списка задач**

**Query Parameters:**
- `status` (опционально) — фильтр по статусу (`open`, `in_progress`, `done`, `confirmed`)
- `room_id` (опционально) — фильтр по комнате
- `limit` (опционально) — количество (по умолчанию 50)
- `offset` (опционально) — смещение

**Пример:**
```bash
curl "http://localhost:8084/api/task?status=open&limit=20"
```

**Response (200 OK):**
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Пример задачи",
      "description": "Описание задачи",
      "status": "open",
      "creator_id": 5,
      "creator_username": "BG",
      "assignee_id": null,
      "assignee_username": null,
      "room_id": 1,
      "created_at": "2026-03-01T10:00:00Z",
      "completed_at": null,
      "confirmed_at": null
    }
  ],
  "total": 1
}
```

---

#### POST /api/task

**Создание новой задачи**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "title": "Найти артефакт",
  "description": "Найти и доставить артефакт 'Золотая рыбка'",
  "assignee_id": 7,
  "room_id": 1
}
```

**Требования:**
- **title:** обязательное поле, 1-200 символов
- **description:** опционально
- **assignee_id:** опционально (ID исполнителя)
- **room_id:** опционально (ID комнаты)

**Response (201 Created):**
```json
{
  "message": "Task created successfully",
  "task_id": 1,
  "title": "Найти артефакт"
}
```

---

#### GET /api/task/:id

**Получение задачи по ID**

**Path Parameters:**
- `id` — ID задачи

**Пример:**
```bash
curl "http://localhost:8084/api/task/1"
```

**Response (200 OK):**
```json
{
  "id": 1,
  "title": "Найти артефакт",
  "description": "Найти и доставить артефакт 'Золотая рыбка'",
  "status": "open",
  "creator_id": 5,
  "creator_username": "BG",
  "assignee_id": 7,
  "assignee_username": "Stalker",
  "room_id": 1,
  "created_at": "2026-03-01T10:00:00Z",
  "completed_at": null,
  "confirmed_at": null
}
```

---

#### PUT /api/task/:id

**Обновление задачи**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request:**
```json
{
  "title": "Обновлённое название",
  "description": "Обновлённое описание",
  "assignee_id": 10
}
```

**Response (200 OK):**
```json
{
  "message": "Task updated successfully",
  "task_id": 1
}
```

---

#### PUT /api/task/:id/complete

**Завершение задачи (статус `done`)**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Пример:**
```bash
curl -X PUT "http://localhost:8084/api/task/1/complete"
```

**Response (200 OK):**
```json
{
  "message": "Task completed successfully",
  "task_id": 1,
  "status": "done",
  "completed_at": "2026-03-01T14:00:00Z"
}
```

---

#### PUT /api/task/:id/confirm

**Подтверждение задачи (статус `confirmed`)**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Пример:**
```bash
curl -X PUT "http://localhost:8084/api/task/1/confirm"
```

**Response (200 OK):**
```json
{
  "message": "Task confirmed successfully",
  "task_id": 1,
  "status": "confirmed",
  "confirmed_at": "2026-03-01T15:00:00Z"
}
```

---

#### DELETE /api/task/:id

**Удаление задачи**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Пример:**
```bash
curl -X DELETE "http://localhost:8084/api/task/1"
```

**Response (200 OK):**
```json
{
  "message": "Task deleted successfully",
  "task_id": 1
}
```

---

### 📊 Задачи по комнатам

#### GET /api/task/room/:room_id

**Получение задач конкретной комнаты**

**Path Parameters:**
- `room_id` — ID комнаты

**Query Parameters:**
- `status` (опционально) — фильтр по статусу

**Пример:**
```bash
curl "http://localhost:8084/api/task/room/1?status=open"
```

**Response (200 OK):**
```json
{
  "room_id": 1,
  "tasks": [
    {
      "id": 1,
      "title": "Задача для комнаты 1",
      "status": "open",
      "assignee_username": "Stalker"
    }
  ],
  "total": 1
}
```

---

### 👤 Мои задачи

#### GET /api/task/my/created

**Получение задач, созданных текущим пользователем**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `status` (опционально) — фильтр по статусу

**Response (200 OK):**
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Моя задача",
      "status": "open",
      "assignee_username": "Stalker",
      "created_at": "2026-03-01T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

#### GET /api/task/my/assigned

**Получение задач, назначенных текущему пользователю**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `status` (опционально) — фильтр по статусу

**Response (200 OK):**
```json
{
  "tasks": [
    {
      "id": 2,
      "title": "Назначенная задача",
      "status": "in_progress",
      "creator_username": "BG",
      "created_at": "2026-03-01T11:00:00Z"
    }
  ],
  "total": 1
}
```

---

## 🗄️ Таблицы базы данных

### 1. Таблица `tasks` (задачи)

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

**Индексы:**
```sql
CREATE INDEX idx_tasks_creator_id ON tasks(creator_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_room_id ON tasks(room_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
```

---

## 🔄 Статусы задач

| Статус | Описание |
|--------|----------|
| `open` | Задача открыта, ожидает исполнителя |
| `in_progress` | Задача в работе |
| `done` | Задача выполнена исполнителем |
| `confirmed` | Задача подтверждена заказчиком |

---

## 📈 Жизненный цикл задачи

```
┌─────────────┐
│    CREATE   │
│   (open)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    ASSIGN   │
│(in_progress)│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   COMPLETE  │
│    (done)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   CONFIRM   │
│ (confirmed) │
└─────────────┘
```

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps task

# Логи
docker-compose logs -f task
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8084` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `stalknet` |
| `DB_PASSWORD` | Пароль БД | `stalknet_secret` |
| `DB_NAME` | Имя БД | `stalknet` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8084/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка количества задач

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT status, COUNT(*) as count FROM tasks GROUP BY status;"
```

### Просмотр последних задач

```bash
docker exec -i stalknet-postgres psql -U stalknet -d stalknet -c \
  "SELECT id, title, status, creator_id, assignee_id, created_at
   FROM tasks
   ORDER BY created_at DESC
   LIMIT 10;"
```

---

## 📝 Примеры использования

### Создание задачи

```bash
curl -X POST "http://localhost:8084/api/task" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Найти артефакт",
    "description": "Найти и доставить артефакт",
    "assignee_id": 7,
    "room_id": 1
  }'
```

### Получение списка задач

```bash
curl "http://localhost:8084/api/task?status=open"
```

### Завершение задачи

```bash
curl -X PUT "http://localhost:8084/api/task/1/complete" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Подтверждение задачи

```bash
curl -X PUT "http://localhost:8084/api/task/1/confirm" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## 📊 SQL запросы

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

### Получить открытые задачи

```sql
SELECT id, title, description, creator_id, assignee_id, created_at
FROM tasks
WHERE status = 'open'
ORDER BY created_at DESC;
```

### Задачи пользователя (созданные)

```sql
SELECT id, title, status, created_at
FROM tasks
WHERE creator_id = 5
ORDER BY created_at DESC;
```

### Задачи пользователя (назначенные)

```sql
SELECT id, title, status, created_at
FROM tasks
WHERE assignee_id = 5
ORDER BY created_at DESC;
```

### Статистика по статусам

```sql
SELECT
    status,
    COUNT(*) as count
FROM tasks
GROUP BY status;
```

### Задачи в комнате

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

## ⚠️ Важные замечания

1. **title** — обязательное поле, максимум 200 символов
2. **status** — по умолчанию `open`
3. **completed_at** — устанавливается при завершении задачи
4. **confirmed_at** — устанавливается при подтверждении заказчиком
5. **assignee_id** — может быть NULL (задача без исполнителя)
6. **🟡 Статус разработки:** сервис в разработке — базовая структура создана, требуется реализация репозитория

---

## 📈 План развития

### Реализовать в будущих версиях:

1. **Репозиторий** — реализация работы с PostgreSQL
2. **Уведомления** — интеграция с Notification Service
3. **Комментарии** — возможность комментирования задач
4. **Вложения** — прикрепление файлов к задачам
5. **История изменений** — логирование изменений задач
6. **Дедлайны** — установка срока выполнения
7. **Приоритеты** — приоритет задач (low, medium, high)
8. **Теги** — категоризация задач по тегам

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs task`
2. Проверьте БД: `docker exec -it stalknet-postgres psql -U stalknet -d stalknet`
3. Проверьте health: `curl http://localhost:8084/health`

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** 🟡 В РАЗРАБОТКЕ
