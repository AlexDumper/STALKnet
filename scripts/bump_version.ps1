# Скрипт автоматического обновления PATCH-версии STALKnet
# Используется перед сборкой Docker

$ErrorActionPreference = "Stop"

# Пути к файлам
$AppJsClient = "client\web\app.js"
$AppJsGateway = "gateway\web\app.js"

# Получаем текущую версию
$Content = Get-Content $AppJsClient -Raw
$Match = [regex]::Match($Content, 'APP_VERSION = "([\d.]+)"')

if (-not $Match.Success) {
    Write-Error "Не удалось получить текущую версию"
    exit 1
}

$CurrentVersion = $Match.Groups[1].Value

# Разбиваем версию на части
$Parts = $CurrentVersion.Split('.')
$Major = [int]$Parts[0]
$Minor = [int]$Parts[1]
$Patch = [int]$Parts[2]

# Увеличиваем PATCH-версию
$Patch++
$NewVersion = "${Major}.${Minor}.${Patch}"

Write-Host "Обновление версии: $CurrentVersion -> $NewVersion" -ForegroundColor Green

# Обновляем версию в файлах
$Files = @($AppJsClient, $AppJsGateway)

foreach ($File in $Files) {
    if (Test-Path $File) {
        $Content = Get-Content $File -Raw
        $Content = $Content -replace "APP_VERSION = `"$CurrentVersion`"", "APP_VERSION = `"$NewVersion`""
        Set-Content $File $Content -NoNewline
        Write-Host "  Обновлён: $File" -ForegroundColor Gray
    }
}

Write-Host "`nВерсия обновлена: $NewVersion" -ForegroundColor Green
