# Скрипт генерации безопасных секретов для STALKnet
# Использование: .\scripts\generate-secrets.ps1

Write-Host "=== STALKnet Secrets Generator ===" -ForegroundColor Cyan
Write-Host ""

# Генерация JWT_SECRET (32 байта в base64)
Write-Host "[1/3] Генерация JWT_SECRET..." -ForegroundColor Yellow
$jwtSecret = [Convert]::ToBase64String((1..64 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
Write-Host "  JWT_SECRET=$jwtSecret" -ForegroundColor Green

# Генерация пароля БД (32 символа)
Write-Host ""
Write-Host "[2/3] Генерация DB_PASSWORD..." -ForegroundColor Yellow
$dbPassword = [Convert]::ToBase64String((1..24 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 })) -replace '[+/=]', ''
Write-Host "  DB_PASSWORD=$dbPassword" -ForegroundColor Green

# Генерация Redis пароля (32 символа)
Write-Host ""
Write-Host "[3/3] Генерация REDIS_PASSWORD..." -ForegroundColor Yellow
$redisPassword = [Convert]::ToBase64String((1..24 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 })) -replace '[+/=]', ''
Write-Host "  REDIS_PASSWORD=$redisPassword" -ForegroundColor Green

# Вывод для копирования
Write-Host ""
Write-Host "=== Скопируйте в .env.production ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "JWT_SECRET=$jwtSecret" -ForegroundColor Gray
Write-Host "DB_PASSWORD=$dbPassword" -ForegroundColor Gray
Write-Host "REDIS_PASSWORD=$redisPassword" -ForegroundColor Gray
Write-Host ""

# Предложение сохранить в файл
$save = Read-Host "Сохранить в .env.production? (y/n)"
if ($save -eq 'y' -or $save -eq 'Y') {
    if (Test-Path ".env.production") {
        $backup = Read-Host "Файл существует. Создать резервную копию? (y/n)"
        if ($backup -eq 'y' -or $backup -eq 'Y') {
            $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
            Copy-Item ".env.production" ".env.production.backup.$timestamp"
            Write-Host "  Резервная копия создана: .env.production.backup.$timestamp" -ForegroundColor Green
        }
    }

    # Создаём новый файл или обновляем существующий
    $content = @"
# Production переменные окружения для STALKnet
# Сгенерировано: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

# База данных
DB_USER=stalknet
DB_PASSWORD=$dbPassword
DB_NAME=stalknet

# JWT Secret
JWT_SECRET=$jwtSecret

# Redis (опционально, если включён пароль)
REDIS_PASSWORD=$redisPassword

# Порты
GATEWAY_PORT=8080
AUTH_PORT=8081
USER_PORT=8082
CHAT_PORT=8083
TASK_PORT=8084
NOTIFICATION_PORT=8085

# Домен (измените на ваш)
APP_DOMAIN=localhost

# Режим
APP_ENV=production
LOG_LEVEL=info
"@

    Set-Content -Path ".env.production" -Value $content -Encoding UTF8
    Write-Host "  Файл .env.production обновлён" -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Готово ===" -ForegroundColor Green
