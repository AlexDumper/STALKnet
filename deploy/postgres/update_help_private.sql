-- Обновление справки: добавление команды /private

UPDATE static_content 
SET content = '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
• /nick <name> - Сменить имя
• /mock <text> - Отправить сообщение
• /mockmsg - Случайное сообщение
• /mocktask - Показать задание
• /private <имя> <текст> - Личное сообщение (видно только получателю)
───',
    updated_at = CURRENT_TIMESTAMP
WHERE content_key = 'help_authorized';

-- Проверка обновления
SELECT content_key, content FROM static_content WHERE content_key = 'help_authorized';
