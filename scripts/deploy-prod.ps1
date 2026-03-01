# Скрипт развертывания STALKnet в продакшн
# Использование: .\scripts\deploy-prod.ps1

$ErrorActionPreference = "Stop"
Write-Host "=== STALKnet Production Deployment ===" -ForegroundColor Cyan
Write-Host ""

# Проверка Docker
Write-Host "[1/6] Проверка Docker..." -ForegroundColor Yellow
$dockerVersion = docker --version
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker не установлен или не запущен"
    exit 1
}
Write-Host "  $dockerVersion" -ForegroundColor Gray

# Проверка наличия .env.production
Write-Host ""
Write-Host "[2/6] Проверка конфигурации..." -ForegroundColor Yellow
if (!(Test-Path ".env.production")) {
    Write-Error "Файл .env.production не найден!"
    Write-Host "  Скопируйте .env.production.example в .env.production"
    Write-Host "  и заполните его необходимыми значениями" -ForegroundColor Yellow
    exit 1
}
Write-Host "  .env.production найден" -ForegroundColor Green

# Загрузка .env.production
Write-Host ""
Write-Host "[3/6] Загрузка переменных окружения..." -ForegroundColor Yellow
Get-Content .env.production | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

# Проверка JWT_SECRET
$jwtSecret = [Environment]::GetEnvironmentVariable("JWT_SECRET", "Process")
if ([string]::IsNullOrEmpty($jwtSecret) -or $jwtSecret -eq "ИЗМЕНИТЕ_ЭТОТ_СЕКРЕТ_НА_СЛУЧАЙНУЮ_СТРОКУ_МИНИМУМ_32_СИМВОЛА") {
    Write-Error "JWT_SECRET не настроен или использует значение по умолчанию!"
    Write-Host "  Измените JWT_SECRET в файле .env.production" -ForegroundColor Yellow
    exit 1
}
Write-Host "  JWT_SECRET настроен" -ForegroundColor Green

# Остановка старых контейнеров
Write-Host ""
Write-Host "[4/6] Остановка старых контейнеров..." -ForegroundColor Yellow
docker-compose -f docker-compose.prod.yml down --remove-orphans 2>$null
Write-Host "  Готово" -ForegroundColor Gray

# Сборка образов
Write-Host ""
Write-Host "[5/6] Сборка Docker образов..." -ForegroundColor Yellow
docker-compose -f docker-compose.prod.yml build --no-cache

if ($LASTEXITCODE -ne 0) {
    Write-Error "Ошибка сборки Docker образов"
    exit 1
}
Write-Host "  Сборка завершена" -ForegroundColor Green

# Запуск сервисов
Write-Host ""
Write-Host "[6/6] Запуск сервисов..." -ForegroundColor Yellow
docker-compose -f docker-compose.prod.yml up -d

if ($LASTEXITCODE -ne 0) {
    Write-Error "Ошибка запуска сервисов"
    exit 1
}

# Ожидание запуска
Write-Host ""
Write-Host "Ожидание запуска сервисов (30 секунд)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# Проверка статуса
Write-Host ""
Write-Host "=== Статус сервисов ===" -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml ps

Write-Host ""
Write-Host "=== Логи (последние 20 строк) ===" -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml logs --tail=20

Write-Host ""
Write-Host "=== Развертывание завершено ===" -ForegroundColor Green
Write-Host "Откройте http://localhost:8080 в браузере" -ForegroundColor Cyan
Write-Host ""
Write-Host "Полезные команды:" -ForegroundColor Yellow
Write-Host "  docker-compose -f docker-compose.prod.yml logs -f     # Логи в реальном времени" -ForegroundColor Gray
Write-Host "  docker-compose -f docker-compose.prod.yml ps          # Статус сервисов" -ForegroundColor Gray
Write-Host "  docker-compose -f docker-compose.prod.yml down        # Остановка" -ForegroundColor Gray
Write-Host "  docker-compose -f docker-compose.prod.yml restart     # Перезапуск" -ForegroundColor Gray
