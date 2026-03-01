-- Обновление статического контента STALKnet
-- Этот скрипт обновляет контент в таблице static_content

-- Справка для гостей (неавторизованные пользователи)
UPDATE static_content SET content = '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
───' WHERE content_key = 'help_guest';

-- Справка для авторизованных пользователей
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

-- Приветственное сообщение (загружается при подключении)
-- С информацией о логировании согласно ФЗ-374 от 06.07.2016
UPDATE static_content SET content = '───
Добро пожаловать в STALKnet!
• /help - Список команд
• /auth - Авторизация
— Автор полностью согласен с требованиями регулятора о предоставлении информации о действиях пользователей. Все сообщения логируются, могут быть просмотрены и прочитаны. Срок хранения - 1 год. Требование составлено на основе ФЗ-374 от 06.07.2016
───' WHERE content_key = 'help_welcome';

-- Новые записи (если не существуют)
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state, content_type, language, is_active)
SELECT 'help_welcome', 'Приветствие', 'Добро пожаловать в STALKnet!
Введите /help для списка команд
Введите /auth для авторизации', 0, 4, 'text', 'ru', TRUE
WHERE NOT EXISTS (SELECT 1 FROM static_content WHERE content_key = 'help_welcome');

INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state, content_type, language, is_active)
SELECT 'help_guest', 'Команды гостя', '───
• /help - Эта справка
• /clear - Очистить экран
• /connect - Статус подключения
• /quit - Выйти
• /auth - Авторизация
• /logout - Выйти из аккаунта
• /login <user> <pass> - Быстрый вход
───', 0, 0, 'text', 'ru', TRUE
WHERE NOT EXISTS (SELECT 1 FROM static_content WHERE content_key = 'help_guest');

INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state, content_type, language, is_active)
SELECT 'help_authorized', 'Команды пользователя', '───
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
───', 4, 4, 'text', 'ru', TRUE
WHERE NOT EXISTS (SELECT 1 FROM static_content WHERE content_key = 'help_authorized');
