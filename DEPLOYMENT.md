# 📦 Продакшн-развертывание STALKnet

## 📋 Требования

| Компонент | Версия | Примечание |
|-----------|--------|------------|
| Docker | 24+ | Контейнеризация |
| Docker Compose | v2.20+ | Оркестрация |
| RAM | 4 GB+ | Рекомендуется 8 GB |
| CPU | 4 cores+ | Рекомендуется 8 cores |
| Disk | 10 GB+ | Для данных и логов |

---

## 🚀 Быстрый старт

### 1. Подготовка окружения

```bash
# Скопируйте шаблон конфигурации
cp .env.production.example .env.production
```

### 2. Настройка переменных окружения

Откройте `.env.production` и измените:

```bash
# ОБЯЗАТЕЛЬНО измените!
DB_PASSWORD=ваш_сложный_пароль_для_БД
JWT_SECRET=случайная_строка_минимум_32_символа

# Пример генерации JWT_SECRET:
# Linux/Mac: openssl rand -base64 32
# PowerShell: [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

### 3. Запуск в продакшн

```bash
# PowerShell
.\scripts\deploy-prod.ps1

# Или вручную:
docker-compose -f docker-compose.prod.yml up -d
```

---

## 🔐 Безопасность

### Обязательные действия перед деплоем:

1. **Измените JWT_SECRET** на случайную строку (минимум 32 символа)
2. **Измените пароль БД** на сложный
3. **Настройте Redis пароль** в `deploy/redis/redis.conf`
4. **Ограничьте доступ к портам** через фаервол
5. **Настройте HTTPS** через reverse proxy (nginx/traefik)

### Рекомендуемые дополнительные меры:

```bash
# 1. Включите пароль в Redis (deploy/redis/redis.conf)
requirepass ваш_сложный_пароль

# 2. Привяжите Redis к localhost (если не нужен внешний доступ)
bind 127.0.0.1

# 3. Закройте порты БД и Redis для внешнего доступа
# В docker-compose.prod.yml закомментируйте:
# ports:
#   - "5432:5432"  # PostgreSQL
#   - "6379:6379"  # Redis
```

---

## 📊 Мониторинг

### Проверка статуса

```bash
# Статус всех сервисов
docker-compose -f docker-compose.prod.yml ps

# Логи в реальном времени
docker-compose -f docker-compose.prod.yml logs -f

# Логи конкретного сервиса
docker-compose -f docker-compose.prod.yml logs -f auth
```

### Health check endpoints

| Сервис | Endpoint |
|--------|----------|
| Gateway | http://localhost:8080/health |
| Auth | http://localhost:8081/health |
| Chat | http://localhost:8083/health |
| Task | http://localhost:8084/health |
| Notification | http://localhost:8085/health |

---

## 🔧 Управление

### Перезапуск сервисов

```bash
# Все сервисы
docker-compose -f docker-compose.prod.yml restart

# Конкретный сервис
docker-compose -f docker-compose.prod.yml restart auth
```

### Остановка

```bash
# Остановка без удаления данных
docker-compose -f docker-compose.prod.yml down

# Полная остановка с удалением данных
docker-compose -f docker-compose.prod.yml down -v
```

### Обновление

```bash
# Пересборка и перезапуск
.\scripts\deploy-prod.ps1

# Или вручную:
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## 🗄️ Работа с базой данных

### Подключение к PostgreSQL

```bash
# Из контейнера
docker exec -it stalknet-postgres psql -U stalknet -d stalknet

# С локальной машины (если порт открыт)
psql -h localhost -U stalknet -d stalknet
```

### Бэкап базы данных

```bash
# Создать бэкап
docker exec stalknet-postgres pg_dump -U stalknet stalknet > backup_$(date +%Y%m%d).sql

# Восстановить из бэкапа
cat backup_20260301.sql | docker exec -i stalknet-postgres psql -U stalknet stalknet
```

### Бэкап Redis

```bash
# Копия RDB файла
docker cp stalknet-redis:/data/dump.rdb ./redis_backup_$(date +%Y%m%d).rdb
```

---

## 📈 Масштабирование

### Увеличение ресурсов

Отредактируйте `docker-compose.prod.yml`:

```yaml
services:
  auth:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
```

### Горизонтальное масштабирование

```bash
# Запуск нескольких реплик auth service
docker-compose -f docker-compose.prod.yml up -d --scale auth=3
```

---

## 🔍 Логирование

### Настройка уровня логирования

В `.env.production`:

```bash
LOG_LEVEL=info  # debug, info, warn, error
```

### Сбор логов

```bash
# Экспорт логов в файл
docker-compose -f docker-compose.prod.yml logs > logs_$(date +%Y%m%d).txt

# Логи за последний час
docker-compose -f docker-compose.prod.yml logs --since 1h
```

---

## 🛡️ HTTPS/SSL настройка

### Вариант 1: Nginx reverse proxy

```nginx
# /etc/nginx/sites-available/stalknet
server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Вариант 2: Traefik

См. `docker-compose.traefik.yml` (требуется дополнительная настройка)

---

## ⚠️ Troubleshooting

### Сервис не запускается

```bash
# Проверьте логи
docker-compose -f docker-compose.prod.yml logs <сервис>

# Проверьте переменные окружения
docker-compose -f docker-compose.prod.yml config
```

### Проблемы с подключением к БД

```bash
# Проверьте статус PostgreSQL
docker-compose -f docker-compose.prod.yml ps postgres

# Проверьте логи БД
docker logs stalknet-postgres
```

### Проблемы с Redis

```bash
# Проверка подключения
docker exec stalknet-redis redis-cli ping

# Должно вернуть: PONG
```

---

## 📝 Чек-лист перед деплоем

- [ ] Изменён `JWT_SECRET` на случайную строку
- [ ] Изменён пароль БД на сложный
- [ ] Настроен пароль Redis (опционально)
- [ ] Порты БД/Redis закрыты для внешнего доступа
- [ ] Настроен фаервол
- [ ] Настроен HTTPS (nginx/traefik)
- [ ] Настроено логирование
- [ ] Настроен мониторинг
- [ ] Создан бэкап конфигурации

---

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи сервисов
2. Проверьте переменные окружения
3. Убедитесь, что все health checks проходят
4. Проверьте доступность портов
