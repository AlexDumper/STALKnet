#!/bin/bash
# Скрипт обновления STALKnet на production-сервере через GitHub
# Использование: ./scripts/update-prod.sh [branch]

set -e

BRANCH=${1:-main}
PROJECT_DIR="/opt/STALKnet"
BACKUP_DIR="/opt/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}=====================================${NC}"
echo -e "${CYAN}  STALKnet Production Update${NC}"
echo -e "${CYAN}=====================================${NC}"
echo ""

# Проверка прав root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Ошибка: Запустите скрипт от root${NC}"
    exit 1
fi

# Переход в директорию проекта
cd "$PROJECT_DIR" || {
    echo -e "${RED}Ошибка: Директория $PROJECT_DIR не найдена${NC}"
    exit 1
}

# ==================== BACKUP ====================
echo -e "${YELLOW}[1/7] Создание резервной копии...${NC}"

mkdir -p "$BACKUP_DIR"

# Бэкап БД
echo "  → Бэкап PostgreSQL..."
docker exec stalknet-postgres pg_dump -U stalknet stalknet > "$BACKUP_DIR/db_backup_$TIMESTAMP.sql"

# Бэкап конфигов
echo "  → Бэкап конфигурации..."
tar -czf "$BACKUP_DIR/config_backup_$TIMESTAMP.tar.gz" .env.production docker-compose.prod.yml 2>/dev/null || true

# Бэкап текущей версии кода
echo "  → Бэкап кода..."
tar -czf "$BACKUP_DIR/code_backup_$TIMESTAMP.tar.gz" . 2>/dev/null || true

echo -e "${GREEN}  ✓ Бэкап завершён: $BACKUP_DIR${NC}"
echo ""

# ==================== PULL ====================
echo -e "${YELLOW}[2/7] Загрузка последних изменений из GitHub...${NC}"

git fetch origin "$BRANCH"
git reset --hard "origin/$BRANCH"

echo -e "${GREEN}  ✓ Код обновлён до ветки: $BRANCH${NC}"
echo ""

# ==================== ENVIRONMENT ====================
echo -e "${YELLOW}[3/7] Проверка переменных окружения...${NC}"

if [ ! -f .env.production ]; then
    echo -e "${YELLOW}  ⚠ .env.production не найден, создаю из примера...${NC}"
    cp .env.production.example .env.production
    echo -e "${RED}  ! ВНИМАНИЕ: Заполните .env.production своими данными!${NC}"
    exit 1
fi

# Проверка JWT_SECRET
if grep -q "ИЗМЕНИТЕ_ЭТОТ_СЕКРЕТ" .env.production; then
    echo -e "${RED}  ✗ JWT_SECRET не настроен!${NC}"
    exit 1
fi

echo -e "${GREEN}  ✓ Переменные окружения проверены${NC}"
echo ""

# ==================== DOCKER LOGIN ====================
echo -e "${YELLOW}[4/7] Вход в GitHub Container Registry...${NC}"

# Читаем токен из файла или переменной окружения
if [ -f /root/.github_token ]; then
    GHCR_TOKEN=$(cat /root/.github_token)
    docker login ghcr.io -u "$(git config user.email | cut -d'@' -f1)" --password-stdin <<< "$GHCR_TOKEN"
    echo -e "${GREEN}  ✓ Вход в GHCR выполнен${NC}"
else
    echo -e "${YELLOW}  ⚠ Токен GHCR не найден, пропускаем вход${NC}"
fi
echo ""

# ==================== RESTART ====================
echo -e "${YELLOW}[5/7] Перезапуск сервисов...${NC}"

# Удаляем старые контейнеры для применения новых переменных окружения
echo "  → Удаление старых контейнеров..."
docker rm -f stalknet-chat 2>/dev/null || true

docker compose -f docker-compose.prod.yml up -d --build

echo -e "${GREEN}  ✓ Сервисы перезапущены${NC}"
echo ""

# ==================== HEALTH CHECK ====================
echo -e "${YELLOW}[6/7] Проверка работоспособности сервисов...${NC}"

echo "  Ожидание запуска (30 секунд)..."
sleep 30

# Проверка health endpoints
HEALTH_CHECKS=(
    "http://localhost:8080/health:Gateway"
    "http://localhost:8081/health:Auth"
    "http://localhost:8083/health:Chat"
)

FAILED=0
for check in "${HEALTH_CHECKS[@]}"; do
    URL="${check%%:*}"
    NAME="${check##*:}"
    
    if curl -sf --max-time 10 "$URL" > /dev/null; then
        echo -e "${GREEN}  ✓ $NAME: OK${NC}"
    else
        echo -e "${RED}  ✗ $NAME: FAILED${NC}"
        FAILED=1
    fi
done

if [ $FAILED -eq 1 ]; then
    echo ""
    echo -e "${RED}=====================================${NC}"
    echo -e "${RED}  Некоторые сервисы не прошли проверку!${NC}"
    echo -e "${RED}=====================================${NC}"
    echo ""
    echo -e "${YELLOW}Проверьте логи:${NC}"
    echo "  docker compose -f docker-compose.prod.yml logs"
    echo ""
    echo -e "${YELLOW}Для отката используйте:${NC}"
    echo "  git reset --hard HEAD~1"
    echo "  docker compose -f docker-compose.prod.yml up -d --build"
    exit 1
fi

echo ""
echo -e "${GREEN}  ✓ Все проверки пройдены${NC}"
echo ""

# ==================== CLEANUP ====================
echo -e "${CYAN}Очистка старых образов...${NC}"
docker image prune -f

# Удаление старых бэкапов (> 7 дней)
echo "Удаление старых бэкапов (> 7 дней)..."
find "$BACKUP_DIR" -name "*.sql" -mtime +7 -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo ""

# ==================== SUMMARY ====================
echo -e "${CYAN}=====================================${NC}"
echo -e "${GREEN}  ✓ Обновление завершено успешно!${NC}"
echo -e "${CYAN}=====================================${NC}"
echo ""

VERSION=$(grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+' client/web/app.js 2>/dev/null | head -1 || echo "unknown")
echo -e "Версия: ${CYAN}v$VERSION${NC}"
echo -e "Время: ${CYAN}$(date)${NC}"
echo ""
echo -e "${YELLOW}Полезные команды:${NC}"
echo "  docker compose -f docker-compose.prod.yml logs -f     # Логи в реальном времени"
echo "  docker compose -f docker-compose.prod.yml ps          # Статус сервисов"
echo "  docker compose -f docker-compose.prod.yml restart     # Перезапуск"
echo ""
echo -e "${GREEN}Приложение доступно: http://localhost:8080${NC}"
