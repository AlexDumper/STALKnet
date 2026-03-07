# 🚀 Развёртывание STALKnet через GitHub

**Автоматическое развёртывание** с использованием GitHub Actions для CI/CD.

---

## 📋 Содержание

1. [Обзор](#обзор)
2. [Требования](#требования)
3. [Настройка GitHub Secrets](#настройка-github-secrets)
4. [Настройка сервера](#настройка-сервера)
5. [Автоматическое развёртывание](#автоматическое-развёртывание)
6. [Ручное обновление](#ручное-обновление)
7. [Откат изменений](#откат-изменений)
8. [Мониторинг](#мониторинг)

---

## 📖 Обзор

### Архитектура развёртывания

```
┌─────────────────┐     GitHub Actions     ┌──────────────────┐
│   GitHub.com    │  ───────────────────►  │ Production Server│
│                 │   CI/CD Pipeline       │  87.242.103.13   │
│  AlexDumper/    │                        │                  │
│  STALKnet       │                        │  Docker Compose  │
│                 │                        │  - Gateway       │
│  - main branch  │                        │  - Auth Service  │
│  - PRs          │                        │  - Chat Service  │
│  - Releases     │                        │  - PostgreSQL    │
│                 │                        │  - Redis         │
└─────────────────┘                        └──────────────────┘
```

### Преимущества GitHub Actions

| Преимущество | Описание |
|--------------|----------|
| ✅ Автоматизация | Развёртывание при каждом пуше в main |
| ✅ Безопасность | Секреты хранятся в GitHub, не в коде |
| ✅ Отслеживаемость | История деплоев в GitHub Actions |
| ✅ Откат | Быстрый откат к предыдущей версии |
| ✅ Health checks | Автоматическая проверка работоспособности |

---

## 🔐 Требования

### Для развёртывания необходимо:

1. **GitHub аккаунт** с доступом к репозиторию
2. **Production сервер** с Docker и Docker Compose
3. **SSH ключ** для подключения к серверу
4. **GitHub Container Registry (GHCR)** доступ

---

## ⚙️ Настройка GitHub Secrets

### Шаг 1: Перейдите в настройки репозитория

1. Откройте https://github.com/AlexDumper/STALKnet
2. **Settings** → **Secrets and variables** → **Actions**

### Шаг 2: Добавьте секреты

| Secret | Описание | Пример |
|--------|----------|--------|
| `SSH_PRIVATE_KEY` | Приватный SSH ключ для подключения к серверу | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `PRODUCTION_HOST` | IP-адрес или домен сервера | `87.242.103.13` |
| `SSH_USER` | Пользователь для подключения | `root` |
| `GITHUB_TOKEN` | Токен для GHCR (автоматически создаётся) | _(автоматически)_ |

### Шаг 3: Добавьте Environment (опционально)

1. **Settings** → **Environments** → **New environment**
2. Название: `production`
3. Добавьте **required reviewers** для approval деплоя

---

## 🖥️ Настройка сервера

### Шаг 1: Подготовка SSH ключа

**На локальной машине (Windows PowerShell):**

```powershell
# Генерация ключа (если нет)
ssh-keygen -t ed25519 -f $env:USERPROFILE\.ssh\id_ed25519

# Копирование публичного ключа на сервер
type $env:USERPROFILE\.ssh\id_ed25519.pub | ssh root@87.242.103.13 "cat >> ~/.ssh/authorized_keys"
```

**На Linux/Mac:**

```bash
# Генерация ключа
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519

# Копирование на сервер
ssh-copy-id -i ~/.ssh/id_ed25519 root@87.242.103.13
```

### Шаг 2: Установка Docker на сервере

```bash
# Обновление системы
apt update && apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Установка Docker Compose
apt install docker-compose-plugin -y

# Проверка
docker --version
docker compose version
```

### Шаг 3: Клонирование репозитория

```bash
cd /opt
git clone https://github.com/AlexDumper/STALKnet.git
cd STALKnet
```

### Шаг 4: Создание конфигурации

```bash
# Копирование шаблона
cp .env.production.example .env.production

# Редактирование (обязательно измените JWT_SECRET и DB_PASSWORD!)
nano .env.production
```

### Шаг 5: Сохранение GitHub токена (опционально)

Для pull образов из GHCR на сервере:

```bash
# Создать файл с токеном
echo "ghp_ваш_токен" > /root/.github_token
chmod 600 /root/.github_token
```

**Создание токена:**
1. GitHub → Settings → Developer settings → Personal access tokens
2. **Generate new token (classic)**
3. Scopes: `read:packages`
4. Скопировать токен

---

## 🚀 Автоматическое развёртывание

### Настройка workflow

Файл workflow: `.github/workflows/deploy.yml`

**Триггеры:**
- Пуш в ветку `main`
- Создание тега `v*`
- Pull request в `main`

### Процесс развёртывания

```
1. Push в main branch
       ↓
2. GitHub Actions запускается
       ↓
3. Build Docker образов
   - Gateway
   - Auth Service
   - Chat Service
   - и т.д.
       ↓
4. Push образов в GHCR
       ↓
5. SSH подключение к серверу
       ↓
6. Бэкап БД и конфигов
       ↓
7. Pull нового кода
       ↓
8. Pull Docker образов
       ↓
9. Перезапуск сервисов
       ↓
10. Health check (30 сек)
       ↓
11. ✅ Деплой завершён
```

### Мониторинг деплоя

1. GitHub → Actions → Deploy to Production
2. Просмотр логов в реальном времени
3. Статус: ✅ Success / ❌ Failed

---

## 🔧 Ручное обновление

### Вариант 1: PowerShell скрипт (локально)

**Файл:** `scripts/update-prod-remote.ps1`

```powershell
cd C:\Users\User\QWEN\STALKnet
.\scripts\update-prod-remote.ps1
```

**Что делает:**
1. Подключение к серверу по SSH
2. Создание бэкапа
3. Pull изменений из GitHub
4. Перезапуск сервисов
5. Проверка health endpoints

### Вариант 2: Bash скрипт (на сервере)

**Файл:** `scripts/update-prod.sh`

```bash
# На сервере
cd /opt/STALKnet

# Обновление до main ветки
./scripts/update-prod.sh main

# Или до конкретной ветки
./scripts/update-prod.sh develop
```

### Вариант 3: Вручную по шагам

```bash
# 1. Подключение к серверу
ssh -i $env:USERPROFILE\.ssh\id_ed25519 root@87.242.103.13

# 2. Переход в директорию
cd /opt/STALKnet

# 3. Бэкап БД
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
docker exec stalknet-postgres pg_dump -U stalknet stalknet > /opt/backups/db_backup_$TIMESTAMP.sql

# 4. Pull кода
git pull origin main

# 5. Login в GHCR (если нужно)
echo "ghp_токен" | docker login ghcr.io -u username --password-stdin

# 6. Pull образов
docker-compose -f docker-compose.prod.yml pull

# 7. Перезапуск
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build

# 8. Проверка
docker-compose -f docker-compose.prod.yml ps
curl http://localhost:8080/health
```

---

## ↩️ Откат изменений

### Вариант 1: Через GitHub Actions (Manual)

1. GitHub → Actions → Deploy to Production
2. **Run workflow** → Выбрать **rollback**
3. Запуск

### Вариант 2: Вручную на сервере

```bash
# На сервере
cd /opt/STALKnet

# Откат к предыдущему коммиту
git reset --hard HEAD~1

# Перезапуск
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build
```

### Вариант 3: Восстановление из бэкапа

```bash
# На сервере
cd /opt/backups

# Показать доступные бэкапы
ls -la

# Восстановление БД
TIMESTAMP=20260307_120000  # Ваш timestamp
docker exec -i stalknet-postgres psql -U stalknet stalknet < db_backup_$TIMESTAMP.sql

# Восстановление кода
cd /opt/STALKnet
tar -xzf /opt/backups/code_backup_$TIMESTAMP.tar.gz

# Перезапуск
docker-compose -f docker-compose.prod.yml up -d
```

---

## 📊 Мониторинг

### Проверка статуса

```bash
# Все сервисы
docker-compose -f docker-compose.prod.yml ps

# Логи в реальном времени
docker-compose -f docker-compose.prod.yml logs -f

# Логи конкретного сервиса
docker-compose logs -f auth
```

### Health endpoints

| Сервис | Endpoint |
|--------|----------|
| Gateway | http://87.242.103.13:8080/health |
| Auth | http://87.242.103.13:8081/health |
| Chat | http://87.242.103.13:8083/health |
| Task | http://87.242.103.13:8084/health |
| Notification | http://87.242.103.13:8085/health |

**Проверка:**

```bash
curl http://87.242.103.13:8080/health
# {"status":"ok"}
```

### GitHub Actions статус

1. GitHub → Actions → Deploy to Production
2. История всех деплоев
3. Логи каждого шага

---

## 🔐 Безопасность

### Рекомендации

1. **Храните секреты в GitHub Secrets**, не в коде
2. **Используйте SSH ключи** вместо паролей
3. **Ограничьте доступ** к production environment
4. **Включите required reviewers** для approval деплоя
5. **Регулярно обновляйте токены** (каждые 90 дней)

### Защита SSH ключа

```powershell
# Правильные права на ключ
icacls $env:USERPROFILE\.ssh\id_ed25519 /inheritance:r
icacls $env:USERPROFILE\.ssh\id_ed25519 /grant:r "$($env:USERNAME):(R)"

# На сервере
chmod 600 ~/.ssh/id_ed25519
chmod 700 ~/.ssh
```

---

## 📝 Troubleshooting

### Деплой не запускается

**Проверьте:**
1. Ветка `main` существует
2. Workflow файл в `.github/workflows/deploy.yml`
3. GitHub Secrets настроены
4. SSH ключ действителен

```bash
# Проверка SSH подключения
ssh -i $env:USERPROFILE\.ssh\id_ed25519 root@87.242.103.13
```

### Ошибка pull образов из GHCR

**Решение:**
```bash
# Login в GHCR на сервере
echo "ghp_токен" | docker login ghcr.io -u username --password-stdin

# Или создать файл с токеном
echo "ghp_токен" > /root/.github_token
chmod 600 /root/.github_token
```

### Health check не проходит

**Проверьте логи:**
```bash
docker-compose -f docker-compose.prod.yml logs auth
docker-compose -f docker-compose.prod.yml logs gateway
```

**Возможные причины:**
- Неправильный JWT_SECRET
- PostgreSQL не запустился
- Redis недоступен

### Откат после неудачного деплоя

```bash
# На сервере
cd /opt/STALKnet

# Показать историю коммитов
git log --oneline -5

# Откат к конкретному коммиту
git reset --hard abc1234

# Перезапуск
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## 📚 Дополнительные ресурсы

| Файл | Описание |
|------|----------|
| [DEPLOYMENT.md](DEPLOYMENT.md) | Общее руководство по развёртыванию |
| [DEPLOYMENT_CLOUD_RU.md](DEPLOYMENT_CLOUD_RU.md) | Развёртывание на Cloud.ru |
| [PRODUCTION_READY.md](PRODUCTION_READY.md) | Продакшн-подготовка |
| `.github/workflows/deploy.yml` | GitHub Actions workflow |
| `scripts/update-prod.sh` | Bash скрипт обновления |
| `scripts/update-prod-remote.ps1` | PowerShell скрипт обновления |

---

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи GitHub Actions
2. Проверьте логи сервисов на сервере
3. Проверьте GitHub Secrets
4. Убедитесь, что SSH ключ действителен

**GitHub Issues:** https://github.com/AlexDumper/STALKnet/issues

---

**Дата обновления:** 2026-03-07
**Версия STALKnet:** v0.1.17
**Статус:** ✅ Готово к развёртыванию через GitHub

---

## 📝 История обновлений

### v0.1.17 - Исправление поиска пользователей

**Проблема:** При отправке приватных сообщений Chat Service не мог найти пользователей.

**Решение:** Добавлена переменная `AUTH_SERVICE_URL=http://auth:8081` в `docker-compose.prod.yml` для сервиса `chat`.

**Изменения:**
- Обновлён `docker-compose.prod.yml`
- Добавлен скрипт `scripts/update-prod.sh`
- Создана документация по GitHub CI/CD
