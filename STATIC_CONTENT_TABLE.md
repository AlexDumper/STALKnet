# 📊 Таблица static_content - Статический контент

> ⚠️ **Этот файл устарел!** Актуальная документация: **[DATABASE.md](DATABASE.md#7-static_content-статический-контент)**

## 📋 Описание

Таблица `static_content` хранит статический контент приложения: справки, инструкции, приветственные сообщения и другой текст который загружается динамически.

---

## 📊 Структура таблицы

```sql
CREATE TABLE static_content (
    id SERIAL PRIMARY KEY,
    content_key VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    min_auth_state INT DEFAULT 0,
    max_auth_state INT DEFAULT 4,
    language VARCHAR(10) DEFAULT 'ru',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Описание полей

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL | Уникальный идентификатор записи |
| `content_key` | VARCHAR(100) | Уникальный ключ контента |
| `title` | VARCHAR(255) | Заголовок контента |
| `content` | TEXT | Содержимое (текст, Markdown, HTML) |
| `content_type` | VARCHAR(20) | Тип контента: `text`, `markdown`, `html` |
| `min_auth_state` | INT | Минимальный уровень авторизации (0-4) |
| `max_auth_state` | INT | Максимальный уровень авторизации (0-4) |
| `language` | VARCHAR(10) | Язык контента (ru, en, etc.) |
| `is_active` | BOOLEAN | Активность записи |
| `created_at` | TIMESTAMP | Дата и время создания |
| `updated_at` | TIMESTAMP | Дата и время обновления |

---

## 🔧 Индексы

```sql
-- Поиск по ключу контента
CREATE INDEX idx_static_content_key ON static_content(content_key);

-- Поиск по уровню авторизации
CREATE INDEX idx_static_content_auth ON static_content(min_auth_state, max_auth_state);

-- Поиск по языку
CREATE INDEX idx_static_content_language ON static_content(language);
```

---

## 🔄 Уровни авторизации

| Уровень | Описание |
|---------|----------|
| `0` | Guest (гость, не авторизован) |
| `1` | EnteringName (ввод имени) |
| `2` | ConfirmCreate (подтверждение создания) |
| `3` | EnteringPassword (ввод пароля) |
| `4` | Authorized (авторизованный пользователь) |

---

## 📝 Примеры записей

### Приветственное сообщение (help_welcome)

```sql
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state)
VALUES (
    'help_welcome',
    'Приветствие',
    'Добро пожаловать в STALKnet!
Введите /help для списка команд
Введите /auth для авторизации',
    0, 4
);
```

### Справка для гостей (help_guest)

```sql
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state)
VALUES (
    'help_guest',
    'Базовые команды',
    '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
───',
    0, 0
);
```

### Справка для авторизованных (help_authorized)

```sql
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state)
VALUES (
    'help_authorized',
    'Полный список команд',
    '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
• /nick <name> - Сменить имя
• /mock <text> - Отправить сообщение
• /mockmsg - Случайное сообщение
• /mocktask - Показать задание
───',
    4, 4
);
```

---

## 📝 Примеры SQL запросов

### Получить контент по ключу

```sql
SELECT content_key, title, content, content_type
FROM static_content
WHERE content_key = 'help_welcome'
  AND is_active = TRUE;
```

---

### Получить контент для уровня авторизации

```sql
SELECT content_key, title, content, content_type
FROM static_content
WHERE min_auth_state <= 4
  AND max_auth_state >= 4
  AND is_active = TRUE
  AND language = 'ru'
ORDER BY content_key;
```

---

### Получить все активные записи

```sql
SELECT 
    content_key,
    title,
    content_type,
    language,
    created_at
FROM static_content
WHERE is_active = TRUE
ORDER BY content_key;
```

---

### Обновить контент

```sql
UPDATE static_content
SET content = 'Новый текст справки',
    updated_at = NOW()
WHERE content_key = 'help_guest';
```

---

### Получить контент по языку

```sql
SELECT content_key, title, content
FROM static_content
WHERE language = 'ru'
  AND is_active = TRUE
ORDER BY content_key;
```

---

## 🔐 Безопасность

### Управление контентом

- Только администраторы могут добавлять/изменять контент
- Изменения контента не требуют пересборки приложения

### Доступ к контенту

- Контент фильтруется по уровню авторизации
- Неактивный контент (`is_active = FALSE`) не отображается

---

## 📊 Связанные таблицы

| Таблица | Связь | Описание |
|---------|-------|----------|
| — | — | Нет внешних ключей |

---

## ⚠️ Важные замечания

1. **content_key** — уникальный ключ для доступа к контенту
2. **content_type** — определяет формат отображения
3. **min_auth_state / max_auth_state** — диапазон уровней авторизации
4. **language** — для поддержки многоязычности
5. **is_active** — позволяет временно отключать контент без удаления

---

## 🔍 API для получения контента

### GET /api/content/:key

**Параметры:**
- `key` — ключ контента (help_welcome, help_guest, etc.)
- `auth_state` — уровень авторизации (0-4)

**Пример:**
```bash
curl "http://localhost:8081/api/content/help_welcome?auth_state=4"
```

**Ответ:**
```json
{
  "key": "help_welcome",
  "type": "text",
  "title": "Приветствие",
  "content": "Добро пожаловать в STALKnet!..."
}
```

---

**Дата создания:** 2026-03-01  
**Версия:** v0.1.11  
**Статус:** ✅ Работает
