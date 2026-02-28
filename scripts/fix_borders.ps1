# Скрипт замены старых рамок на новые в app.js

$file = "client\web\app.js"
$content = Get-Content $file -Raw -Encoding UTF8

# Заменяем длинные рамки
$content = $content -replace '╭────────────────────────────────────────────╮', '───'
$content = $content -replace '╰────────────────────────────────────────────╯', '───'
# Заменяем короткие рамки
$content = $content -replace '╭────────────────────────────────────╮', '───'
$content = $content -replace '╰────────────────────────────────────╯', '───'
# Удаляем вертикальные палочки с пробелом
$content = $content -replace '│ ', ''

Set-Content $file $content -Encoding UTF8 -NoNewline
Write-Host "Готово!" -ForegroundColor Green
