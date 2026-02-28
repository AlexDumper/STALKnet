-- Инициализация базы данных STALKnet

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'offline',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- Комнаты чата
CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by INTEGER REFERENCES users(id),
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rooms_name ON rooms(name);

-- Участники комнат
CREATE TABLE IF NOT EXISTS room_members (
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

-- Сообщения
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    room_id INTEGER REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_room_id ON messages(room_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- Задачи
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    assignee_id INTEGER REFERENCES users(id),
    room_id INTEGER REFERENCES rooms(id),
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    confirmed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_creator_id ON tasks(creator_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_room_id ON tasks(room_id);

-- Статусы задач: open, in_progress, done, confirmed

-- Статический контент (инструкции, справка)
CREATE TABLE IF NOT EXISTS static_content (
    id SERIAL PRIMARY KEY,
    content_key VARCHAR(100) NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    content_type VARCHAR(20) DEFAULT 'text',
    min_auth_state INT DEFAULT 0,
    max_auth_state INT DEFAULT 4,
    language VARCHAR(10) DEFAULT 'ru',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_static_content_key ON static_content(content_key);
CREATE INDEX idx_static_content_auth ON static_content(min_auth_state, max_auth_state);
CREATE INDEX idx_static_content_language ON static_content(language);

-- Начальные данные для справки
INSERT INTO static_content (content_key, title, content, min_auth_state, max_auth_state) VALUES
('help_guest', 'Базовые команды', 
'╭────────────────────────────────────────────╮
│ Доступные команды:
│ /help - Показать эту справку
│ /clear - Очистить экран
│ /connect - Статус подключения
│ /quit - Выйти из аккаунта и приложения
│ /auth - Авторизация
│ /logout - Выйти из аккаунта
│ /login <user> <pass> - Быстрый вход
╰────────────────────────────────────────────╯', 0, 0),

('help_authorized', 'Полный список команд',
'╭────────────────────────────────────────────╮
│ Доступные команды:
│ /help - Показать эту справку
│ /clear - Очистить экран
│ /connect - Статус подключения
│ /quit - Выйти из аккаунта и приложения
│ /auth - Авторизация
│ /logout - Выйти из аккаунта
│ /login <user> <pass> - Быстрый вход
│ /nick <name> - Сменить имя
│ /mock <text> - Отправить сообщение
│ /mockmsg - Случайное сообщение
│ /mocktask - Показать задание
╰────────────────────────────────────────────╯', 4, 4),

('help_welcome', 'Добро пожаловать',
'╭────────────────────────────────────────────╮
│ Добро пожаловать в STALKnet!
│ Введите /help для списка команд
│ Введите /auth для авторизации
╰────────────────────────────────────────────╯', 0, 0);
