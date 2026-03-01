# 🔔 Notification Service

## 📋 Описание

**Notification Service** — микросервис для управления уведомлениями пользователей в системе STALKnet.

Сервис отвечает за:
- ✅ Создание и хранение уведомлений
- ✅ Отправку push-уведомлений (WebSocket)
- ✅ Управление статусом прочтения
- ✅ Фильтрацию по типам уведомлений
- ✅ Кэширование в Redis

---

## 🏗️ Архитектура

```
┌─────────────┐     HTTP GET       ┌──────────────────┐
│ Web Client  │ ─────────────────► │ Notification     │
│  (Browser)  │  /api/notification │ Service :8085    │
└─────────────┘                    └────────┬─────────┘
                                            │
                         ┌──────────────────┴──────────────────┐
                         ▼                                     ▼
                  ┌──────────────┐                     ┌──────────────┐
                  │    Redis     │                     │  WebSocket   │
                  │ (notifications)                   │   Hub        │
                  └──────────────┘                     └──────────────┘
```

### Поток данных

#### Создание уведомления:
1. Другой сервис отправляет событие (Task Service, Chat Service)
2. Notification Service создаёт уведомление
3. Сохраняет в Redis
4. Отправляет через WebSocket пользователю

#### Получение уведомлений:
1. Клиент отправляет GET `/api/notification/unread`
2. Notification Service получает из Redis
3. Возвращает список непрочитанных уведомлений

---

## 🔧 API Endpoints

### 🔔 Уведомления

#### GET /api/notification/unread

**Получение непрочитанных уведомлений**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `limit` (опционально) — количество (по умолчанию 50)
- `type` (опционально) — фильтр по типу

**Пример:**
```bash
curl "http://localhost:8085/api/notification/unread?limit=20"
```

**Response (200 OK):**
```json
{
  "notifications": [
    {
      "id": 1,
      "type": "task_assigned",
      "title": "Новая задача",
      "message": "Вам назначена задача: Найти артефакт",
      "created_at": "2026-03-01T12:00:00Z",
      "is_read": false
    },
    {
      "id": 2,
      "type": "task_completed",
      "title": "Задача выполнена",
      "message": "Задача 'Найти артефакт' выполнена",
      "created_at": "2026-03-01T13:00:00Z",
      "is_read": false
    }
  ],
  "total": 2
}
```

---

#### PUT /api/notification/unread/:id/read

**Отметка уведомления как прочитанного**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Path Parameters:**
- `id` — ID уведомления

**Пример:**
```bash
curl -X PUT "http://localhost:8085/api/notification/unread/1/read"
```

**Response (200 OK):**
```json
{
  "message": "Notification marked as read",
  "notification_id": 1
}
```

---

#### PUT /api/notification/read-all

**Отметка всех уведомлений как прочитанные**

**Headers:**
```
Authorization: Bearer <access_token>
```

**Пример:**
```bash
curl -X PUT "http://localhost:8085/api/notification/read-all"
```

**Response (200 OK):**
```json
{
  "message": "All notifications marked as read",
  "marked_count": 5
}
```

---

### 🔌 WebSocket

#### GET /ws/notification

**WebSocket подключение для получения уведомлений в реальном времени**

**Query Parameters:**
- `user_id` — ID пользователя

**Пример подключения:**
```javascript
const ws = new WebSocket(
  `ws://localhost:8085/ws/notification?user_id=5`
);

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('Новое уведомление:', notification);
};
```

**Формат уведомлений (server → client):**
```json
{
  "id": 1,
  "type": "task_assigned",
  "title": "Новая задача",
  "message": "Вам назначена задача: Найти артефакт",
  "created_at": "2026-03-01T12:00:00Z"
}
```

---

## 📋 Типы уведомлений

| Тип | Описание |
|-----|----------|
| `task_assigned` | Назначена новая задача |
| `task_completed` | Задача выполнена |
| `task_confirmed` | Задача подтверждена |
| `task_comment` | Добавлен комментарий к задаче |
| `chat_message` | Новое сообщение в чате (упоминание) |
| `user_mention` | Упоминание пользователя |
| `system` | Системное уведомление |

---

## 🗄️ Хранение данных

### Redis (уведомления)

**Формат ключей:**
- `notification:user:<user_id>:unread` — список непрочитанных уведомлений
- `notification:user:<user_id>:all` — все уведомления пользователя
- `notification:<id>` — данные уведомления

**TTL:**
- Непрочитанные уведомления: 30 дней
- Прочитанные уведомления: 7 дней

**Структура уведомления:**
```json
{
  "id": 1,
  "user_id": 5,
  "type": "task_assigned",
  "title": "Новая задача",
  "message": "Вам назначена задача: Найти артефакт",
  "data": {
    "task_id": 10,
    "creator_id": 3
  },
  "is_read": false,
  "created_at": "2026-03-01T12:00:00Z"
}
```

---

## 🔄 Интеграция с другими сервисами

### Task Service

- Создание задачи → уведомление создателю и исполнителю
- Изменение статуса → уведомление заинтересованным лицам
- Комментарий к задаче → уведомление автору задачи

### Chat Service

- Упоминание пользователя → уведомление упомянутому
- Личное сообщение → уведомление получателю

### Auth Service

- Вход с нового устройства → уведомление пользователю
- Смена имени → системное уведомление

---

## 🚀 Запуск

### Docker Compose

```bash
# Запуск всех сервисов
docker-compose up -d

# Проверка статуса
docker-compose ps notification

# Логи
docker-compose logs -f notification
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервиса | `8085` |
| `REDIS_HOST` | Хост Redis | `localhost` |
| `REDIS_PORT` | Порт Redis | `6379` |

---

## 🔍 Мониторинг

### Health Check

```bash
curl http://localhost:8085/health
```

**Ответ:**
```json
{"status": "ok"}
```

### Проверка количества уведомлений

```bash
docker exec stalknet-redis redis-cli KEYS "notification:user:*:unread"
```

### Просмотр очереди уведомлений

```bash
docker exec stalknet-redis redis-cli LRANGE "notification:user:5:unread" 0 -1
```

---

## 📝 Примеры использования

### Получение непрочитанных уведомлений

```bash
curl "http://localhost:8085/api/notification/unread" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Отметка уведомления как прочитанного

```bash
curl -X PUT "http://localhost:8085/api/notification/unread/1/read" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Отметка всех уведомлений как прочитанные

```bash
curl -X PUT "http://localhost:8085/api/notification/read-all" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### WebSocket подключение

```javascript
const ws = new WebSocket(
  `ws://localhost:8085/ws/notification?user_id=5`
);

ws.onopen = () => {
  console.log('Connected to notification service');
};

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log(`${notification.title}: ${notification.message}`);
};
```

---

## 📊 Примеры уведомлений

### Назначение задачи

```json
{
  "type": "task_assigned",
  "title": "Новая задача",
  "message": "Вам назначена задача: Найти артефакт",
  "data": {
    "task_id": 10,
    "creator_id": 3,
    "creator_username": "BG"
  }
}
```

### Выполнение задачи

```json
{
  "type": "task_completed",
  "title": "Задача выполнена",
  "message": "Задача 'Найти артефакт' выполнена пользователем Stalker",
  "data": {
    "task_id": 10,
    "assignee_id": 7,
    "assignee_username": "Stalker"
  }
}
```

### Упоминание в чате

```json
{
  "type": "user_mention",
  "title": "Вас упомянули",
  "message": "BG упомянул вас в сообщении",
  "data": {
    "room_id": 1,
    "message_id": 123,
    "mentioned_by": "BG"
  }
}
```

---

## ⚠️ Важные замечания

1. **TTL уведомлений:**
   - Непрочитанные: 30 дней
   - Прочитанные: 7 дней

2. **WebSocket** — при отключении уведомления сохраняются в Redis

3. **Лимиты:**
   - Максимум 100 непрочитанных уведомлений на пользователя
   - При превышении старые удаляются

4. **🟡 Статус разработки:** сервис в разработке — базовая структура создана, требуется реализация

---

## 📈 План развития

### Реализовать в будущих версиях:

1. **Redis интеграция** — реализация работы с Redis
2. **WebSocket Hub** — управление подключениями
3. **Email уведомления** — отправка на email
4. **Push уведомления** — браузерные push
5. **Шаблоны** — шаблоны для типов уведомлений
6. **Настройки** — пользовательские настройки уведомлений
7. **Группировка** — группировка однотипных уведомлений
8. **Приоритеты** — приоритет уведомлений (low, normal, high, urgent)

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs notification`
2. Проверьте Redis: `docker exec stalknet-redis redis-cli ping`
3. Проверьте health: `curl http://localhost:8085/health`

---

**Дата создания:** 2026-03-01
**Версия:** v0.1.11
**Статус:** 🟡 В РАЗРАБОТКЕ
