# 🔌 Доступ к серверу STALKnet

## 📋 Production-сервер

| Параметр | Значение |
|----------|----------|
| **IP-адрес** | `87.242.103.13` |
| **Пользователь SSH** | `root` |
| **Порт SSH** | `22` |
| **SSH-ключ** | `~/.ssh/id_ed25519` |

---

## 🔑 SSH-подключение

### Windows (PowerShell)

```powershell
# Полный путь к ключу
ssh -i C:\Users\User\.ssh\id_ed25519 root@87.242.103.13

# Через переменную окружения
ssh -i $env:USERPROFILE\.ssh\id_ed25519 root@87.242.103.13
```

### Linux / macOS

```bash
ssh -i ~/.ssh/id_ed25519 root@87.242.103.13
```

### Проверка ключа

```powershell
# Показать отпечаток ключа
ssh-keygen -lf C:\Users\User\.ssh\id_ed25519

# Показать все ключи
dir $env:USERPROFILE\.ssh\id_*
```

---

## 🗄️ Подключение к PostgreSQL

Порт 5432 открыт для внешнего подключения.

### Через psql (командная строка)

**Windows (PowerShell):**
```powershell
$env:PGPASSWORD="stalknet_secret"; psql -h 87.242.103.13 -U stalknet -d stalknet
```

**Linux / macOS:**
```bash
PGPASSWORD=stalknet_secret psql -h 87.242.103.13 -U stalknet -d stalknet
```

### Через connection string

```bash
psql "postgresql://stalknet:stalknet_secret@87.242.103.13:5432/stalknet"
```

### Параметры для GUI-клиентов (DBeaver, pgAdmin, DataGrip)

```
Host:     87.242.103.13
Port:     5432
Database: stalknet
Username: stalknet
Password: stalknet_secret
```

---

## 🐳 Управление сервисами на сервере

### Подключиться по SSH
```bash
ssh -i ~/.ssh/id_ed25519 root@87.242.103.13
```

### Проверить статус контейнеров
```bash
docker ps --filter "name=stalknet"
```

### Просмотр логов
```bash
# Все сервисы
docker-compose -f docker-compose.prod.yml logs -f

# Конкретный сервис
docker-compose -f docker-compose.prod.yml logs -f auth
```

### Перезапуск сервисов
```bash
# Все сервисы
docker-compose -f docker-compose.prod.yml restart

# Конкретный сервис
docker-compose -f docker-compose.prod.yml restart auth
```

### Остановка и запуск
```bash
# Остановка
docker-compose -f docker-compose.prod.yml down

# Запуск
docker-compose -f docker-compose.prod.yml up -d
```

---

## 🔍 Диагностика

### Проверка доступности порта 5432 (PostgreSQL)

**Windows (PowerShell):**
```powershell
Test-NetConnection 87.242.103.13 -Port 5432
```

**Linux:**
```bash
nc -zv 87.242.103.13 5432
```

### Проверка доступности порта 22 (SSH)

**Windows (PowerShell):**
```powershell
Test-NetConnection 87.242.103.13 -Port 22
```

---

## ⚠️ Решение проблем

### SSH: `Permission denied (publickey)`

**Причины:**
1. Неверный SSH-ключ
2. Ключ не добавлен на сервер
3. Неправильный пользователь (используйте `root`)

**Решение:**
```bash
# Убедитесь, что используете правильный ключ
ssh -i C:\Users\User\.ssh\id_ed25519 root@87.242.103.13
```

### PostgreSQL: `Connection refused`

**Причины:**
1. Порт 5432 закрыт в группе безопасности Cloud.ru
2. PostgreSQL не запущен

**Решение:**
```bash
# Подключиться по SSH
ssh -i ~/.ssh/id_ed25519 root@87.242.103.13

# Проверить статус PostgreSQL
docker ps | grep postgres

# Перезапустить PostgreSQL
docker-compose -f docker-compose.prod.yml restart postgres
```

### PostgreSQL: `Password authentication failed`

**Решение:**
- Проверьте пароль в файле `.env` или `.env.production`
- По умолчанию: `stalknet_secret`

---

## 📊 Полезные SQL-запросы

```sql
-- Проверка таблиц
\dt

-- Количество пользователей
SELECT COUNT(*) FROM users;

-- Последние сообщения чата
SELECT username, content, timestamp 
FROM chat_messages 
ORDER BY timestamp DESC 
LIMIT 10;

-- Активные сессии (без logout)
SELECT * FROM user_sessions 
WHERE logout_time IS NULL;

-- События пользователей
SELECT event_type, username, timestamp 
FROM user_events 
ORDER BY timestamp DESC 
LIMIT 10;
```

---

## 📝 Файлы конфигурации

| Файл | Описание |
|------|----------|
| `.env` | Локальная конфигурация |
| `.env.production` | Production-конфигурация (если существует) |
| `docker-compose.yml` | Docker-конфигурация |
| `docker-compose.prod.yml` | Production Docker-конфигурация |

---

**Дата обновления:** 2026-03-02  
**Статус:** ✅ Актуально
