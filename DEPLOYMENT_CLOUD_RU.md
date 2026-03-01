# 🚀 Развёртывание STALKnet на Cloud.ru

## 📋 Описание

Пошаговое руководство по развёртыванию STALKnet на российской облачной платформе **Cloud.ru** (бывший SberCloud).

---

## 🎯 Почему Cloud.ru

| Преимущество | Описание |
|--------------|----------|
| 🇷🇺 Российский провайдер | Соответствие ФЗ-152 (персональные данные) |
| 💰 Бесплатный тариф | 2 vCPU, 4 ГБ RAM, 30 ГБ NVMe — 0 ₽ |
| ⚡ Быстрый старт | Развёртывание за 15 минут |
| 🔒 Надёжность | SLA 99.99% |
| 🐳 Docker support | Официальная поддержка контейнеров |

---

## 📊 Тарифы VPS

| Тариф | vCPU | RAM | Диск | Цена | Назначение |
|-------|------|-----|------|------|------------|
| **Бесплатно** | 2 vCPU | 4 ГБ | 30 ГБ NVMe | **0 ₽** | Тестирование, staging |
| **Стандарт** | 1 vCPU | 2 ГБ | 30 ГБ NVMe | ~300 ₽/мес | Минимум production |
| **Оптимальный** | 2 vCPU | 4 ГБ | 50 ГБ NVMe | ~600 ₽/мес | Production |
| **Расширенный** | 4 vCPU | 8 ГБ | 80 ГБ NVMe | ~1200 ₽/мес | Production с запасом |

### Требования STALKnet

| Компонент | Мин. RAM | Рекомендуется |
|-----------|----------|---------------|
| PostgreSQL | 512 МБ - 1 ГБ | 2 ГБ |
| Redis | 100-500 МБ | 500 МБ |
| Микросервисы (7 шт.) | 1 ГБ | 2 ГБ |
| Gateway | 200 МБ | 500 МБ |
| **Итого** | **~2.5 ГБ** | **~5 ГБ** |

**Вывод:** Бесплатный тариф (4 ГБ) подходит для тестирования. Для production — от 4 ГБ RAM.

---

## 📝 Пошаговый план

### Этап 1: Регистрация и создание сервера

#### 1.1 Регистрация на Cloud.ru

1. Перейти на [cloud.ru](https://cloud.ru)
2. Нажать **"Зарегистрироваться"**
3. Ввести email, телефон
4. Подтвердить регистрацию

#### 1.2 Создание VPS

1. В панели: **Создать ресурс** → **Виртуальная машина**
2. Выбрать конфигурацию:
   - **ОС:** Ubuntu 22.04 LTS
   - **vCPU:** 2
   - **RAM:** 4 ГБ
   - **Диск:** 30 ГБ NVMe (или больше)
   - **Белый IP:** включить (~147 ₽/мес)
3. Нажать **"Создать"**

#### 1.3 Получение данных для подключения

После создания сервера:
- **IP-адрес:** записать (например, `185.123.45.67`)
- **Логин:** `root`
- **Пароль:** получить в панели (или использовать SSH-ключ)

---

### Этап 2: Настройка сервера

#### 2.1 Подключение по SSH

**Windows (PowerShell):**
```powershell
ssh root@185.123.45.67
```

**Linux/Mac:**
```bash
ssh root@185.123.45.67
```

**С SSH-ключом:**
```bash
# Генерация ключа
ssh-keygen -t ed25519

# Копирование ключа на сервер
ssh-copy-id root@185.123.45.67

# Подключение
ssh -i ~/.ssh/id_ed25519 root@185.123.45.67
```

---

#### 2.2 Обновление системы

```bash
apt update && apt upgrade -y
```

---

#### 2.3 Установка Docker

```bash
# Автоматическая установка
curl -fsSL https://get.docker.com | sh

# Запуск Docker
systemctl enable docker
systemctl start docker

# Проверка
docker --version
```

---

#### 2.4 Установка Docker Compose

```bash
apt install docker-compose-plugin -y

# Проверка
docker compose version
```

---

#### 2.5 Настройка фаервола (UFW)

```bash
# Установка
apt install ufw -y

# Разрешение SSH
ufw allow 22/tcp

# Разрешение HTTP/HTTPS
ufw allow 80/tcp
ufw allow 443/tcp

# Включение
ufw enable
ufw status
```

---

### Этап 3: Развёртывание STALKnet

#### 3.1 Клонирование репозитория

```bash
cd /opt
git clone https://github.com/AlexDumper/STALKnet.git
cd STALKnet
```

---

#### 3.2 Создание конфигурации

```bash
# Копирование шаблона
cp .env.production.example .env.production

# Редактирование
nano .env.production
```

**Изменить переменные:**
```bash
# Обязательно изменить!
JWT_SECRET=<случайная строка 64+ символа>
DB_PASSWORD=<сложный пароль 20+ символов>

# Пример генерации:
# openssl rand -base64 48
```

---

#### 3.3 Запуск сервисов

```bash
# Запуск production-конфигурации
docker compose -f docker-compose.prod.yml up -d

# Проверка статуса
docker compose ps

# Просмотр логов
docker compose logs -f
```

---

#### 3.4 Инициализация базы данных

```bash
# Выполнить один раз
docker exec stalknet-postgres psql -U stalknet -d stalknet -f /docker-entrypoint-initdb.d/init.sql
```

---

### Этап 4: Проверка работоспособности

#### 4.1 Health check

```bash
# Gateway
curl http://localhost:8080/health

# Auth Service
curl http://localhost:8081/health

# Chat Service
curl http://localhost:8083/health
```

**Ожидаемый ответ:**
```json
{"status": "ok"}
```

---

#### 4.2 Проверка с внешнего устройства

```bash
# С локального компьютера
curl http://<IP-адрес>:8080/health
```

**Открыть браузер:**
```
http://<IP-адрес>:8080
```

---

### Этап 5: Настройка домена (опционально)

#### 5.1 Привязка домена

**В DNS регистратора домена:**
```
A запись: stalknet.yourdomain.com → <IP-адрес>
```

**Пример для Reg.ru:**
1. Личный кабинет → Домены → ваш домен → DNS
2. Добавить запись:
   - Тип: `A`
   - Subdomain: `stalknet`
   - IP: `<IP-адрес>`

---

#### 5.2 Настройка HTTPS через Cloudflare Tunnel

**Преимущества:**
- Бесплатно
- Автоматический HTTPS
- Защита от DDoS
- Не нужно открывать порты

---

**Шаг 1: Регистрация на Cloudflare**

1. Перейти на [cloudflare.com](https://cloudflare.com)
2. Зарегистрироваться
3. Добавить домен (если ещё не добавлен)

---

**Шаг 2: Создание туннеля**

```bash
# На сервере
mkdir -p /opt/cloudflared
cd /opt/cloudflared

# Логин (откроется браузер)
docker run --rm -it -v $(pwd):/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel login

# Создание туннеля
docker run --rm -v $(pwd):/etc/cloudflared \
  cloudflare/cloudflared:latest tunnel create stalknet
```

**Сохранить ID туннеля** (выведется в консоли).

---

**Шаг 3: Настройка конфигурации**

```bash
# Создание файла конфигурации
nano config.yml
```

**Содержимое:**
```yaml
tunnel: stalknet
credentials-file: /etc/cloudflared/<ID-туннеля>.json

ingress:
  - hostname: stalknet.yourdomain.com
    service: http://stalknet-gateway:8080
    path: /
  
  - service: http_status:404
```

---

**Шаг 4: Запуск Cloudflare Tunnel**

```bash
# Запуск в фоновом режиме
docker run -d \
  --name cloudflared \
  --restart unless-stopped \
  -v $(pwd):/etc/cloudflared \
  cloudflare/cloudflared:latest \
  tunnel --config /etc/cloudflared/config.yml run
```

---

**Шаг 5: Привязка домена в Cloudflare**

1. Cloudflare Dashboard → Zero Trust → Access → Tunnels
2. Выбрать туннель `stalknet`
3. Add Public Hostname:
   - Subdomain: `stalknet`
   - Domain: `yourdomain.com`
   - Service: `HTTP: stalknet-gateway:8080`
4. Save

**Готово!** Доступ по HTTPS:
```
https://stalknet.yourdomain.com
```

---

### Этап 6: Безопасность

#### 6.1 Закрытие портов БД и Redis

**В `docker-compose.prod.yml`:**

```yaml
services:
  postgres:
    # ports:
    #   - "5432:5432"  # Закомментировать!
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    # ports:
    #   - "6379:6379"  # Закомментировать!
    volumes:
      - redis_data:/data
```

**Перезапуск:**
```bash
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

---

#### 6.2 Настройка Fail2ban

```bash
# Установка
apt install fail2ban -y

# Запуск
systemctl enable fail2ban
systemctl start fail2ban

# Проверка
systemctl status fail2ban
```

**Защита SSH от брутфорса включена по умолчанию.**

---

#### 6.3 Резервное копирование

**Скрипт бэкапа БД:**

```bash
# Создание файла
nano /opt/backup.sh
```

**Содержимое:**
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"
mkdir -p $BACKUP_DIR

# Бэкап PostgreSQL
docker exec stalknet-postgres pg_dump -U stalknet stalknet > \
  $BACKUP_DIR/db_$DATE.sql

# Бэкап Redis
docker cp stalknet-redis:/data/dump.rdb \
  $BACKUP_DIR/redis_$DATE.rdb

# Удаление старых бэкапов (> 7 дней)
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.rdb" -mtime +7 -delete

echo "Backup completed: $DATE"
```

**Запуск по расписанию (cron):**
```bash
chmod +x /opt/backup.sh

# Редактирование crontab
crontab -e

# Добавить строку (ежедневно в 3:00)
0 3 * * * /opt/backup.sh >> /var/log/backup.log 2>&1
```

---

## 💰 Стоимость владения

### Тестирование (бесплатно)

| Ресурс | Цена |
|--------|------|
| VPS (2 vCPU, 4 ГБ, 30 ГБ) | **0 ₽** |
| Белый IP | 0 ₽ (временно) |
| **Итого** | **0 ₽/мес** |

---

### Production

| Ресурс | Цена |
|--------|------|
| VPS (2 vCPU, 4 ГБ, 50 ГБ) | ~600 ₽ |
| Белый IP | ~147 ₽ |
| Бэкапы (S3, 50 ГБ) | ~50 ₽ |
| **Итого** | **~800 ₽/мес** |

---

## 🔧 Управление сервисами

### Проверка статуса

```bash
# Все сервисы
docker compose -f docker-compose.prod.yml ps

# Логи в реальном времени
docker compose -f docker-compose.prod.yml logs -f

# Логи конкретного сервиса
docker compose logs -f auth
```

---

### Перезапуск

```bash
# Все сервисы
docker compose -f docker-compose.prod.yml restart

# Конкретный сервис
docker compose restart auth
```

---

### Остановка

```bash
# Без удаления данных
docker compose -f docker-compose.prod.yml down

# С удалением данных
docker compose -f docker-compose.prod.yml down -v
```

---

### Обновление

```bash
# Обновление кода
git pull

# Пересборка и перезапуск
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d --build
```

---

## 📊 Мониторинг

### Проверка ресурсов

```bash
# Использование CPU/RAM
docker stats

# Свободное место
df -h

# Оперативная память
free -h
```

---

### Health check endpoints

| Сервис | Endpoint |
|--------|----------|
| Gateway | http://<IP>:8080/health |
| Auth | http://<IP>:8081/health |
| Chat | http://<IP>:8083/health |
| Task | http://<IP>:8084/health |
| Notification | http://<IP>:8085/health |
| Compliance | http://<IP>:8086/health |

---

### Логи

```bash
# Последние 100 строк
docker compose logs --tail=100

# Экспорт в файл
docker compose logs > logs.txt
```

---

## ⚠️ Troubleshooting

### Сервис не запускается

```bash
# Проверка логов
docker compose logs <сервис>

# Проверка переменных окружения
docker compose config
```

---

### Проблемы с подключением к БД

```bash
# Проверка статуса PostgreSQL
docker compose ps postgres

# Логи БД
docker logs stalknet-postgres

# Подключение к БД
docker exec -it stalknet-postgres psql -U stalknet -d stalknet
```

---

### Проблемы с Redis

```bash
# Проверка подключения
docker exec stalknet-redis redis-cli ping

# Должно вернуть: PONG
```

---

### Невозможно подключиться к серверу

```bash
# Проверка фаервола на сервере
ufw status

# Проверка группы безопасности в панели Cloud.ru
# (должны быть разрешены порты 22, 80, 443)
```

---

## 📋 Чек-лист перед запуском

- [ ] Зарегистрирован аккаунт на Cloud.ru
- [ ] Создан VPS (Ubuntu 22.04, 2+ vCPU, 4+ ГБ RAM)
- [ ] Получен белый IP
- [ ] Настроен SSH (ключи вместо пароля)
- [ ] Установлен Docker и Docker Compose
- [ ] Настроен фаервол (UFW)
- [ ] Склонирован репозиторий STALKnet
- [ ] Изменены `JWT_SECRET` и `DB_PASSWORD`
- [ ] Закрыты порты БД и Redis
- [ ] Настроен Fail2ban
- [ ] Настроено резервное копирование
- [ ] Проверены health check endpoints
- [ ] (Опционально) Настроен Cloudflare Tunnel для HTTPS

---

## 📞 Поддержка Cloud.ru

- **Документация:** [cloud.ru/docs](https://cloud.ru/docs)
- **Техподдержка:** support@cloud.ru
- **Телефон:** 8 800 555-XX-XX

---

## 📚 Дополнительные ресурсы

- [Официальная документация Docker](https://docs.docker.com)
- [Документация STALKnet](README.md)
- [DEPLOYMENT.md](DEPLOYMENT.md) — общее руководство по развёртыванию
- [PRODUCTION_READY.md](PRODUCTION_READY.md) — продакшн-подготовка

---

**Дата обновления:** 2026-03-01
**Версия STALKnet:** v0.1.11
**Статус:** ✅ Готово к развёртыванию
