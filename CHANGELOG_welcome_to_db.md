# 📝 Изменения: Приветственное сообщение перенесено в базу данных

## 📋 Описание изменений

Приветственное сообщение, которое ранее было зашито в код веб-клиента, теперь хранится в таблице `static_content` базы данных PostgreSQL.

## 🎯 Цель изменений

- **Централизованное управление контентом** - все тексты интерфейса в одном месте
- **Гибкость** - возможность менять приветствие без пересборки приложения
- **Консистентность** - единый подход ко всему статическому контенту

## 📁 Изменённые файлы

### 1. База данных

**Файлы:**
- `deploy/postgres/init.sql` - обновлены начальные данные
- `deploy/postgres/update_content.sql` - скрипт обновления существующих записей

**Новые данные в таблице `static_content`:**

```sql
-- Приветственное сообщение (help_welcome)
content_key: 'help_welcome'
title: 'Приветствие'
content: |
  Добро пожаловать в STALKnet!
  Введите /help для списка команд
  Введите /auth для авторизации
min_auth_state: 0
max_auth_state: 4
```

### 2. Веб-клиент

**Файлы:**
- `client/web/app.js`
- `gateway/web/app.js`

**Изменения:**

**До:**
```javascript
setTimeout(() => {
    connected = true;
    updateStatus();
    loadContent("help_welcome", function() {
        // Дефолтное приветствие (зашито в код)
        addMessage("---", "system");
        addMessage("Добро пожаловать в STALKnet!", "system");
        addMessage("Введите /help для списка команд", "system");
        addMessage("Введите /auth для авторизации", "system");
        addMessage("---", "system");
    });
}, 1000);
```

**После:**
```javascript
setTimeout(() => {
    connected = true;
    updateStatus();
    loadContent("help_welcome", function() {
        // Fallback только если БД недоступна
        console.warn("Контент help_welcome не загружен из БД");
    });
}, 1000);
```

**Обновлена функция `loadContent`:**
```javascript
async function loadContent(key, callback) {
    try {
        const resp = await fetch(API_BASE + "/api/content/" + key + "?auth_state=" + authState);
        if (resp.ok) {
            const data = await resp.json();
            if (data.content) {
                const lines = data.content.split("\n");
                
                // Добавляем разделители для приветственного сообщения
                if (key === "help_welcome") {
                    addMessage("---", "system");
                }
                
                lines.forEach(line => {
                    if (line.trim()) {
                        addMessage(line, "system");
                    }
                });
                
                // Добавляем закрывающие разделители
                if (key === "help_welcome") {
                    addMessage("---", "system");
                }
                
                return;
            }
        }
    } catch (e) {
        console.log("Failed to load content:", e.message);
    }
    if (callback) callback();
}
```

## 🔄 Применение изменений

### Для существующей базы данных

```bash
# Выполнить скрипт обновления
docker exec -i stalknet-postgres psql -U stalknet -d stalknet < deploy/postgres/update_content.sql
```

### Для новой базы данных

При инициализации новой БД через `docker-compose up -d` данные будут загружены автоматически из `init.sql`.

### Пересборка gateway

```bash
# Пересобрать образ gateway
docker-compose build gateway

# Перезапустить сервис
docker-compose restart gateway
```

## 📊 Результат

**Приветственное сообщение в веб-клиенте:**
```
---
Добро пожаловать в STALKnet!
Введите /help для списка команд
Введите /auth для авторизации
---
```

**API endpoint:**
```bash
curl "http://localhost:8080/api/content/help_welcome?auth_state=0"
```

**Ответ:**
```json
{
  "content": "Добро пожаловать в STALKnet!\nВведите /help для списка команд\nВведите /auth для авторизации",
  "key": "help_welcome",
  "title": "Приветствие",
  "type": "text"
}
```

## ✅ Проверка

1. Откройте веб-клиент: http://localhost:8080
2. Должно появиться приветственное сообщение из базы данных
3. Проверьте через pgAdmin:
   ```sql
   SELECT content_key, title, content 
   FROM static_content 
   WHERE content_key = 'help_welcome';
   ```

## 📝 Примечания

- При удалении базы данных и повторной инициализации контент будет восстановлен из `init.sql`
- Для изменения текста приветствия обновите запись в таблице `static_content`:
  ```sql
  UPDATE static_content 
  SET content = 'Ваш новый текст приветствия'
  WHERE content_key = 'help_welcome';
  ```
- Изменения вступят в силу после перезагрузки страницы в браузере (кэш не используется)

---

**Дата внесения изменений:** 2026-03-01  
**Версия:** v0.1.7
