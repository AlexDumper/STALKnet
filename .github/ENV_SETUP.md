# GitHub Actions Environment Variables для STALKnet

Эти переменные необходимо настроить в **GitHub Secrets** для автоматического развёртывания.

---

## 🔐 Required Secrets

Настройте в GitHub: **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

### Обязательные секреты

| Secret Name | Описание | Как получить |
|-------------|----------|--------------|
| `SSH_PRIVATE_KEY` | Приватный SSH ключ для подключения к серверу | См. ниже |
| `PRODUCTION_HOST` | IP-адрес или домен сервера | Ваш сервер |
| `SSH_USER` | Пользователь для SSH подключения | Обычно `root` |

---

## 📝 Пошаговая инструкция

### Шаг 1: Создание SSH ключа (если нет)

**Windows PowerShell:**
```powershell
# Генерация ключа
ssh-keygen -t ed25519 -f $env:USERPROFILE\.ssh\github_actions

# Показать публичный ключ
type $env:USERPROFILE\.ssh\github_actions.pub
```

**Linux/Mac:**
```bash
# Генерация ключа
ssh-keygen -t ed25519 -f ~/.ssh/github_actions

# Показать публичный ключ
cat ~/.ssh/github_actions.pub
```

### Шаг 2: Копирование ключа на сервер

**Windows PowerShell:**
```powershell
# Копирование публичного ключа на сервер
type $env:USERPROFILE\.ssh\github_actions.pub | ssh root@87.242.103.13 "cat >> ~/.ssh/authorized_keys"
```

**Linux/Mac:**
```bash
# Копирование на сервер
ssh-copy-id -i ~/.ssh/github_actions root@87.242.103.13
```

### Шаг 3: Добавление секрета в GitHub

1. Откройте https://github.com/AlexDumper/STALKnet/settings/secrets/actions
2. **New repository secret**
3. Заполните:

**SSH_PRIVATE_KEY:**
```
Name: SSH_PRIVATE_KEY
Value: (содержимое файла $env:USERPROFILE\.ssh\github_actions)
```

**Пример содержимого ключа:**
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBHK2Y5r... (полный ключ)
-----END OPENSSH PRIVATE KEY-----
```

**PRODUCTION_HOST:**
```
Name: PRODUCTION_HOST
Value: 87.242.103.13
```

**SSH_USER:**
```
Name: SSH_USER
Value: root
```

---

## 🐳 GitHub Container Registry (GHCR)

### Автоматический доступ

GitHub Actions автоматически имеет доступ к GHCR через `GITHUB_TOKEN`.

### Ручной токен (для сервера)

Если на сервере нужно pullить образы из GHCR:

1. GitHub → Settings → Developer settings → Personal access tokens
2. **Generate new token (classic)**
3. Scopes: `read:packages`
4. Скопировать токен

**На сервере:**
```bash
# Создать файл с токеном
echo "ghp_ваш_токен" > /root/.github_token
chmod 600 /root/.github_token
```

---

## 🌍 Environments (опционально)

Для дополнительного контроля деплоя:

1. GitHub → Settings → Environments
2. **New environment**: `production`
3. **Required reviewers**: добавьте ревьюеров
4. **Deployment branches**: `main`

Теперь деплой в production потребует approval.

---

## 📋 Проверка настройки

### Тестовый деплой

1. Сделайте пуш в ветку `main`:
```bash
git add .
git commit -m "test: проверка CI/CD"
git push origin main
```

2. Проверьте GitHub Actions:
   - https://github.com/AlexDumper/STALKnet/actions
   - Должен запуститься workflow **Deploy to Production**

3. Проверьте статус:
   - ✅ Все шаги зелёные
   - Health checks пройдены

### Проверка подключения

**Тест SSH из GitHub Actions:**

Добавьте тестовый workflow `.github/workflows/test-ssh.yml`:

```yaml
name: Test SSH Connection

on: workflow_dispatch

jobs:
  test-ssh:
    runs-on: ubuntu-latest
    steps:
      - name: Test SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.SSH_PRIVATE_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -H ${{ secrets.PRODUCTION_HOST }} >> ~/.ssh/known_hosts
          ssh -i ~/.ssh/id_ed25519 ${{ secrets.SSH_USER }}@${{ secrets.PRODUCTION_HOST }} "echo 'SSH connection successful!'"
```

Запустите вручную: **Actions** → **Test SSH Connection** → **Run workflow**

---

## 🔧 Обновление секретов

### Смена SSH ключа

```bash
# 1. Создать новый ключ
ssh-keygen -t ed25519 -f $env:USERPROFILE\.ssh\github_actions_new

# 2. Скопировать на сервер
type $env:USERPROFILE\.ssh\github_actions_new.pub | ssh root@87.242.103.13 "cat >> ~/.ssh/authorized_keys"

# 3. Обновить секрет в GitHub
# 4. Удалить старый ключ из authorized_keys на сервере
```

### Смена пароля БД

1. Сгенерировать новый пароль:
```powershell
.\scripts\generate-secrets.ps1
```

2. Обновить `.env.production` на сервере:
```bash
ssh root@87.242.103.13
cd /opt/STALKnet
nano .env.production
```

3. Перезапустить сервисы:
```bash
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d
```

---

## 📊 Переменные окружения (Environment Variables)

Эти переменные **не являются секретами** и могут быть установлены в `.env.production`:

| Переменная | Описание | Пример |
|------------|----------|--------|
| `DB_USER` | Пользователь PostgreSQL | `stalknet` |
| `DB_PASSWORD` | Пароль PostgreSQL | (сгенерировать) |
| `DB_NAME` | Имя базы данных | `stalknet` |
| `JWT_SECRET` | Секрет JWT токенов | (сгенерировать) |
| `REDIS_PASSWORD` | Пароль Redis | (опционально) |
| `GATEWAY_PORT` | Порт Gateway | `8080` |
| `APP_ENV` | Режим приложения | `production` |
| `LOG_LEVEL` | Уровень логирования | `info` |

---

## 🚨 Troubleshooting

### Ошибка: "Permission denied (publickey)"

**Причины:**
- Неправильный SSH_PRIVATE_KEY
- Ключ не добавлен в `~/.ssh/authorized_keys` на сервере
- Неправильные права на ключе

**Решение:**
```bash
# Проверка прав на сервере
ssh root@87.242.103.13 "ls -la ~/.ssh/"

# Должно быть:
# id_ed25519.pub -rw-r--r--
# authorized_keys -rw-------
```

### Ошибка: "Host key verification failed"

**Решение:**
GitHub Actions автоматически добавляет ключ сервера в known_hosts.

### Ошибка: "Docker login failed"

**Причины:**
- Неправильный GITHUB_TOKEN
- Нет доступа к packages

**Решение:**
Проверьте токен: `read:packages` scope должен быть включён.

---

## 📚 Дополнительные ресурсы

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [DEPLOYMENT.md](DEPLOYMENT.md) - общее руководство по развёртыванию
- [GITHUB_DEPLOY.md](GITHUB_DEPLOY.md) - развёртывание через GitHub

---

**Дата обновления:** 2026-03-07
**Версия STALKnet:** v0.1.16
