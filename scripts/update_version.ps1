# Скрипт автоматического обновления версии STALKnet для Windows

$ErrorActionPreference = "Stop"

# Пути к файлам
$AppJsClient = "client\web\app.js"
$AppJsGateway = "gateway\web\app.js"
$IndexHtmlClient = "client\web\index.html"
$IndexHtmlGateway = "gateway\web\index.html"

# Получаем текущую версию
$CurrentVersion = Select-String -Path $AppJsClient -Pattern 'APP_VERSION = "([\d.]+)"' | ForEach-Object { $_.Matches.Groups[1].Value }

if (-not $CurrentVersion) {
    Write-Error "Не удалось получить текущую версию"
    exit 1
}

# Разбиваем версию на части
$Parts = $CurrentVersion.Split('.')
$Major = [int]$Parts[0]
$Minor = [int]$Parts[1]
$Patch = [int]$Parts[2]

# Получаем количество изменённых файлов с последнего коммита
$ChangedFiles = git diff --name-only "HEAD~1" 2>$null | Measure-Object | Select-Object -ExpandProperty Count

# Если нет изменений с прошлого коммита, выходим
if ($ChangedFiles -eq 0) {
    Write-Host "Нет изменений для обновления версии" -ForegroundColor Yellow
    exit 0
}

# Определяем тип обновления на основе изменений
$ReadMeChanged = git diff --name-only "HEAD~1" 2>$null | Select-String "README.md"

if ($ChangedFiles -gt 10 -or $ReadMeChanged) {
    # Минорное обновление
    $Minor++
    $Patch = 0
    $UpdateType = "minor"
} else {
    # Патч
    $Patch++
    $UpdateType = "patch"
}

$NewVersion = "${Major}.${Minor}.${Patch}"

Write-Host "Обновление версии: $CurrentVersion -> $NewVersion (тип: $UpdateType)" -ForegroundColor Green
Write-Host "Изменено файлов: $ChangedFiles" -ForegroundColor Cyan

# Обновляем версию в файлах
$Files = @($AppJsClient, $AppJsGateway, $IndexHtmlClient, $IndexHtmlGateway)

foreach ($File in $Files) {
    if (Test-Path $File) {
        $Content = Get-Content $File -Raw
        $Content = $Content -replace "APP_VERSION = `"$CurrentVersion`"", "APP_VERSION = `"$NewVersion`""
        $Content = $Content -replace "v${CurrentVersion}", "v${NewVersion}"
        Set-Content $File $Content -NoNewline
        Write-Host "  Обновлён: $File" -ForegroundColor Gray
    }
}

Write-Host "`nВерсия обновлена!" -ForegroundColor Green
Write-Host "Не забудьте закоммитить изменения и пересобрать Docker:" -ForegroundColor Yellow
Write-Host "  git add -A" -ForegroundColor Cyan
Write-Host "  git commit -m `"chore: bump version to $NewVersion`"" -ForegroundColor Cyan
Write-Host "  docker-compose build gateway" -ForegroundColor Cyan
Write-Host "  docker-compose up -d" -ForegroundColor Cyan
