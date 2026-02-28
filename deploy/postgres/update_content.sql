-- Обновление статического контента
UPDATE static_content SET content = '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
───' WHERE content_key = 'help_guest';

UPDATE static_content SET content = '───
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
───' WHERE content_key = 'help_authorized';

UPDATE static_content SET content = '───
Добро пожаловать в STALKnet!
• /help - Список команд
• /auth - Авторизация
───' WHERE content_key = 'help_welcome';
