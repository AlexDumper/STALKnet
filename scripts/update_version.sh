#!/bin/bash
# Скрипт автоматического обновления версии STALKnet

# Текущая версия
CURRENT_VERSION=$(grep 'APP_VERSION = "' client/web/app.js | cut -d'"' -f2)

# Разбиваем версию на части
MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1)
MINOR=$(echo $CURRENT_VERSION | cut -d. -f2)
PATCH=$(echo $CURRENT_VERSION | cut -d. -f3)

# Получаем количество изменённых файлов с последнего коммита
CHANGED_FILES=$(git diff --name-only HEAD~1 2>/dev/null | wc -l)

# Если нет изменений с прошлого коммита, выходим
if [ "$CHANGED_FILES" -eq 0 ]; then
    echo "Нет изменений для обновления версии"
    exit 0
fi

# Определяем тип обновления на основе изменений
# Если изменён README или много файлов (>10) - минорное обновление
# Иначе - патч
if [ "$CHANGED_FILES" -gt 10 ] || git diff --name-only HEAD~1 | grep -q "README.md"; then
    # Минорное обновление
    MINOR=$((MINOR + 1))
    PATCH=0
    UPDATE_TYPE="minor"
else
    # Патч
    PATCH=$((PATCH + 1))
    UPDATE_TYPE="patch"
fi

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

echo "Обновление версии: $CURRENT_VERSION -> $NEW_VERSION (тип: $UPDATE_TYPE)"
echo "Изменено файлов: $CHANGED_FILES"

# Обновляем версию в файлах
sed -i "s/APP_VERSION = \"$CURRENT_VERSION\"/APP_VERSION = \"$NEW_VERSION\"/" client/web/app.js
sed -i "s/APP_VERSION = \"$CURRENT_VERSION\"/APP_VERSION = \"$NEW_VERSION\"/" gateway/web/app.js

# Обновляем версию в index.html (запасной вариант)
sed -i "s/v${CURRENT_VERSION}/v${NEW_VERSION}/g" client/web/index.html
sed -i "s/v${CURRENT_VERSION}/v${NEW_VERSION}/g" gateway/web/index.html

echo "Версия обновлена!"
