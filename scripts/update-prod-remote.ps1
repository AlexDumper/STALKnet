# Скрипт обновления STALKnet на production-сервере через GitHub
# Использование: .\scripts\update-prod-remote.ps1

$ErrorActionPreference = "Stop"

# Настройки
$ServerHost = "87.242.103.13"
$ServerUser = "root"
$SSHKeyPath = "$env:USERPROFILE\.ssh\id_ed25519"
$Branch = "main"
$ProjectDir = "/opt/STALKnet"

Write-Host "=== STALKnet Remote Production Update ===" -ForegroundColor Cyan
Write-Host ""

# Проверка SSH ключа
Write-Host "[1/6] Проверка SSH ключа..." -ForegroundColor Yellow
if (!(Test-Path $SSHKeyPath)) {
    Write-Error "SSH ключ не найден: $SSHKeyPath"
    exit 1
}
Write-Host "  Ключ найден: $SSHKeyPath" -ForegroundColor Green

# Проверка подключения к серверу
Write-Host ""
Write-Host "[2/6] Проверка подключения к серверу..." -ForegroundColor Yellow
$testConnection = Test-NetConnection -ComputerName $ServerHost -Port 22 -WarningAction SilentlyContinue
if (!$testConnection.TcpTestSucceeded) {
    Write-Error "Не удалось подключиться к серверу $ServerHost:22"
    exit 1
}
Write-Host "  Сервер доступен: $ServerHost" -ForegroundColor Green

# Создание бэкапа
Write-Host ""
Write-Host "[3/6] Создание резервной копии на сервере..." -ForegroundColor Yellow

$backupScript = @'
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"
mkdir -p $BACKUP_DIR

echo "Бэкап PostgreSQL..."
docker exec stalknet-postgres pg_dump -U stalknet stalknet > $BACKUP_DIR/db_backup_$TIMESTAMP.sql

echo "Бэкап конфигурации..."
cd /opt/STALKnet
tar -czf $BACKUP_DIR/config_backup_$TIMESTAMP.tar.gz .env.production docker-compose.prod.yml

echo "Backup completed: $TIMESTAMP"
'@

ssh -i $SSHKeyPath "$ServerUser@$ServerHost" $backupScript
Write-Host "  Бэкап создан" -ForegroundColor Green

# Обновление кода
Write-Host ""
Write-Host "[4/6] Загрузка изменений из GitHub..." -ForegroundColor Yellow

$pullScript = @"
cd $ProjectDir
git fetch origin $Branch
git reset --hard origin/$Branch
echo "Code updated to branch: $Branch"
"@

ssh -i $SSHKeyPath "$ServerUser@$ServerHost" $pullScript
Write-Host "  Код обновлён" -ForegroundColor Green

# Перезапуск сервисов
Write-Host ""
Write-Host "[5/6] Перезапуск сервисов..." -ForegroundColor Yellow

$restartScript = @'
cd /opt/STALKnet

# Login to GHCR (если есть токен)
if [ -f /root/.github_token ]; then
    GHCR_TOKEN=$(cat /root/.github_token)
    echo $GHCR_TOKEN | docker login ghcr.io -u $(git config user.email | cut -d'@' -f1) --password-stdin
fi

# Pull и перезапуск
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build

echo "Services restarted"
'@

ssh -i $SSHKeyPath "$ServerUser@$ServerHost" $restartScript
Write-Host "  Сервисы перезапущены" -ForegroundColor Green

# Проверка health
Write-Host ""
Write-Host "[6/6] Проверка работоспособности..." -ForegroundColor Yellow

$healthScript = @'
echo "Waiting 30 seconds..."
sleep 30

echo "Checking health endpoints..."
curl -f http://localhost:8080/health || exit 1
curl -f http://localhost:8081/health || exit 1
curl -f http://localhost:8083/health || exit 1

echo "All health checks passed!"
'@

$healthResult = ssh -i $SSHKeyPath "$ServerUser@$ServerHost" $healthScript
Write-Host "  $healthResult" -ForegroundColor Green

Write-Host ""
Write-Host "=== Обновление завершено ===" -ForegroundColor Green
Write-Host ""
Write-Host "Приложение доступно: http://$ServerHost`:8080" -ForegroundColor Cyan
