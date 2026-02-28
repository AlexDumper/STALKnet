# Скрипт сборки STALKnet с автоматическим обновлением версии
# Использование: .\scripts\build.ps1

$ErrorActionPreference = "Stop"

Write-Host "=== STALKnet Build Script ===" -ForegroundColor Cyan
Write-Host ""

# Шаг 1: Обновление версии
Write-Host "[1/4] Обновление версии..." -ForegroundColor Yellow
& ".\scripts\bump_version.ps1"

# Шаг 2: Сборка Docker образа
Write-Host ""
Write-Host "[2/4] Сборка Docker образа gateway..." -ForegroundColor Yellow
docker-compose build gateway

if ($LASTEXITCODE -ne 0) {
    Write-Error "Ошибка сборки Docker образа"
    exit 1
}

# Шаг 3: Перезапуск контейнеров
Write-Host ""
Write-Host "[3/4] Перезапуск контейнеров..." -ForegroundColor Yellow
docker-compose up -d gateway

if ($LASTEXITCODE -ne 0) {
    Write-Error "Ошибка запуска контейнеров"
    exit 1
}

# Шаг 4: Вывод информации
Write-Host ""
Write-Host "[4/4] Информация о контейнерах:" -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "=== Сборка завершена ===" -ForegroundColor Green
Write-Host "Откройте http://localhost:8080 в браузере" -ForegroundColor Cyan
